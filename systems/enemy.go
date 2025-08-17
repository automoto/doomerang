package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// AI state constants - should match factory
const (
	enemyStatePatrol = "patrol"
	enemyStateChase  = "chase"
	enemyStateAttack = "attack"
	enemyStateHit    = "hit"
)

func UpdateEnemies(ecs *ecs.ECS) {
	// Get player position for AI decisions
	playerEntry, _ := components.Player.First(ecs.World)
	var playerObject *resolv.Object
	if playerEntry != nil {
		playerObject = cfg.GetObject(playerEntry)
	}

	tags.Enemy.Each(ecs.World, func(e *donburi.Entry) {
		// Skip if enemy is in death sequence
		if e.HasComponent(components.Death) {
			if anim := components.Animation.Get(e); anim != nil && anim.CurrentAnimation != nil {
				anim.CurrentAnimation.Update()
			}
			return
		}

		// Skip if invulnerable
		enemy := components.Enemy.Get(e)
		if enemy.InvulnFrames > 0 {
			enemy.InvulnFrames--
		}

		// Update AI behavior
		updateEnemyAI(ecs, e, playerObject)

		// Update animation state
		updateEnemyAnimation(enemy, components.Physics.Get(e), components.State.Get(e), components.Animation.Get(e))
	})
}

func updateEnemyAI(ecs *ecs.ECS, enemyEntry *donburi.Entry, playerObject *resolv.Object) {
	enemy := components.Enemy.Get(enemyEntry)
	physics := components.Physics.Get(enemyEntry)
	enemyObject := cfg.GetObject(enemyEntry)
	state := components.State.Get(enemyEntry)
	state.StateTimer++

	// Update attack cooldown
	if enemy.AttackCooldown > 0 {
		enemy.AttackCooldown--
	}

	// No AI if no player
	if playerObject == nil {
		return
	}

	// Calculate distance to player
	distanceToPlayer := math.Abs(playerObject.X - enemyObject.X)

	// State machine
	switch state.CurrentState {
	case enemyStatePatrol:
		handlePatrolState(enemy, physics, state, enemyObject, playerObject, distanceToPlayer)
	case enemyStateChase:
		handleChaseState(ecs, enemyEntry, playerObject, distanceToPlayer)
	case enemyStateAttack:
		handleAttackState(ecs, enemyEntry)
	case enemyStateHit:
		// Stunned for a short period
		if state.StateTimer > 15 { // 15 frames of hitstun
			state.CurrentState = enemyStateChase
			state.StateTimer = 0
		}
	}
}

func handlePatrolState(enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	// Check if should start chasing
	if distanceToPlayer <= enemy.ChaseRange {
		state.CurrentState = enemyStateChase
		state.StateTimer = 0
		return
	}

	// Patrol behavior - move back and forth
	if enemy.Direction.X > 0 {
		physics.SpeedX += enemy.PatrolSpeed * 1.1
		// Turn around if hit right boundary
		if enemyObject.X >= enemy.PatrolRight {
			enemy.Direction.X = -1
		}
	} else {
		physics.SpeedX -= enemy.PatrolSpeed * 1.1
		// Turn around if hit left boundary
		if enemyObject.X <= enemy.PatrolLeft {
			enemy.Direction.X = 1
		}
	}
}

func handleChaseState(ecs *ecs.ECS, enemyEntry *donburi.Entry, playerObject *resolv.Object, distanceToPlayer float64) {
	enemy := components.Enemy.Get(enemyEntry)
	physics := components.Physics.Get(enemyEntry)
	state := components.State.Get(enemyEntry)
	enemyObject := cfg.GetObject(enemyEntry)
	// Check if should attack
	if distanceToPlayer <= enemy.AttackRange && enemy.AttackCooldown == 0 {
		state.CurrentState = enemyStateAttack
		state.StateTimer = 0
		return
	}

	// Check if should stop chasing (player too far)
	if distanceToPlayer > enemy.ChaseRange*1.5 { // Hysteresis to prevent flapping
		state.CurrentState = enemyStatePatrol
		state.StateTimer = 0
		return
	}

	// Face the player
	if playerObject.X > enemyObject.X {
		enemy.Direction.X = 1
	} else {
		enemy.Direction.X = -1
	}

	// Move towards player if not within stopping distance
	if distanceToPlayer > enemy.StoppingDistance {
		if playerObject.X > enemyObject.X {
			physics.SpeedX = enemy.ChaseSpeed
		} else {
			physics.SpeedX = -enemy.ChaseSpeed
		}
	}
}

func handleAttackState(ecs *ecs.ECS, enemyEntry *donburi.Entry) {
	enemy := components.Enemy.Get(enemyEntry)
	state := components.State.Get(enemyEntry)
	enemyObject := cfg.GetObject(enemyEntry)
	// Create a hitbox on the first frame of the attack
	if state.StateTimer == 1 {
		CreateHitbox(ecs, enemyEntry, enemyObject, "punch", false)
	}

	// Attack animation duration (simplified - using timer)
	attackDuration := 30 // 30 frames for attack

	if state.StateTimer >= attackDuration {
		// Attack finished
		state.CurrentState = enemyStateChase
		state.StateTimer = 0
		enemy.AttackCooldown = 60 // 1 second cooldown
		return
	}

	// Don't apply movement input during attack - let friction naturally slow down
}

func updateEnemyAnimation(enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, animData *components.AnimationData) {
	// Simple animation state based on movement and AI state
	var targetState string

	switch state.CurrentState {
	case enemyStateAttack:
		targetState = cfg.Punch01 // Use punch animation for attacks
	case enemyStateHit:
		targetState = cfg.Hit
	default:
		if physics.OnGround == nil {
			targetState = cfg.Jump
		} else if physics.SpeedX != 0 {
			targetState = cfg.Running
		} else {
			targetState = cfg.Idle
		}
	}

	// Update animation if changed
	if animData.CurrentAnimation != animData.Animations[targetState] {
		animData.SetAnimation(targetState)
	}

	if animData.CurrentAnimation != nil {
		animData.CurrentAnimation.Update()
	}
}
