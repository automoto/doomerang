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
	playerAccel        = 0.75
	playerAttackAccel = 0.1
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
	melee := components.MeleeAttack.Get(playerEntry)
	playerObject := cfg.GetObject(playerEntry)

	handlePlayerInput(player, physics, melee, components.State.Get(playerEntry), playerObject)
	updatePlayerState(playerEntry, player, physics, melee, components.State.Get(playerEntry), components.Animation.Get(playerEntry))

	// Decrement invulnerability timer
	if player.InvulnFrames > 0 {
		player.InvulnFrames--
	}
}

func handlePlayerInput(player *components.PlayerData, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData, playerObject *resolv.Object) {
	// Only allow new actions if not in a locked state
	if !isInLockedState(state.CurrentState) {
		// Combat inputs
		// Melee attack
		if inpututil.IsKeyJustPressed(ebiten.KeyZ) {
			if physics.OnGround == nil {
				// if in the air, and not already attacking, do a jump attack
				if !isInAttackState(state.CurrentState) {
					state.CurrentState = cfg.StateAttackingJump
					state.StateTimer = 0
					melee.IsAttacking = true
				}
			} else {
				melee.IsCharging = true
				melee.ChargeTime = 0
			}
		}

		// Attack release
		if melee.IsCharging && inpututil.IsKeyJustReleased(ebiten.KeyZ) {
			melee.IsCharging = false
			melee.IsAttacking = true
		}

		if ebiten.IsKeyPressed(ebiten.KeyDown) && physics.OnGround != nil { // Guard/Crouch
			// Logic to be moved to updatePlayerState
		}

		// Jumping - only allow if not in attack state
		if !isInAttackState(state.CurrentState) {
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

	// Horizontal movement
	accel := playerAccel
	if isInAttackState(state.CurrentState) {
		accel = playerAttackAccel
	}

	if physics.WallSliding == nil {
		if ebiten.IsKeyPressed(ebiten.KeyRight) {
			physics.SpeedX += accel
			player.Direction.X = 1
		}
		if ebiten.IsKeyPressed(ebiten.KeyLeft) {
			physics.SpeedX -= accel
			player.Direction.X = -1
		}
	}
}

func updatePlayerState(playerEntry *donburi.Entry, player *components.PlayerData, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData, animData *components.AnimationData) {
	state.StateTimer++

	// Main state machine logic
	switch state.CurrentState {
	case cfg.Idle, cfg.Running:
		// Transition to charging
		if melee.IsCharging {
			state.CurrentState = cfg.StateChargingAttack
			state.StateTimer = 0
		} else if ebiten.IsKeyPressed(ebiten.KeyDown) && physics.OnGround != nil {
			state.CurrentState = cfg.Crouch
			state.StateTimer = 0
		} else {
			transitionToMovementState(playerEntry, player, physics, state)
		}

	case cfg.StateChargingAttack:
		// Transition to attacking when charge is released
		if !melee.IsCharging {
			if melee.IsAttacking {
				if melee.ComboStep == 0 {
					melee.ComboStep = 1
					state.CurrentState = cfg.StateAttackingPunch
				} else {
					melee.ComboStep = 0
					state.CurrentState = cfg.StateAttackingKick
				}
				state.StateTimer = 0
			} else {
				// If button is released without attacking (e.g. interrupted)
				transitionToMovementState(playerEntry, player, physics, state)
			}
		} else {
			melee.ChargeTime++
		}

	case cfg.StateAttackingPunch, cfg.StateAttackingKick:
		// Transition back to movement after attack animation finishes
		if animData.CurrentAnimation != nil && animData.CurrentAnimation.Looped {
			melee.IsAttacking = false
			transitionToMovementState(playerEntry, player, physics, state)
		}

	case cfg.StateAttackingJump:
		// Transition back to jump after attack animation finishes
		if animData.CurrentAnimation != nil && animData.CurrentAnimation.Looped {
			melee.IsAttacking = false
			state.CurrentState = cfg.Jump
			state.StateTimer = 0
		}

	case cfg.Hit, cfg.Stunned, cfg.Knockback:
		// Transition back to movement after hitstun/knockback duration
		if state.StateTimer > 15 {
			transitionToMovementState(playerEntry, player, physics, state)
		}

	case cfg.Crouch:
		// Transition back to movement when down key is released
		if !ebiten.IsKeyPressed(ebiten.KeyDown) {
			transitionToMovementState(playerEntry, player, physics, state)
		}

	case cfg.Jump:
		// Transition to idle/running when landing on the ground
		if physics.OnGround != nil {
			transitionToMovementState(playerEntry, player, physics, state)
		} else if physics.WallSliding != nil {
			state.CurrentState = cfg.WallSlide
			state.StateTimer = 0
		}

	default:
		// Default to movement state for any unhandled cases
		transitionToMovementState(playerEntry, player, physics, state)
	}

	// --- Animation Update ---
	var anim string
	switch state.CurrentState {
	case cfg.StateAttackingPunch:
		anim = cfg.Punch01
	case cfg.StateAttackingKick:
		anim = cfg.Kick01
	case cfg.StateAttackingJump:
		anim = cfg.Kick02
	default:
		anim = state.CurrentState
	}

	// Check if the animation exists before setting it
	if newAnim, ok := animData.Animations[anim]; ok {
		if animData.CurrentAnimation != newAnim {
			animData.SetAnimation(anim)
		}
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
	return state == cfg.StateAttackingPunch || state == cfg.StateAttackingKick || state == cfg.StateAttackingJump
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
	} else if state.CurrentState == cfg.Jump {
		if physics.SpeedX != 0 {
			state.CurrentState = cfg.Running
		} else {
			state.CurrentState = cfg.Idle
		}
	} else if physics.SpeedX != 0 {
		state.CurrentState = cfg.Running
	} else {
		state.CurrentState = cfg.Idle
	}
	state.StateTimer = 0
	player.ComboCounter = 0
}
