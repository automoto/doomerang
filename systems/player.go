package systems

import (
	"fmt"
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
		if inpututil.IsKeyJustPressed(ebiten.KeyC) { // Kick
			startKickCombo(player)
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
				player.FacingRight = true
			}

			if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				player.SpeedX -= playerAccel
				player.FacingRight = false
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

func resolvePlayerCollisions(player *components.PlayerData, playerObject *resolv.Object) {
	// Handle horizontal movement and collision
	dx := player.SpeedX
	if dx != 0 {
		// Check if horizontal movement would cause a collision
		if check := playerObject.Check(dx, 0, "solid"); check != nil {
			// Debug output (only if DEBUG_COLLISION env var is set)
			if os.Getenv("DEBUG_COLLISION") != "" {
				fmt.Printf("Horizontal collision detected! dx=%.2f, player pos: (%.2f, %.2f)\n", dx, playerObject.X, playerObject.Y)
				if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
					for i, solid := range solids {
						fmt.Printf("  Solid %d: pos=(%.2f, %.2f), size=(%.2f, %.2f)\n", i, solid.X, solid.Y, solid.W, solid.H)
					}
				}
			}

			// Check if we're actually colliding with a wall (not just positioned next to ground)
			shouldStop := false
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				for _, solid := range solids {
					// Check if this solid object is actually blocking horizontal movement
					// by testing if the player's center would be inside the solid after movement
					playerCenterY := playerObject.Y + playerObject.H/2
					solidTop := solid.Y
					solidBottom := solid.Y + solid.H

					// Only stop if the player's vertical center would be within the solid's vertical bounds
					if playerCenterY >= solidTop && playerCenterY <= solidBottom {
						shouldStop = true
						break
					}
				}
			}

			if shouldStop {
				dx = 0
				player.SpeedX = 0

				// Set wall sliding only if player is in the air
				if player.OnGround == nil {
					if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
						player.WallSliding = solids[0]
					}
				}
			}
		}
	}
	playerObject.X += dx

	// Handle vertical movement and collision
	player.OnGround = nil
	dy := player.SpeedY
	dy = math.Max(math.Min(dy, 16), -16)

	checkDistance := dy
	if dy >= 0 {
		checkDistance++
	}

	if check := playerObject.Check(0, checkDistance, "solid", "platform", "ramp"); check != nil {
		// Handle upward collision with solid objects
		if dy < 0 {
			if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
				dy = check.ContactWithObject(solids[0]).Y()
				player.SpeedY = 0
			} else if len(check.Cells) > 0 && check.Cells[0].ContainsTags("solid") {
				if slide := check.SlideAgainstCell(check.Cells[0], "solid"); slide != nil {
					playerObject.X += slide.X()
				}
			}
		} else {
			// Handle downward collision - check in order of priority
			// 1. Ramps first
			if ramps := check.ObjectsByTags("ramp"); len(ramps) > 0 {
				ramp := ramps[0]
				if contactSet := playerObject.Shape.Intersection(dx, 8, ramp.Shape); dy >= 0 && contactSet != nil {
					dy = contactSet.TopmostPoint()[1] - playerObject.Bottom() + 0.1
					player.OnGround = ramp
					player.SpeedY = 0
				}
			}

			// 2. Platforms second (if no ramp collision)
			if player.OnGround == nil {
				if platforms := check.ObjectsByTags("platform"); len(platforms) > 0 {
					platform := platforms[0]
					if platform != player.IgnorePlatform && player.SpeedY >= 0 && playerObject.Bottom() < platform.Y+4 {
						dy = check.ContactWithObject(platform).Y()
						player.OnGround = platform
						player.SpeedY = 0
					}
				}
			}

			// 3. Solid ground last (if no other collision)
			if player.OnGround == nil {
				if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
					solid := solids[0]
					// Make sure we're falling down onto the solid object
					if player.SpeedY >= 0 {
						dy = check.ContactWithObject(solid).Y()
						player.SpeedY = 0
						player.OnGround = solid
					}
				}
			}

			// Clear wall sliding and ignore platform when on ground
			if player.OnGround != nil {
				player.WallSliding = nil
				player.IgnorePlatform = nil
			}
		}
	}
	playerObject.Y += dy

	// Check for wall sliding disengage
	wallNext := 1.0
	if !player.FacingRight {
		wallNext = -1
	}

	if c := playerObject.Check(wallNext, 0, "solid"); player.WallSliding != nil && c == nil {
		player.WallSliding = nil
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
	case cfg.Kick02:
		if player.StateTimer > 40 { // 40 frames for kick2 animation
			transitionToMovementState(player)
		}
	case cfg.Kick03:
		if player.StateTimer > 45 { // 45 frames for kick3 animation
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
	return state == cfg.Punch01 || state == cfg.Punch02 || state == cfg.Punch03 ||
		state == cfg.Kick01 || state == cfg.Kick02 || state == cfg.Kick03
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
			player.CurrentState = cfg.Punch03
			player.ComboCounter++
			player.StateTimer = 0
		}
	default:
		player.CurrentState = cfg.Punch01
		player.ComboCounter = 1
		player.StateTimer = 0
	}
}

func startKickCombo(player *components.PlayerData) {
	switch player.CurrentState {
	case cfg.Kick01:
		if player.StateTimer > 12 { // Allow combo after 12 frames
			player.CurrentState = cfg.Kick02
			player.ComboCounter++
			player.StateTimer = 0
		}
	case cfg.Kick02:
		if player.StateTimer > 10 { // Allow combo after 10 frames
			player.CurrentState = cfg.Kick03
			player.ComboCounter++
			player.StateTimer = 0
		}
	default:
		player.CurrentState = cfg.Kick01
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
			if !player.FacingRight {
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
