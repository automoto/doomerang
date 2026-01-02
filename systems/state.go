package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateStates(ecs *ecs.ECS) {
	// Player state
	components.Player.Each(ecs.World, func(e *donburi.Entry) {
		player := components.Player.Get(e)
		physics := components.Physics.Get(e)
		state := components.State.Get(e)
		updatePlayerStateTags(e, player, physics, state)
	})

	// Enemy state
	components.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		enemy := components.Enemy.Get(e)
		physics := components.Physics.Get(e)
		state := components.State.Get(e)
		updateEnemyStateTags(e, enemy, physics, state)
	})
}

func updatePlayerStateTags(e *donburi.Entry, player *components.PlayerData, physics *components.PhysicsData, state *components.StateData) {
	if state.CurrentState == state.PreviousState {
		return
	}

	// Remove all state tags
	removeAllStateTags(e)

	// Add the current state tag
	switch state.CurrentState {
	case cfg.Idle:
		donburi.Add(e, components.Idle, &components.IdleState{})
	case cfg.Running:
		donburi.Add(e, components.Running, &components.RunningState{})
	case cfg.Jump:
		donburi.Add(e, components.Jumping, &components.JumpingState{})
	// Falling case removed as there is no explicit Falling state in StateID yet, handled by Jump or Logic
	case cfg.WallSlide:
		donburi.Add(e, components.WallSliding, &components.WallSlidingState{})
	case cfg.StateAttackingPunch, cfg.StateAttackingKick, cfg.StateAttackingJump:
		donburi.Add(e, components.Attacking, &components.AttackingState{})
	case cfg.Crouch:
		donburi.Add(e, components.Crouching, &components.CrouchingState{})
	case cfg.Stunned, cfg.Knockback, cfg.Hit:
		donburi.Add(e, components.Stunned, &components.StunnedState{})
	}

	state.PreviousState = state.CurrentState
}

func updateEnemyStateTags(e *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData) {
	if state.CurrentState == state.PreviousState {
		return
	}

	// Remove all state tags
	removeAllStateTags(e)

	// Add the current state tag
	switch state.CurrentState {
	case cfg.Idle, cfg.StatePatrol: // Map Patrol to Idle tag for now, or maybe Running depending on logic
		donburi.Add(e, components.Idle, &components.IdleState{})
	case cfg.Running, cfg.StateChase:
		donburi.Add(e, components.Running, &components.RunningState{})
	case cfg.StateAttackingPunch:
		donburi.Add(e, components.Attacking, &components.AttackingState{})
	case cfg.Stunned, cfg.Hit:
		donburi.Add(e, components.Stunned, &components.StunnedState{})
	}

	state.PreviousState = state.CurrentState
}

func removeAllStateTags(e *donburi.Entry) {
	donburi.Remove[components.IdleState](e, components.Idle)
	donburi.Remove[components.RunningState](e, components.Running)
	donburi.Remove[components.JumpingState](e, components.Jumping)
	donburi.Remove[components.FallingState](e, components.Falling)
	donburi.Remove[components.WallSlidingState](e, components.WallSliding)
	donburi.Remove[components.AttackingState](e, components.Attacking)
	donburi.Remove[components.CrouchingState](e, components.Crouching)
	donburi.Remove[components.StunnedState](e, components.Stunned)
}
