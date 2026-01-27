package systems

import (
	"math"
	"os"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems/factory"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// slopeSurfaceOffset is a small offset to keep the player slightly above the slope surface
// to prevent z-fighting and ensure stable ground detection
const slopeSurfaceOffset = 0.1

func UpdateCollisions(ecs *ecs.ECS) {
	tags.Player.Each(ecs.World, func(e *donburi.Entry) {
		player := components.Player.Get(e)
		physics := components.Physics.Get(e)
		obj := components.Object.Get(e)

		resolveObjectHorizontalCollision(physics, obj.Object, true)
		resolveObjectVerticalCollision(physics, obj.Object)
		updateWallSliding(player, physics, obj.Object)

		// Check for dead zone collision
		if checkDeadZone(obj.Object) {
			handleDeadZoneHit(ecs, e)
		}
	})

	tags.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		physics := components.Physics.Get(e)
		obj := components.Object.Get(e)

		resolveObjectHorizontalCollision(physics, obj.Object, false)
		resolveObjectVerticalCollision(physics, obj.Object)

		// Kill enemy if they hit a dead zone
		if checkDeadZone(obj.Object) {
			health := components.Health.Get(e)
			health.Current = 0
		}
	})
}

// resolveObjectHorizontalCollision handles horizontal movement and wall collision for any object
func resolveObjectHorizontalCollision(physics *components.PhysicsData, object *resolv.Object, allowWallSlide bool) {
	dx := physics.SpeedX
	if dx == 0 {
		return
	}

	// Check for ramp collision in front (walking uphill)
	if rampCheck := object.Check(dx, 0, tags.ResolvRamp); rampCheck != nil {
		if ramps := rampCheck.ObjectsByTags(tags.ResolvRamp); len(ramps) > 0 {
			object.X += dx
			snapToSlopeSurface(physics, object, ramps[0])
			return
		}
	}

	// Check for ramp below (walking downhill or staying on slope)
	// Use a small downward check to detect ramps we're standing on or walking down
	if rampCheck := object.Check(dx, 1, tags.ResolvRamp); rampCheck != nil {
		if ramps := rampCheck.ObjectsByTags(tags.ResolvRamp); len(ramps) > 0 {
			object.X += dx
			snapToSlopeSurface(physics, object, ramps[0])
			return
		}
	}

	check := object.Check(dx, 0, "solid", "character")
	if check == nil {
		object.X += dx
		return
	}

	// Debug collision detection if enabled
	debugHorizontalCollision(dx, object, check)

	// Check for collisions with solid objects (walls)
	if shouldStopHorizontalMovement(object, check) {
		physics.SpeedX = 0
		if allowWallSlide {
			setWallSlidingIfAirborne(physics, check)
		}
		dx = 0 // Stop movement
	}

	// Check for collisions with other characters
	if characters := check.ObjectsByTags("character"); len(characters) > 0 {
		// Gentle push-back instead of a hard stop
		contact := check.ContactWithObject(characters[0])
		if contact.X() != 0 { // If there is penetration
			// Apply a small, fixed pushback
			if dx > 0 {
				dx = -1
			} else {
				dx = 1
			}
		} else {
			// If just touching, use the contact point to slide along the other character
			dx = contact.X()
		}
	}

	object.X += dx
}

// resolveObjectVerticalCollision handles vertical movement and ground/platform collision for any object
func resolveObjectVerticalCollision(physics *components.PhysicsData, object *resolv.Object) {
	physics.OnGround = nil
	dy := clampVerticalSpeed(physics.SpeedY)

	checkDistance := dy
	if dy >= 0 {
		checkDistance++
	}

	check := object.Check(0, checkDistance, "solid", "platform", "ramp")
	if check == nil {
		object.Y += dy
		return
	}

	if dy < 0 {
		dy = handleUpwardCollision(physics, object, check)
	} else {
		dy = handleDownwardCollision(physics, object, check, dy)
	}

	object.Y += dy
}

// updateWallSliding checks if player should disengage from wall sliding
func updateWallSliding(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object) {
	if physics.WallSliding == nil {
		return
	}

	wallDirection := player.Direction.X

	if check := playerObject.Check(wallDirection, 0, "solid"); check == nil {
		physics.WallSliding = nil
	}
}

// Helper functions for collision resolution

func debugHorizontalCollision(dx float64, object *resolv.Object, check *resolv.Collision) {
	if os.Getenv("DEBUG_COLLISION") == "" {
		return
	}
	// Debug print code removed for cleanliness
}

func shouldStopHorizontalMovement(object *resolv.Object, check *resolv.Collision) bool {
	solids := check.ObjectsByTags("solid")
	if len(solids) == 0 {
		return false
	}

	objectBottom := object.Y + object.H

	for _, solid := range solids {
		if objectBottom > solid.Y && object.Y < solid.Y+solid.H {
			return true
		}
	}

	return false
}

func setWallSlidingIfAirborne(physics *components.PhysicsData, check *resolv.Collision) {
	if physics.OnGround != nil {
		return
	}

	if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
		physics.WallSliding = solids[0]
	}
}

func clampVerticalSpeed(speedY float64) float64 {
	return math.Max(math.Min(speedY, 16), -16)
}

func handleUpwardCollision(physics *components.PhysicsData, object *resolv.Object, check *resolv.Collision) float64 {
	if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
		physics.SpeedY = 0
		return check.ContactWithObject(solids[0]).Y()
	}

	if len(check.Cells) > 0 && check.Cells[0].ContainsTags("solid") {
		if slide := check.SlideAgainstCell(check.Cells[0], "solid"); slide != nil {
			object.X += slide.X()
		}
	}

	return physics.SpeedY
}

func handleDownwardCollision(physics *components.PhysicsData, object *resolv.Object, check *resolv.Collision, dy float64) float64 {
	// Try collision in priority order: ramps, platforms, solids
	if newDy, handled := tryRampCollision(physics, object, check, dy); handled {
		return newDy
	}

	if newDy, handled := tryPlatformCollision(physics, object, check); handled {
		return newDy
	}

	if newDy, handled := trySolidCollision(physics, check); handled {
		return newDy
	}

	return dy
}

func tryRampCollision(physics *components.PhysicsData, object *resolv.Object, check *resolv.Collision, dy float64) (float64, bool) {
	ramps := check.ObjectsByTags(tags.ResolvRamp)
	if len(ramps) == 0 {
		return dy, false
	}

	ramp := ramps[0]

	// Only handle when falling or stationary (dy >= 0)
	if dy < 0 {
		return dy, false
	}

	surfaceY := getSlopeSurfaceY(object, ramp)
	playerBottom := object.Y + object.H

	if playerBottom+dy >= surfaceY {
		physics.OnGround = ramp
		physics.SpeedY = 0
		return surfaceY - playerBottom + slopeSurfaceOffset, true
	}

	return dy, false
}

// snapToSlopeSurface adjusts object Y position to stay on slope surface.
// Always snaps to surface when called - the Check() that triggers this
// already ensures we're close enough to the ramp to be walking on it.
func snapToSlopeSurface(physics *components.PhysicsData, object *resolv.Object, ramp *resolv.Object) {
	surfaceY := getSlopeSurfaceY(object, ramp)
	object.Y = surfaceY - object.H + slopeSurfaceOffset
	physics.OnGround = ramp
	physics.SpeedY = 0
}

// getSlopeSurfaceY calculates the slope surface Y at the object's center X position
func getSlopeSurfaceY(object *resolv.Object, ramp *resolv.Object) float64 {
	playerCenterX := object.X + object.W/2
	relativeX := clampFloat(playerCenterX-ramp.X, 0, ramp.W)
	slope := relativeX / ramp.W

	switch {
	case ramp.HasTags(tags.Slope45UpRight):
		// Surface rises from left (Y+H) to right (Y)
		return ramp.Y + ramp.H*(1-slope)
	case ramp.HasTags(tags.Slope45UpLeft):
		// Surface falls from left (Y) to right (Y+H)
		return ramp.Y + ramp.H*slope
	default:
		return ramp.Y
	}
}

// clampFloat constrains a value to the range [min, max]
func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func tryPlatformCollision(physics *components.PhysicsData, object *resolv.Object, check *resolv.Collision) (float64, bool) {
	if physics.OnGround != nil {
		return 0, false // Already grounded from ramp
	}

	platforms := check.ObjectsByTags("platform")
	if len(platforms) == 0 {
		return 0, false
	}

	platform := platforms[0]

	// Check platform collision conditions
	if platform == physics.IgnorePlatform ||
		physics.SpeedY < 0 ||
		object.Bottom() >= platform.Y+4 {
		return 0, false
	}

	physics.OnGround = platform
	physics.SpeedY = 0
	return check.ContactWithObject(platform).Y(), true
}

func trySolidCollision(physics *components.PhysicsData, check *resolv.Collision) (float64, bool) {
	if physics.OnGround != nil {
		clearGroundedState(physics)
		return 0, false // Already grounded
	}

	solids := check.ObjectsByTags("solid")
	if len(solids) == 0 {
		return 0, false
	}

	solid := solids[0]

	// Only land on solid if falling down
	if physics.SpeedY >= 0 {
		physics.OnGround = solid
		physics.SpeedY = 0
		clearGroundedState(physics)
		return check.ContactWithObject(solid).Y(), true
	}

	return 0, false
}

func clearGroundedState(physics *components.PhysicsData) {
	if physics.OnGround != nil {
		physics.WallSliding = nil
		physics.IgnorePlatform = nil
	}
}

// checkDeadZone returns true if the object is colliding with a dead zone
func checkDeadZone(obj *resolv.Object) bool {
	check := obj.Check(0, 0, tags.ResolvDeadZone)
	return check != nil
}

// handleDeadZoneHit triggers death sequence with visual effects and delay
func handleDeadZoneHit(ecs *ecs.ECS, e *donburi.Entry) {
	// Early return if already in death sequence
	if e.HasComponent(components.Death) {
		return
	}

	lives := components.Lives.Get(e)
	lives.Lives--

	obj := components.Object.Get(e)
	centerX := obj.X + obj.W/2
	centerY := obj.Y + obj.H/2

	// Visual effects
	TriggerScreenShake(ecs, cfg.DeathZone.ScreenShakeIntensity, cfg.DeathZone.ScreenShakeDuration)
	factory.SpawnExplosion(ecs, centerX, centerY, cfg.DeathZone.ExplosionScale)
	PlaySFX(ecs, cfg.SoundHit)

	// Stop movement
	physics := components.Physics.Get(e)
	physics.SpeedX = 0
	physics.SpeedY = 0

	// Add death component with delay
	e.AddComponent(components.Death)
	components.Death.Set(e, &components.DeathData{
		Timer:       cfg.DeathZone.RespawnDelayFrames,
		IsDeathZone: true,
	})
}
