package systems

import (
	"image"
	"image/color"
	"math"
	"os"

	cfg "github.com/automoto/doomerang/config"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const (
	playerFriction = 0.5
	playerAccel    = 0.75
	playerMaxSpeed = 6.0
	playerJumpSpd  = 15.0
	playerGravity  = 0.75
)

func UpdatePlayer(ecs *ecs.ECS) {
	playerEntry, _ := components.Player.First(ecs.World)
	if playerEntry == nil {
		return
	}

	// If the player is in death sequence, only advance animation and return.
	if playerEntry.HasComponent(components.Death) {
		if anim := components.Animation.Get(playerEntry); anim != nil && anim.CurrentAnimation != nil {
			anim.CurrentAnimation.Update()
		}
		return
	}

	player := components.Player.Get(playerEntry)
	playerObject := cfg.GetObject(playerEntry)

	handlePlayerInput(player, playerObject)
	applyPlayerPhysics(player)
	resolvePlayerCollisions(player, playerObject)
	updatePlayerState(player, components.Animation.Get(playerEntry))
}

func handlePlayerInput(player *components.PlayerData, playerObject *resolv.Object) {
	// Only allow new actions if not in a locked state
	if !isInLockedState(player.CurrentState) {
		// Combat inputs
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) { // Punch
			startPunchCombo(player)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) && player.OnGround != nil { // Guard/Crouch
			if player.CurrentState != cfg.Crouch {
				player.CurrentState = cfg.Crouch
				player.StateTimer = 0
			}
		}
	}

	// Movement inputs - only allow if not in attack state
	if !isInAttackState(player.CurrentState) {
		// Horizontal movement is only possible when not wall-sliding.
		if player.WallSliding == nil {
			if ebiten.IsKeyPressed(ebiten.KeyRight) {
				player.SpeedX += playerAccel
				player.Direction.X = 1
			}

			if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				player.SpeedX -= playerAccel
				player.Direction.X = -1
			}
		}

		// Check for jumping.
		if inpututil.IsKeyJustPressed(ebiten.KeyX) || ebiten.IsGamepadButtonPressed(0, 0) || ebiten.IsGamepadButtonPressed(1, 0) {
			isTryingToDrop := ebiten.IsKeyPressed(ebiten.KeyDown)
			canDropDown := player.OnGround != nil && player.OnGround.HasTags("platform")

			if isTryingToDrop && canDropDown {
				player.IgnorePlatform = player.OnGround
			} else {
				if player.OnGround != nil {
					player.SpeedY = -playerJumpSpd
				} else if player.WallSliding != nil {
					// Wall-jumping
					player.SpeedY = -playerJumpSpd
					if player.WallSliding.X > playerObject.X {
						player.SpeedX = -playerMaxSpeed
					} else {
						player.SpeedX = playerMaxSpeed
					}
					player.WallSliding = nil
				}
			}
		}
	}
}

func applyPlayerPhysics(player *components.PlayerData) {
	// Apply friction and horizontal speed limiting.
	if player.SpeedX > playerFriction {
		player.SpeedX -= playerFriction
	} else if player.SpeedX < -playerFriction {
		player.SpeedX += playerFriction
	} else {
		player.SpeedX = 0
	}

	if player.SpeedX > playerMaxSpeed {
		player.SpeedX = playerMaxSpeed
	} else if player.SpeedX < -playerMaxSpeed {
		player.SpeedX = -playerMaxSpeed
	}

	player.SpeedY += playerGravity
	if player.WallSliding != nil && player.SpeedY > 1 {
		player.SpeedY = 1
	}
}

// resolveHorizontalCollision handles player horizontal movement and wall collision
func resolveHorizontalCollision(player *components.PlayerData, playerObject *resolv.Object) {
	dx := player.SpeedX
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
		player.SpeedX = 0
		setWallSlidingIfAirborne(player, check)
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
func resolveVerticalCollision(player *components.PlayerData, playerObject *resolv.Object) {
	player.OnGround = nil
	dy := clampVerticalSpeed(player.SpeedY)

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
		dy = handleUpwardCollision(player, playerObject, check)
	} else {
		dy = handleDownwardCollision(player, playerObject, check, dy)
	}

	playerObject.Y += dy
}

// updateWallSliding checks if player should disengage from wall sliding
func updateWallSliding(player *components.PlayerData, playerObject *resolv.Object) {
	if player.WallSliding == nil {
		return
	}

	wallDirection := player.Direction.X

	if check := playerObject.Check(wallDirection, 0, "solid"); check == nil {
		player.WallSliding = nil
	}
}

func resolvePlayerCollisions(player *components.PlayerData, playerObject *resolv.Object) {
	resolveHorizontalCollision(player, playerObject)
	resolveVerticalCollision(player, playerObject)
	updateWallSliding(player, playerObject)
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

func setWallSlidingIfAirborne(player *components.PlayerData, check *resolv.Collision) {
	if player.OnGround != nil {
		return
	}

	if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
		player.WallSliding = solids[0]
	}
}

func clampVerticalSpeed(speedY float64) float64 {
	return math.Max(math.Min(speedY, 16), -16)
}

func handleUpwardCollision(player *components.PlayerData, playerObject *resolv.Object, check *resolv.Collision) float64 {
	if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
		player.SpeedY = 0
		return check.ContactWithObject(solids[0]).Y()
	}

	if len(check.Cells) > 0 && check.Cells[0].ContainsTags("solid") {
		if slide := check.SlideAgainstCell(check.Cells[0], "solid"); slide != nil {
			playerObject.X += slide.X()
		}
	}

	return player.SpeedY
}

func handleDownwardCollision(player *components.PlayerData, playerObject *resolv.Object, check *resolv.Collision, dy float64) float64 {
	// Try collision in priority order: ramps, platforms, solids
	if newDy, handled := tryRampCollision(player, playerObject, check, dy); handled {
		return newDy
	}

	if newDy, handled := tryPlatformCollision(player, playerObject, check); handled {
		return newDy
	}

	if newDy, handled := trySolidCollision(player, check); handled {
		return newDy
	}

	return dy
}

func tryRampCollision(player *components.PlayerData, playerObject *resolv.Object, check *resolv.Collision, dy float64) (float64, bool) {
	ramps := check.ObjectsByTags("ramp")
	if len(ramps) == 0 {
		return dy, false
	}

	ramp := ramps[0]
	contactSet := playerObject.Shape.Intersection(0, 8, ramp.Shape)

	if dy >= 0 && contactSet != nil {
		player.OnGround = ramp
		player.SpeedY = 0
		return contactSet.TopmostPoint()[1] - playerObject.Bottom() + 0.1, true
	}

	return dy, false
}

func tryPlatformCollision(player *components.PlayerData, playerObject *resolv.Object, check *resolv.Collision) (float64, bool) {
	if player.OnGround != nil {
		return 0, false // Already grounded from ramp
	}

	platforms := check.ObjectsByTags("platform")
	if len(platforms) == 0 {
		return 0, false
	}

	platform := platforms[0]

	// Check platform collision conditions
	if platform == player.IgnorePlatform ||
		player.SpeedY < 0 ||
		playerObject.Bottom() >= platform.Y+4 {
		return 0, false
	}

	player.OnGround = platform
	player.SpeedY = 0
	return check.ContactWithObject(platform).Y(), true
}

func trySolidCollision(player *components.PlayerData, check *resolv.Collision) (float64, bool) {
	if player.OnGround != nil {
		clearGroundedState(player)
		return 0, false // Already grounded
	}

	solids := check.ObjectsByTags("solid")
	if len(solids) == 0 {
		return 0, false
	}

	solid := solids[0]

	// Only land on solid if falling down
	if player.SpeedY >= 0 {
		player.OnGround = solid
		player.SpeedY = 0
		clearGroundedState(player)
		return check.ContactWithObject(solid).Y(), true
	}

	return 0, false
}

func clearGroundedState(player *components.PlayerData) {
	if player.OnGround != nil {
		player.WallSliding = nil
		player.IgnorePlatform = nil
	}
}

func updatePlayerState(player *components.PlayerData, animData *components.AnimationData) {
	player.StateTimer++

	// Handle state transitions based on current state
	switch player.CurrentState {
	case cfg.Punch01:
		if player.StateTimer > 30 { // 30 frames for punch1 animation
			transitionToMovementState(player)
		}
	case cfg.Punch02:
		if player.StateTimer > 20 { // 20 frames for punch2 animation
			transitionToMovementState(player)
		}
	case cfg.Punch03:
		if player.StateTimer > 35 { // 35 frames for punch3 animation
			transitionToMovementState(player)
		}
	case cfg.Kick01:
		if player.StateTimer > 45 { // 45 frames for kick1 animation
			transitionToMovementState(player)
		}
	case cfg.Hit, cfg.Stunned:
		if player.StateTimer > 30 { // 30 frames of hitstun
			transitionToMovementState(player)
		}
	case cfg.Crouch:
		if !ebiten.IsKeyPressed(ebiten.KeyDown) {
			transitionToMovementState(player)
		}
	default:
		// Handle movement states based on physics
		transitionToMovementState(player)
	}

	// Update animation based on current state
	if animData.CurrentAnimation != animData.Animations[player.CurrentState] {
		animData.SetAnimation(player.CurrentState)
	}

	if animData.CurrentAnimation != nil {
		animData.CurrentAnimation.Update()
	}
}

// Helper functions for state management
func isInLockedState(state string) bool {
	return state == cfg.Hit || state == cfg.Stunned || state == cfg.Knockback
}

func isInAttackState(state string) bool {
	return state == cfg.Punch01 || state == cfg.Punch02 || state == cfg.Punch03 || state == cfg.Kick01
}

func startPunchCombo(player *components.PlayerData) {
	switch player.CurrentState {
	case cfg.Punch01:
		if player.StateTimer > 10 { // Allow combo after 10 frames
			player.CurrentState = cfg.Punch02
			player.ComboCounter++
			player.StateTimer = 0
		}
	case cfg.Punch02:
		if player.StateTimer > 8 { // Allow combo after 8 frames
			player.CurrentState = cfg.Kick01
			player.ComboCounter++
			player.StateTimer = 0
		}
	default:
		player.CurrentState = cfg.Punch01
		player.ComboCounter = 1
		player.StateTimer = 0
	}
}


func transitionToMovementState(player *components.PlayerData) {
	if player.WallSliding != nil {
		player.CurrentState = cfg.WallSlide
	} else if player.OnGround == nil {
		player.CurrentState = cfg.Jump
	} else if player.SpeedX != 0 {
		player.CurrentState = cfg.Running
	} else {
		player.CurrentState = cfg.Idle
	}
	player.StateTimer = 0
	player.ComboCounter = 0
}

func DrawPlayer(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	tags.Player.Each(ecs.World, func(e *donburi.Entry) {
		player := components.Player.Get(e)
		o := cfg.GetObject(e)
		animData := components.Animation.Get(e)

		if animData.CurrentAnimation != nil && animData.SpriteSheets[animData.CurrentSheet] != nil {
			// Calculate the source rectangle for the current frame.
			frame := animData.CurrentAnimation.Frame()
			sx := frame * animData.FrameWidth
			sy := 0
			srcRect := image.Rect(sx, sy, sx+animData.FrameWidth, sy+animData.FrameHeight)

			// Create draw options.
			op := &ebiten.DrawImageOptions{}

			// Anchor the sprite at its bottom-center so that the feet line up with the
			// bottom of the collision box.
			op.GeoM.Translate(-float64(animData.FrameWidth)/2, -float64(animData.FrameHeight))

			// Flip the sprite if facing left.
			if player.Direction.X < 0 {
				op.GeoM.Scale(-1, 1)
			}

			// Move the sprite so that its bottom-center aligns with the bottom-center
			// of the (smaller) collision box.
			op.GeoM.Translate(o.X+o.W/2, o.Y+o.H)

			// Apply the camera translation.
			op.GeoM.Translate(float64(width)/2-camera.Position.X, float64(height)/2-camera.Position.Y)

			// Draw the current frame.
			screen.DrawImage(animData.SpriteSheets[animData.CurrentSheet].SubImage(srcRect).(*ebiten.Image), op)
		} else {
			// Fallback to rectangle if no animation is available
			playerColor := color.RGBA{0, 255, 60, 255}
			if player.OnGround == nil {
				playerColor = color.RGBA{200, 0, 200, 255}
			}
			// This debug draw doesn't need to be camera-aware, as it's for debugging.
			vector.DrawFilledRect(screen, float32(o.X), float32(o.Y), float32(o.W), float32(o.H), playerColor, false)
		}
	})
}
