package systems

import (
	cfg "github.com/automoto/doomerang/config"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/systems/factory"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdatePlayer(ecs *ecs.ECS) {
	playerEntry, ok := components.Player.First(ecs.World)
	if !ok {
		return
	}

	// If the player is in death sequence, only advance animation and return.
	// The entity will be removed by the death system.
	if playerEntry.HasComponent(components.Death) {
		if anim := components.Animation.Get(playerEntry); anim != nil && anim.CurrentAnimation != nil {
			anim.CurrentAnimation.Update()
		}
		return
	}

	// Get input state
	input := getOrCreateInput(ecs)

	player := components.Player.Get(playerEntry)
	physics := components.Physics.Get(playerEntry)
	melee := components.MeleeAttack.Get(playerEntry)
	state := components.State.Get(playerEntry)
	animData := components.Animation.Get(playerEntry)
	playerObject := components.Object.Get(playerEntry).Object

	handlePlayerInput(input, player, physics, melee, state, playerObject)
	updatePlayerState(ecs, input, playerEntry, player, physics, melee, state, animData)

	// Decrement invulnerability timer
	if player.InvulnFrames > 0 {
		player.InvulnFrames--
	}
}

func handlePlayerInput(input *components.InputData, player *components.PlayerData, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData, playerObject *resolv.Object) {
	// Get action states from input component
	attackAction := input.Actions[cfg.ActionAttack]
	jumpAction := input.Actions[cfg.ActionJump]
	crouchAction := input.Actions[cfg.ActionCrouch]
	moveLeftAction := input.Actions[cfg.ActionMoveLeft]
	moveRightAction := input.Actions[cfg.ActionMoveRight]

	// Process combat and jump inputs only if not in a locked state
	if !isInLockedState(state.CurrentState) {
		handleMeleeInput(attackAction, physics, melee, state)

		if !isInAttackState(state.CurrentState) {
			handleJumpInput(jumpAction, crouchAction, physics, playerObject)
		}
	}

	// Horizontal movement (always processed)
	handleMovementInput(moveLeftAction, moveRightAction, player, physics, state)
}

func handleMeleeInput(attackAction components.ActionState, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData) {
	// Attack release
	if melee.IsCharging && attackAction.JustReleased {
		melee.IsCharging = false
		melee.IsAttacking = true
	}

	if !attackAction.JustPressed {
		return
	}

	// On ground - start charging
	if physics.OnGround != nil {
		melee.IsCharging = true
		melee.ChargeTime = 0
		return
	}

	// In air - jump attack if not already attacking
	if !isInAttackState(state.CurrentState) {
		state.CurrentState = cfg.StateAttackingJump
		state.StateTimer = 0
		melee.IsAttacking = true
	}
}

func handleJumpInput(jumpAction, crouchAction components.ActionState, physics *components.PhysicsData, playerObject *resolv.Object) {
	if !jumpAction.JustPressed {
		return
	}

	// Drop-through platform
	if crouchAction.Pressed && physics.OnGround != nil && physics.OnGround.HasTags("platform") {
		physics.IgnorePlatform = physics.OnGround
		return
	}

	// Normal jump from ground
	if physics.OnGround != nil {
		physics.SpeedY = -cfg.Player.JumpSpeed
		return
	}

	// Wall jump
	if physics.WallSliding == nil {
		return
	}
	physics.SpeedY = -cfg.Player.JumpSpeed
	if physics.WallSliding.X > playerObject.X {
		physics.SpeedX = -physics.MaxSpeed
	} else {
		physics.SpeedX = physics.MaxSpeed
	}
	physics.WallSliding = nil
}

func handleMovementInput(moveLeftAction, moveRightAction components.ActionState, player *components.PlayerData, physics *components.PhysicsData, state *components.StateData) {
	if physics.WallSliding != nil {
		return
	}

	accel := cfg.Player.Acceleration
	if isInAttackState(state.CurrentState) {
		accel = cfg.Player.AttackAccel
	}

	if moveRightAction.Pressed {
		physics.SpeedX += accel
		player.Direction.X = cfg.DirectionRight
	}
	if moveLeftAction.Pressed {
		physics.SpeedX -= accel
		player.Direction.X = cfg.DirectionLeft
	}
}

func updatePlayerState(ecs *ecs.ECS, input *components.InputData, playerEntry *donburi.Entry, player *components.PlayerData, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData, animData *components.AnimationData) {
	state.StateTimer++

	// Get action states from input component
	boomerangAction := input.Actions[cfg.ActionBoomerang]
	crouchAction := input.Actions[cfg.ActionCrouch]

	// Main state machine logic
	switch state.CurrentState {
	case cfg.Idle, cfg.Running:
		// Transition to charging
		if melee.IsCharging {
			state.CurrentState = cfg.StateChargingAttack
			state.StateTimer = 0
		} else if boomerangAction.Pressed && physics.OnGround != nil && player.ActiveBoomerang == nil {
			// Start Charging Boomerang
			state.CurrentState = cfg.StateChargingBoomerang
			player.BoomerangChargeTime = 0
			state.StateTimer = 0
		} else if crouchAction.Pressed && physics.OnGround != nil {
			state.CurrentState = cfg.Crouch
			state.StateTimer = 0
		} else {
			transitionToMovementState(player, physics, state)
		}

	case cfg.StateChargingAttack:
		// Still charging - increment and continue
		if melee.IsCharging {
			melee.ChargeTime++
			break
		}
		// Released but not attacking (interrupted)
		if !melee.IsAttacking {
			transitionToMovementState(player, physics, state)
			break
		}
		// Execute attack based on combo step
		if melee.ComboStep == 0 {
			melee.ComboStep = 1
			state.CurrentState = cfg.StateAttackingPunch
		} else {
			melee.ComboStep = 0
			state.CurrentState = cfg.StateAttackingKick
		}
		state.StateTimer = 0

	case cfg.StateChargingBoomerang:
		// Still charging
		if boomerangAction.Pressed {
			if player.BoomerangChargeTime < cfg.Boomerang.MaxChargeTime {
				player.BoomerangChargeTime++
			}
			physics.SpeedX = 0
			break
		}
		// Released - throw!
		state.CurrentState = cfg.Throw
		state.StateTimer = 0
		factory.CreateBoomerang(ecs, playerEntry, float64(player.BoomerangChargeTime))

	case cfg.Throw:
		// Stop movement while throwing
		physics.SpeedX = 0

		// Wait for animation to finish
		if animationLooped(animData) {
			transitionToMovementState(player, physics, state)
		}

	case cfg.StateAttackingPunch, cfg.StateAttackingKick:
		// Transition back to movement after attack animation finishes
		if animationLooped(animData) {
			melee.IsAttacking = false
			melee.HasSpawnedHitbox = false
			transitionToMovementState(player, physics, state)
		}

	case cfg.StateAttackingJump:
		// Transition back to jump after attack animation finishes
		if animationLooped(animData) {
			melee.IsAttacking = false
			melee.HasSpawnedHitbox = false
			state.CurrentState = cfg.Jump
			state.StateTimer = 0
		}

	case cfg.Hit, cfg.Stunned, cfg.Knockback:
		// Transition back to movement after hitstun/knockback duration
		if state.StateTimer > cfg.Player.InvulnFrames {
			transitionToMovementState(player, physics, state)
		}

	case cfg.Crouch:
		// Transition back to movement when down key is released
		if !crouchAction.Pressed {
			transitionToMovementState(player, physics, state)
		}

	case cfg.Jump:
		// Transition to idle/running when landing on the ground
		if physics.OnGround != nil {
			transitionToMovementState(player, physics, state)
		} else if physics.WallSliding != nil {
			state.CurrentState = cfg.WallSlide
			state.StateTimer = 0
		}

	default:
		// Default to movement state for any unhandled cases
		transitionToMovementState(player, physics, state)
	}

	updatePlayerAnimation(state, animData)
}

// Helper functions for state management
func isInLockedState(state cfg.StateID) bool {
	return state == cfg.Hit || state == cfg.Stunned || state == cfg.Knockback || state == cfg.StateChargingBoomerang || state == cfg.Throw
}

func isInAttackState(state cfg.StateID) bool {
	return state == cfg.StateAttackingPunch || state == cfg.StateAttackingKick || state == cfg.StateAttackingJump
}

func animationLooped(animData *components.AnimationData) bool {
	return animData != nil && animData.CurrentAnimation != nil && animData.CurrentAnimation.Looped
}

func updatePlayerAnimation(state *components.StateData, animData *components.AnimationData) {
	if animData == nil {
		return
	}

	var anim cfg.StateID
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

	animData.SetAnimation(anim)

	if animData.CurrentAnimation != nil {
		animData.CurrentAnimation.Update()
	}
}

func transitionToMovementState(player *components.PlayerData, physics *components.PhysicsData, state *components.StateData) {
	if physics.WallSliding != nil {
		state.CurrentState = cfg.WallSlide
	} else if physics.OnGround == nil {
		state.CurrentState = cfg.Jump
	} else if physics.SpeedX != 0 {
		state.CurrentState = cfg.Running
	} else {
		state.CurrentState = cfg.Idle
	}
	state.StateTimer = 0
	player.ComboCounter = 0
}
