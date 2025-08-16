package systems

import (
	cfg "github.com/automoto/doomerang/config"

	"github.com/automoto/doomerang/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const (
	playerJumpSpd = 15.0
	playerAccel   = 0.75
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
	physics := components.Physics.Get(playerEntry)
	playerObject := cfg.GetObject(playerEntry)

	handlePlayerInput(player, physics, components.State.Get(playerEntry), playerObject)
	updatePlayerState(playerEntry, player, physics, components.State.Get(playerEntry), components.Animation.Get(playerEntry))
}

func handlePlayerInput(player *components.PlayerData, physics *components.PhysicsData, state *components.StateData, playerObject *resolv.Object) {
	// Only allow new actions if not in a locked state
	if !isInLockedState(state.CurrentState) {
		// Combat inputs
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) { // Punch
			startPunchCombo(player, state)
		}
		if ebiten.IsKeyPressed(ebiten.KeyDown) && physics.OnGround != nil { // Guard/Crouch
			if state.CurrentState != cfg.Crouch {
				state.CurrentState = cfg.Crouch
				state.StateTimer = 0
			}
		}
	}

	// Movement inputs - only allow if not in attack state
	if !isInAttackState(state.CurrentState) {
		// Horizontal movement is only possible when not wall-sliding.
		if physics.WallSliding == nil {
			if ebiten.IsKeyPressed(ebiten.KeyRight) {
				physics.SpeedX += playerAccel
				player.Direction.X = 1
			}

			if ebiten.IsKeyPressed(ebiten.KeyLeft) {
				physics.SpeedX -= playerAccel
				player.Direction.X = -1
			}
		}

		// Check for jumping.
		if inpututil.IsKeyJustPressed(ebiten.KeyX) || ebiten.IsGamepadButtonPressed(0, 0) || ebiten.IsGamepadButtonPressed(1, 0) {
			isTryingToDrop := ebiten.IsKeyPressed(ebiten.KeyDown)
			canDropDown := physics.OnGround != nil && physics.OnGround.HasTags("platform")

			if isTryingToDrop && canDropDown {
				physics.IgnorePlatform = physics.OnGround
			} else {
				if physics.OnGround != nil {
					physics.SpeedY = -playerJumpSpd
				} else if physics.WallSliding != nil {
					// Wall-jumping
					physics.SpeedY = -playerJumpSpd
					if physics.WallSliding.X > playerObject.X {
						physics.SpeedX = -physics.MaxSpeed
					} else {
						physics.SpeedX = physics.MaxSpeed
					}
					physics.WallSliding = nil
				}
			}
		}
	}
}


func updatePlayerState(playerEntry *donburi.Entry, player *components.PlayerData, physics *components.PhysicsData, state *components.StateData, animData *components.AnimationData) {
	state.StateTimer++

	// Handle state transitions based on current state
	switch state.CurrentState {
	case cfg.Punch01:
		if state.StateTimer > 30 { // 30 frames for punch1 animation
			transitionToMovementState(playerEntry, player, physics, state)
		}
	case cfg.Punch02:
		if state.StateTimer > 20 { // 20 frames for punch2 animation
			transitionToMovementState(playerEntry, player, physics, state)
		}
	case cfg.Punch03:
		if state.StateTimer > 35 { // 35 frames for punch3 animation
			transitionToMovementState(playerEntry, player, physics, state)
		}
	case cfg.Kick01:
		if state.StateTimer > 45 { // 45 frames for kick1 animation
			transitionToMovementState(playerEntry, player, physics, state)
		}
	case cfg.Hit, cfg.Stunned:
		if state.StateTimer > 30 { // 30 frames of hitstun
			transitionToMovementState(playerEntry, player, physics, state)
		}
	case cfg.Crouch:
		if !ebiten.IsKeyPressed(ebiten.KeyDown) {
			transitionToMovementState(playerEntry, player, physics, state)
		}
	default:
		// Handle movement states based on physics
		transitionToMovementState(playerEntry, player, physics, state)
	}

	// Update animation based on current state
	if animData.CurrentAnimation != animData.Animations[state.CurrentState] {
		animData.SetAnimation(state.CurrentState)
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

func startPunchCombo(player *components.PlayerData, state *components.StateData) {
	switch state.CurrentState {
	case cfg.Punch01:
		if state.StateTimer > 10 { // Allow combo after 10 frames
			state.CurrentState = cfg.Punch02
			player.ComboCounter++
			state.StateTimer = 0
		}
	case cfg.Punch02:
		if state.StateTimer > 8 { // Allow combo after 8 frames
			state.CurrentState = cfg.Kick01
			player.ComboCounter++
			state.StateTimer = 0
		}
	default:
		state.CurrentState = cfg.Punch01
		player.ComboCounter = 1
		state.StateTimer = 0
	}
}

func transitionToMovementState(e *donburi.Entry, player *components.PlayerData, physics *components.PhysicsData, state *components.StateData) {
	if physics.WallSliding != nil {
		state.CurrentState = cfg.WallSlide
	} else if physics.OnGround == nil {
		if physics.SpeedY > 0 {
			state.CurrentState = cfg.Jump // There is no falling animation yet
		} else {
			state.CurrentState = cfg.Jump
		}
	} else if physics.SpeedX != 0 {
		state.CurrentState = cfg.Running
	} else {
		state.CurrentState = cfg.Idle
	}
	state.StateTimer = 0
	player.ComboCounter = 0
}

