package systems

import (
	"math"
	"os"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateCollisions(ecs *ecs.ECS) {
	tags.Player.Each(ecs.World, func(e *donburi.Entry) {
		player := components.Player.Get(e)
		physics := components.Physics.Get(e)
		obj := cfg.GetObject(e)
		resolvePlayerCollisions(player, physics, obj)
	})
	tags.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		enemy := components.Enemy.Get(e)
		physics := components.Physics.Get(e)
		obj := cfg.GetObject(e)
		resolveEnemyCollisions(enemy, physics, obj)
	})
}

// resolveHorizontalCollision handles player horizontal movement and wall collision
func resolveHorizontalCollision(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object) {
	dx := physics.SpeedX
	if dx == 0 {
		return
	}

	check := playerObject.Check(dx, 0, "solid", "character")
	if check == nil {
		playerObject.X += dx
		return
	}

	// Debug collision detection if enabled
	debugHorizontalCollision(dx, playerObject, check)

	// Check for collisions with solid objects (walls)
	if shouldStopHorizontalMovement(playerObject, check) {
		physics.SpeedX = 0
		setWallSlidingIfAirborne(player, physics, check)
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

	playerObject.X += dx
}

// resolveVerticalCollision handles player vertical movement and ground/platform collision
func resolveVerticalCollision(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object) {
	physics.OnGround = nil
	dy := clampVerticalSpeed(physics.SpeedY)

	checkDistance := dy
	if dy >= 0 {
		checkDistance++
	}

	check := playerObject.Check(0, checkDistance, "solid", "platform", "ramp")
	if check == nil {
		playerObject.Y += dy
		return
	}

	if dy < 0 {
		dy = handleUpwardCollision(player, physics, playerObject, check)
	} else {
		dy = handleDownwardCollision(player, physics, playerObject, check, dy)
	}

	playerObject.Y += dy
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

func resolvePlayerCollisions(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object) {
	resolveHorizontalCollision(player, physics, playerObject)
	resolveVerticalCollision(player, physics, playerObject)
	updateWallSliding(player, physics, playerObject)
}

// Helper functions for collision resolution

func debugHorizontalCollision(dx float64, playerObject *resolv.Object, check *resolv.Collision) {
	if os.Getenv("DEBUG_COLLISION") == "" {
		return
	}

	// fmt.Printf("Horizontal collision detected! dx=%.2f, player pos: (%.2f, %.2f)\n",
	// 	dx, playerObject.X, playerObject.Y)

	// if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
	// 	for i, solid := range solids {
	// 		fmt.Printf("  Solid %d: pos=(%.2f, %.2f), size=(%.2f, %.2f)\n",
	// 			i, solid.X, solid.Y, solid.W, solid.H)
	// 	}
	// }
}

func shouldStopHorizontalMovement(playerObject *resolv.Object, check *resolv.Collision) bool {
	solids := check.ObjectsByTags("solid")
	if len(solids) == 0 {
		return false
	}

	playerCenterY := playerObject.Y + playerObject.H/2

	for _, solid := range solids {
		// Only stop if player's center would be within solid's vertical bounds
		if playerCenterY >= solid.Y && playerCenterY <= solid.Y+solid.H {
			return true
		}
	}

	return false
}

func setWallSlidingIfAirborne(player *components.PlayerData, physics *components.PhysicsData, check *resolv.Collision) {
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

func handleUpwardCollision(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object, check *resolv.Collision) float64 {
	if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
		physics.SpeedY = 0
		return check.ContactWithObject(solids[0]).Y()
	}

	if len(check.Cells) > 0 && check.Cells[0].ContainsTags("solid") {
		if slide := check.SlideAgainstCell(check.Cells[0], "solid"); slide != nil {
			playerObject.X += slide.X()
		}
	}

	return physics.SpeedY
}

func handleDownwardCollision(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object, check *resolv.Collision, dy float64) float64 {
	// Try collision in priority order: ramps, platforms, solids
	if newDy, handled := tryRampCollision(player, physics, playerObject, check, dy); handled {
		return newDy
	}

	if newDy, handled := tryPlatformCollision(player, physics, playerObject, check); handled {
		return newDy
	}

	if newDy, handled := trySolidCollision(player, physics, check); handled {
		return newDy
	}

	return dy
}

func tryRampCollision(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object, check *resolv.Collision, dy float64) (float64, bool) {
	ramps := check.ObjectsByTags("ramp")
	if len(ramps) == 0 {
		return dy, false
	}

	ramp := ramps[0]
	contactSet := playerObject.Shape.Intersection(0, 8, ramp.Shape)

	if dy >= 0 && contactSet != nil {
		physics.OnGround = ramp
		physics.SpeedY = 0
		return contactSet.TopmostPoint()[1] - playerObject.Bottom() + 0.1, true
	}

	return dy, false
}

func tryPlatformCollision(player *components.PlayerData, physics *components.PhysicsData, playerObject *resolv.Object, check *resolv.Collision) (float64, bool) {
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
		playerObject.Bottom() >= platform.Y+4 {
		return 0, false
	}

	physics.OnGround = platform
	physics.SpeedY = 0
	return check.ContactWithObject(platform).Y(), true
}

func trySolidCollision(player *components.PlayerData, physics *components.PhysicsData, check *resolv.Collision) (float64, bool) {
	if physics.OnGround != nil {
		clearGroundedState(player, physics)
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
		clearGroundedState(player, physics)
		return check.ContactWithObject(solid).Y(), true
	}

	return 0, false
}

func clearGroundedState(player *components.PlayerData, physics *components.PhysicsData) {
	if physics.OnGround != nil {
		physics.WallSliding = nil
		physics.IgnorePlatform = nil
	}
}

func resolveEnemyCollisions(enemy *components.EnemyData, physics *components.PhysicsData, enemyObject *resolv.Object) {
	// Horizontal collision - only stop for actual walls, not ground
	dx := physics.SpeedX
	if dx != 0 {
		if check := enemyObject.Check(dx, 0, "solid", "character"); check != nil {
			// Check for collisions with solid objects (walls)
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				shouldStop := false
				for _, solid := range solids {
					enemyCenterY := enemyObject.Y + enemyObject.H/2
					if enemyCenterY >= solid.Y && enemyCenterY <= solid.Y+solid.H {
						shouldStop = true
						break
					}
				}
				if shouldStop {
					physics.SpeedX = 0
					dx = 0
				}
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
		}
	}
	enemyObject.X += dx

	// Vertical collision - simplified ground detection
	physics.OnGround = nil
	dy := physics.SpeedY
	dy = math.Max(math.Min(dy, 16), -16)

	if check := enemyObject.Check(0, dy, "solid", "platform"); check != nil {
		if dy > 0 { // Falling down
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				dy = check.ContactWithObject(solids[0]).Y()
				physics.OnGround = solids[0]
				physics.SpeedY = 0
			} else if platforms := check.ObjectsByTags("platform"); len(platforms) > 0 {
				platform := platforms[0]
				if physics.SpeedY >= 0 && enemyObject.Bottom() < platform.Y+4 {
					dy = check.ContactWithObject(platform).Y()
					physics.OnGround = platform
					physics.SpeedY = 0
				}
			}
		} else { // Moving up
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				dy = check.ContactWithObject(solids[0]).Y()
				physics.SpeedY = 0
			}
		}
	}
	enemyObject.Y += dy
}
