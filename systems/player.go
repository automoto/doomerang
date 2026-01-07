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

	handlePlayerInput(ecs, input, player, physics, melee, state, playerObject)
	updatePlayerState(ecs, input, playerEntry, player, physics, melee, state, animData)

	// Decrement invulnerability timer
	if player.InvulnFrames > 0 {
		player.InvulnFrames--
	}
}

func handlePlayerInput(e *ecs.ECS, input *components.InputData, player *components.PlayerData, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData, playerObject *resolv.Object) {
	// Get action states from input component
	attackAction := input.Actions[cfg.ActionAttack]
	jumpAction := input.Actions[cfg.ActionJump]
	crouchAction := input.Actions[cfg.ActionCrouch]
	moveLeftAction := input.Actions[cfg.ActionMoveLeft]
	moveRightAction := input.Actions[cfg.ActionMoveRight]

	// Process combat and jump inputs only if not in a locked state
	if !isInLockedState(state.CurrentState) {
		handleMeleeInput(attackAction, physics, melee, state, player, playerObject)

		if !isInAttackState(state.CurrentState) {
			handleJumpInput(e, jumpAction, crouchAction, physics, playerObject)
		}
	}

	// Horizontal movement (always processed)
	handleMovementInput(moveLeftAction, moveRightAction, player, physics, state)
}

func handleMeleeInput(attackAction components.ActionState, physics *components.PhysicsData, melee *components.MeleeAttackData, state *components.StateData, player *components.PlayerData, playerObject *resolv.Object) {
	// Wall kick: attack during wall slide = wall jump + kick away from wall
	if physics.WallSliding != nil {
		if !attackAction.JustPressed || isInAttackState(state.CurrentState) {
			return
		}
		performWallKick(physics, player, playerObject, state, melee)
		return
	}

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

func handleJumpInput(e *ecs.ECS, jumpAction, crouchAction components.ActionState, physics *components.PhysicsData, playerObject *resolv.Object) {
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
		PlaySFX(e, cfg.SoundJump)
		// Spawn jump dust and squash/stretch
		factory.SpawnJumpDust(e, playerObject.X+playerObject.W/2, playerObject.Y+playerObject.H)
		if playerEntry, ok := components.Player.First(e.World); ok {
			TriggerSquashStretch(playerEntry, cfg.SquashStretch.JumpScaleX, cfg.SquashStretch.JumpScaleY)
		}
		return
	}

	// Wall jump
	if physics.WallSliding == nil {
		return
	}
	physics.SpeedY = -cfg.Player.JumpSpeed
	PlaySFX(e, cfg.SoundJump)
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

	// Block movement input during slide - friction is applied in state machine
	if state.CurrentState == cfg.StateSliding {
		return
	}

	// Apply friction when crouching - gradually slow down
	if state.CurrentState == cfg.Crouch {
		friction := cfg.Player.Friction * 2.0
		if physics.SpeedX > friction {
			physics.SpeedX -= friction
		} else if physics.SpeedX < -friction {
			physics.SpeedX += friction
		} else {
			physics.SpeedX = 0
		}
		return // No acceleration while crouching
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

	// Get player object for hitbox modifications
	playerObject := components.Object.Get(playerEntry).Object

	// Main state machine logic
	switch state.CurrentState {
	case cfg.Idle, cfg.Running:
		// Transition to charging
		if melee.IsCharging {
			state.CurrentState = cfg.StateChargingAttack
			state.StateTimer = 0
		} else if boomerangAction.Pressed && player.ActiveBoomerang == nil {
			// Start Charging Boomerang (allowed in air too)
			state.CurrentState = cfg.StateChargingBoomerang
			player.BoomerangChargeTime = 0
			state.StateTimer = 0
		} else if crouchAction.JustPressed && physics.OnGround != nil {
			// Slide if moving fast enough, otherwise crouch
			if absFloat(physics.SpeedX) >= cfg.Player.SlideSpeedThreshold {
				PlaySFX(ecs, cfg.SoundSlide)
				enterSlideState(state, playerObject)
				factory.SpawnSlideDust(ecs, playerObject.X+playerObject.W/2, playerObject.Y+playerObject.H)
			} else {
				state.CurrentState = cfg.Crouch
				state.StateTimer = 0
			}
		} else if crouchAction.Pressed && physics.OnGround != nil && state.CurrentState != cfg.Running {
			state.CurrentState = cfg.Crouch
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
			// Spawn charge VFX after holding for a bit (not on quick throws)
			if player.BoomerangChargeTime == 15 && player.ChargeVFX == nil {
				player.ChargeVFX = factory.SpawnChargeVFX(ecs, playerObject.X+playerObject.W/2, playerObject.Y+playerObject.H)
				PlaySFX(ecs, cfg.SoundBoomerangCharge)
			}
			// Update charge VFX position to follow player's feet
			if player.ChargeVFX != nil {
				factory.UpdateChargeVFXPosition(player.ChargeVFX, playerObject.X+playerObject.W/2, playerObject.Y+playerObject.H)
			}
			// Apply friction instead of instant stop for smoother feel
			applyThrowFriction(physics)
			break
		}
		// Released - throw!
		// Destroy charge VFX if it exists
		if player.ChargeVFX != nil {
			factory.DestroyChargeVFX(ecs, player.ChargeVFX)
			player.ChargeVFX = nil
		}
		state.CurrentState = cfg.Throw
		state.StateTimer = 0
		PlaySFX(ecs, cfg.SoundBoomerangThrow)
		factory.CreateBoomerang(ecs, playerEntry, float64(player.BoomerangChargeTime))

	case cfg.Throw:
		// Apply friction instead of instant stop for smoother feel
		applyThrowFriction(physics)

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

	case cfg.StateSliding:
		// Friction is handled by physics system with cfg.Player.SlideFriction
		speed := absFloat(physics.SpeedX)
		canStandUp := !crouchAction.Pressed && state.StateTimer > cfg.Player.SlideRecoveryFrames
		slideStopped := speed < cfg.Player.SlideMinSpeed

		if !slideStopped && !canStandUp {
			break
		}
		restoreSlideHitbox(playerObject)
		if slideStopped && crouchAction.Pressed {
			state.CurrentState = cfg.Crouch
		} else {
			transitionToMovementState(player, physics, state)
		}

	case cfg.Jump:
		// Allow boomerang throw while jumping
		if boomerangAction.Pressed && player.ActiveBoomerang == nil {
			state.CurrentState = cfg.StateChargingBoomerang
			player.BoomerangChargeTime = 0
			state.StateTimer = 0
			break
		}
		// Transition to idle/running when landing on the ground
		if physics.OnGround != nil {
			PlaySFX(ecs, cfg.SoundLand)
			// Spawn landing dust and squash/stretch
			factory.SpawnLandDust(ecs, playerObject.X+playerObject.W/2, playerObject.Y+playerObject.H)
			TriggerSquashStretch(playerEntry, cfg.SquashStretch.LandScaleX, cfg.SquashStretch.LandScaleY)
			transitionToMovementState(player, physics, state)
		} else if physics.WallSliding != nil {
			state.CurrentState = cfg.WallSlide
			state.StateTimer = 0
			PlaySFX(ecs, cfg.SoundWallAttach)
		}

	case cfg.WallSlide:
		// Transition when no longer wall sliding
		if physics.WallSliding == nil {
			transitionToMovementState(player, physics, state)
		}

	default:
		// Default to movement state for any unhandled cases
		transitionToMovementState(player, physics, state)
	}

	updatePlayerAnimation(state, animData)
}

// Helper functions for state management
func isInLockedState(state cfg.StateID) bool {
	return state == cfg.Hit || state == cfg.Stunned || state == cfg.Knockback || state == cfg.StateChargingBoomerang || state == cfg.Throw || state == cfg.StateSliding
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

// applyThrowFriction applies gradual friction during boomerang throw instead of instant stop
func applyThrowFriction(physics *components.PhysicsData) {
	// Use moderate friction for gradual slowdown during throw
	friction := cfg.Player.Friction * 1.0
	if physics.SpeedX > friction {
		physics.SpeedX -= friction
	} else if physics.SpeedX < -friction {
		physics.SpeedX += friction
	} else {
		physics.SpeedX = 0
	}
}

// restoreSlideHitbox restores the player hitbox to normal height after sliding
func restoreSlideHitbox(playerObject *resolv.Object) {
	normalHeight := float64(cfg.Player.CollisionHeight)
	if playerObject.H >= normalHeight {
		return
	}
	heightDiff := normalHeight - playerObject.H
	playerObject.H = normalHeight
	playerObject.Y -= heightDiff
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func performWallKick(physics *components.PhysicsData, player *components.PlayerData, playerObject *resolv.Object, state *components.StateData, melee *components.MeleeAttackData) {
	wallCenterX := physics.WallSliding.X + physics.WallSliding.W/2
	playerCenterX := playerObject.X + playerObject.W/2

	physics.SpeedY = -cfg.Player.JumpSpeed
	if wallCenterX > playerCenterX {
		physics.SpeedX = -physics.MaxSpeed
		player.Direction.X = cfg.DirectionLeft
	} else {
		physics.SpeedX = physics.MaxSpeed
		player.Direction.X = cfg.DirectionRight
	}
	physics.WallSliding = nil
	state.CurrentState = cfg.StateAttackingJump
	state.StateTimer = 0
	melee.IsAttacking = true
}

func enterSlideState(state *components.StateData, playerObject *resolv.Object) {
	state.CurrentState = cfg.StateSliding
	state.StateTimer = 0
	heightDiff := playerObject.H - cfg.Player.SlideHitboxHeight
	playerObject.H = cfg.Player.SlideHitboxHeight
	playerObject.Y += heightDiff
}
