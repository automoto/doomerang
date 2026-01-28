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
	attackAction := GetAction(input, cfg.ActionAttack)
	jumpAction := GetAction(input, cfg.ActionJump)
	crouchAction := GetAction(input, cfg.ActionCrouch)
	moveLeftAction := GetAction(input, cfg.ActionMoveLeft)
	moveRightAction := GetAction(input, cfg.ActionMoveRight)

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

	// Block movement input during slide and crouch - handled separately in state machine
	if state.CurrentState == cfg.StateSliding || state.CurrentState == cfg.Crouch {
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
	boomerangAction := GetAction(input, cfg.ActionBoomerang)
	crouchAction := GetAction(input, cfg.ActionCrouch)

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
				enterCrouchState(state, playerObject)
			}
		} else if crouchAction.Pressed && physics.OnGround != nil && state.CurrentState != cfg.Running {
			enterCrouchState(state, playerObject)
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
		aimX, aimY := calculateBoomerangAim(input, player.Direction.X)
		factory.CreateBoomerang(ecs, playerEntry, float64(player.BoomerangChargeTime), aimX, aimY)

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
		// Allow slow crouch-walking
		applyCrouchMovement(input, player, physics)

		// Transition back to movement when down key is released
		if !crouchAction.Pressed {
			// Try to stand up (may push player horizontally if partially blocked)
			if !tryStandUp(playerObject, player.Direction.X) {
				break // Completely blocked - stay crouched
			}
			transitionToMovementState(player, physics, state)
		}

	case cfg.StateSliding:
		// Friction is handled by physics system with cfg.Player.SlideFriction
		speed := absFloat(physics.SpeedX)
		wantsToStandUp := !crouchAction.Pressed && state.StateTimer > cfg.Player.SlideRecoveryFrames
		slideStopped := speed < cfg.Player.SlideMinSpeed

		if !slideStopped && !wantsToStandUp {
			break
		}

		// Try to stand up (may push player horizontally if partially blocked)
		if !tryStandUp(playerObject, player.Direction.X) {
			// Completely blocked - transition to crouch (keeps reduced hitbox)
			physics.SpeedX = 0
			enterCrouchState(state, playerObject)
			break
		}

		if slideStopped && crouchAction.Pressed {
			enterCrouchState(state, playerObject)
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

// calculateBoomerangAim returns the aim direction vector based on input.
// Returns (aimX, aimY) where the vector represents the throw direction.
func calculateBoomerangAim(input *components.InputData, facingX float64) (aimX, aimY float64) {
	upAction := GetAction(input, cfg.ActionMoveUp)
	downAction := GetAction(input, cfg.ActionCrouch) // Down/Crouch key for aiming down
	leftAction := GetAction(input, cfg.ActionMoveLeft)
	rightAction := GetAction(input, cfg.ActionMoveRight)

	horizontal := leftAction.Pressed || rightAction.Pressed

	if upAction.Pressed && !downAction.Pressed {
		if horizontal {
			return facingX, -1.0 // Diagonal up
		}
		return 0, -1.0 // Straight up
	}
	if downAction.Pressed && !upAction.Pressed {
		if horizontal {
			return facingX, 1.0 // Diagonal down
		}
		return 0, 1.0 // Straight down
	}
	return facingX, 0 // Forward (default)
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

func applyThrowFriction(physics *components.PhysicsData) {
	applyFriction(physics, cfg.Player.Friction)
}

// tryStandUp attempts to restore full height hitbox, pushing player horizontally if needed.
func tryStandUp(playerObject *resolv.Object, facingX float64) bool {
	normalHeight := float64(cfg.Player.CollisionHeight)
	if playerObject.H >= normalHeight {
		return true
	}

	heightDiff := normalHeight - playerObject.H

	if playerObject.Check(0, -heightDiff, "solid") == nil {
		playerObject.H = normalHeight
		playerObject.Y -= heightDiff
		return true
	}

	// Blocked above - try pushing horizontally
	const pushDistance = 12.0
	for _, dir := range []float64{facingX, -facingX} {
		for offset := 1.0; offset <= pushDistance; offset++ {
			if playerObject.Check(offset*dir, -heightDiff, "solid") == nil {
				playerObject.X += offset * dir
				playerObject.H = normalHeight
				playerObject.Y -= heightDiff
				return true
			}
		}
	}

	return false
}

func absFloat(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func applyCrouchMovement(input *components.InputData, player *components.PlayerData, physics *components.PhysicsData) {
	left := GetAction(input, cfg.ActionMoveLeft).Pressed
	right := GetAction(input, cfg.ActionMoveRight).Pressed

	switch {
	case right:
		physics.SpeedX = cfg.Player.CrouchWalkSpeed
		player.Direction.X = cfg.DirectionRight
	case left:
		physics.SpeedX = -cfg.Player.CrouchWalkSpeed
		player.Direction.X = cfg.DirectionLeft
	default:
		applyFriction(physics, cfg.Player.Friction)
	}
}

func applyFriction(physics *components.PhysicsData, friction float64) {
	switch {
	case physics.SpeedX > friction:
		physics.SpeedX -= friction
	case physics.SpeedX < -friction:
		physics.SpeedX += friction
	default:
		physics.SpeedX = 0
	}
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
	reduceHitboxForCrouch(playerObject)
}

func enterCrouchState(state *components.StateData, playerObject *resolv.Object) {
	state.CurrentState = cfg.Crouch
	state.StateTimer = 0
	reduceHitboxForCrouch(playerObject)
}

func reduceHitboxForCrouch(playerObject *resolv.Object) {
	targetHeight := cfg.Player.SlideHitboxHeight
	if playerObject.H <= targetHeight {
		return
	}
	heightDiff := playerObject.H - targetHeight
	playerObject.H = targetHeight
	playerObject.Y += heightDiff
}
