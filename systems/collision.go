package systems

import (
	"math"
	"os"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateCollisions(ecs *ecs.ECS) {
	tags.Player.Each(ecs.World, func(e *donburi.Entry) {
		player := components.Player.Get(e)
		physics := components.Physics.Get(e)
		obj := components.Object.Get(e)

		resolveObjectHorizontalCollision(physics, obj.Object, true)
		resolveObjectVerticalCollision(physics, obj.Object)
		updateWallSliding(player, physics, obj.Object)
	})

	tags.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		physics := components.Physics.Get(e)
		obj := components.Object.Get(e)

		resolveObjectHorizontalCollision(physics, obj.Object, false)
		resolveObjectVerticalCollision(physics, obj.Object)
	})
}

// resolveObjectHorizontalCollision handles horizontal movement and wall collision for any object
func resolveObjectHorizontalCollision(physics *components.PhysicsData, object *resolv.Object, allowWallSlide bool) {
	dx := physics.SpeedX
	if dx == 0 {
		return
	}

	check := object.Check(dx, 0, "solid", "character")
	if check == nil {
		object.X += dx
		return
	}

	debugHorizontalCollision(dx, object, check)

	if shouldStopHorizontalMovement(object, check) {
		physics.SpeedX = 0
		if allowWallSlide {
			setWallSlidingIfAirborne(physics, check)
		}
		dx = 0
	}

	dx = handleCharacterCollision(check, dx)
	object.X += dx
}

func handleCharacterCollision(check *resolv.Collision, dx float64) float64 {
	characters := check.ObjectsByTags("character")
	if len(characters) == 0 {
		return dx
	}

	contact := check.ContactWithObject(characters[0])
	if contact.X() == 0 {
		return contact.X()
	}

	// Apply small fixed pushback in opposite direction
	if dx > 0 {
		return -1
	}
	return 1
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

	objectCenterY := object.Y + object.H/2

	for _, solid := range solids {
		// Only stop if object's center would be within solid's vertical bounds
		if objectCenterY >= solid.Y && objectCenterY <= solid.Y+solid.H {
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
	ramps := check.ObjectsByTags("ramp")
	if len(ramps) == 0 {
		return dy, false
	}

	ramp := ramps[0]
	contactSet := object.Shape.Intersection(0, 8, ramp.Shape)

	if dy >= 0 && contactSet != nil {
		physics.OnGround = ramp
		physics.SpeedY = 0
		return contactSet.TopmostPoint()[1] - object.Bottom() + 0.1, true
	}

	return dy, false
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
