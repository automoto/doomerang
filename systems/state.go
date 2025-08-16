package systems

import (
	"github.com/automoto/doomerang/components"
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
	// Remove all state tags
	removeAllStateTags(e)

	// Add the current state tag
	switch state.CurrentState {
	case "idle":
		donburi.Add(e, components.Idle, &components.IdleState{})
	case "running":
		donburi.Add(e, components.Running, &components.RunningState{})
	case "jumping":
		donburi.Add(e, components.Jumping, &components.JumpingState{})
	case "falling":
		donburi.Add(e, components.Falling, &components.FallingState{})
	case "wallsliding":
		donburi.Add(e, components.WallSliding, &components.WallSlidingState{})
	case "attacking":
		donburi.Add(e, components.Attacking, &components.AttackingState{})
	case "crouching":
		donburi.Add(e, components.Crouching, &components.CrouchingState{})
	case "stunned":
		donburi.Add(e, components.Stunned, &components.StunnedState{})
	}
}

func updateEnemyStateTags(e *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData) {
	// Remove all state tags
	removeAllStateTags(e)

	// Add the current state tag
	switch state.CurrentState {
	case "idle":
		donburi.Add(e, components.Idle, &components.IdleState{})
	case "running":
		donburi.Add(e, components.Running, &components.RunningState{})
	case "attacking":
		donburi.Add(e, components.Attacking, &components.AttackingState{})
	case "stunned":
		donburi.Add(e, components.Stunned, &components.StunnedState{})
	}
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
