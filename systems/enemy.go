package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
	math2 "github.com/yohamta/donburi/features/math"
)

func UpdateEnemies(ecs *ecs.ECS) {
	// Get player position for AI decisions
	playerEntry, _ := components.Player.First(ecs.World)
	var playerObject *resolv.Object
	if playerEntry != nil {
		playerObject = components.Object.Get(playerEntry).Object
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

		// Update health bar timer
		if e.HasComponent(components.HealthBar) {
			healthBar := components.HealthBar.Get(e)
			healthBar.TimeToLive--
			if healthBar.TimeToLive <= 0 {
				donburi.Remove[components.HealthBarData](e, components.HealthBar)
			}
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
	enemyObject := components.Object.Get(enemyEntry)
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
	case cfg.StatePatrol:
		handlePatrolState(ecs, enemyEntry, enemy, physics, state, enemyObject.Object, playerObject, distanceToPlayer)
	case cfg.StateChase:
		handleChaseState(ecs, enemyEntry, playerObject, distanceToPlayer)
	case cfg.StateAttackingPunch:
		handleAttackState(ecs, enemyEntry)
	case cfg.Hit:
		// Stunned for a short period
		typeName := "Guard"
		enemyType, ok := cfg.Enemy.Types[typeName]
		if !ok {
			enemyType = cfg.Enemy.Types["Guard"]
		}

		if state.StateTimer > enemyType.HitstunDuration {
			state.CurrentState = cfg.StateChase
			state.StateTimer = 0
		}
	}
}

func handlePatrolState(ecs *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	// Check if should start chasing
	if distanceToPlayer <= enemy.ChaseRange {
		state.CurrentState = cfg.StateChase
		state.StateTimer = 0
		return
	}

	// If enemy has a custom patrol path, use it
	if enemy.PatrolPathName != "" {
		handleCustomPatrol(ecs, enemyEntry, enemy, physics, state, enemyObject)
	} else {
		// Default patrol behavior - move back and forth
		if enemy.Direction.X > 0 {
			physics.SpeedX = enemy.PatrolSpeed
			// Turn around if hit right boundary
			if enemyObject.X >= enemy.PatrolRight {
				enemy.Direction.X = -1
			}
		} else {
			physics.SpeedX = -enemy.PatrolSpeed
			// Turn around if hit left boundary
			if enemyObject.X <= enemy.PatrolLeft {
				enemy.Direction.X = 1
			}
		}
	}
}

func handleCustomPatrol(ecs *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject *resolv.Object) {
	// Get the current level to access patrol paths
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		// Fallback to default patrol if no level found
		handleDefaultPatrol(enemy, physics, enemyObject)
		return
	}

	levelData := components.Level.Get(levelEntry)
	currentLevel := levelData.CurrentLevel

	// Find the patrol path by name
	patrolPath, exists := currentLevel.PatrolPaths[enemy.PatrolPathName]
	if !exists || len(patrolPath.Points) < 2 {
		// Fallback to default patrol if path not found or invalid
		handleDefaultPatrol(enemy, physics, enemyObject)
		return
	}

	// For 2-point polylines, implement back-and-forth patrol between start and end points
	startPoint := patrolPath.Points[0]
	endPoint := patrolPath.Points[1]

	// Ensure startPoint is the leftmost point to align with Direction logic
	if startPoint.X > endPoint.X {
		startPoint, endPoint = endPoint, startPoint
	}

	// Determine which direction we should be moving based on current position
	var targetPoint math2.Vec2
	if enemy.Direction.X > 0 {
		targetPoint = endPoint
	} else {
		targetPoint = startPoint
	}

	// Set the speed directly, bypassing friction for patrol
	physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X

	// If we are close to the target, or have overshot it, switch directions
	// Check distance for proximity flip
	if math.Abs(targetPoint.X-enemyObject.X) < enemy.PatrolSpeed {
		enemy.Direction.X *= -1
		return
	}

	// Check for overshoot based on direction
	if enemy.Direction.X > 0 { // Moving Right towards End
		if enemyObject.X > targetPoint.X {
			enemy.Direction.X = -1
		}
	} else { // Moving Left towards Start
		if enemyObject.X < targetPoint.X {
			enemy.Direction.X = 1
		}
	}
}

func handleDefaultPatrol(enemy *components.EnemyData, physics *components.PhysicsData, enemyObject *resolv.Object) {
	// Default patrol behavior - move back and forth
	if enemy.Direction.X > 0 {
		physics.SpeedX = enemy.PatrolSpeed
		// Turn around if hit right boundary
		if enemyObject.X >= enemy.PatrolRight {
			enemy.Direction.X = -1
		}
	} else {
		physics.SpeedX = -enemy.PatrolSpeed
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
	enemyObject := components.Object.Get(enemyEntry)
	// Check if should attack
	if distanceToPlayer <= enemy.AttackRange && enemy.AttackCooldown == 0 {
		state.CurrentState = cfg.StateAttackingPunch
		state.StateTimer = 0
		return
	}

	// Check if should stop chasing (player too far)
	if distanceToPlayer > enemy.ChaseRange*cfg.Enemy.HysteresisMultiplier { // Hysteresis to prevent flapping
		state.CurrentState = cfg.StatePatrol
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
	enemyObject := components.Object.Get(enemyEntry)
	// Create a hitbox on the first frame of the attack
	if state.StateTimer == 1 {
		CreateHitbox(ecs, enemyEntry, enemyObject.Object, "punch", false)
	}

	// Attack animation duration (simplified - using timer)
	typeName := "Guard"
	enemyType, ok := cfg.Enemy.Types[typeName]
	if !ok {
		enemyType = cfg.Enemy.Types["Guard"]
	}

	if state.StateTimer >= enemyType.AttackDuration {
		// Attack finished
		state.CurrentState = cfg.StateChase
		state.StateTimer = 0
		enemy.AttackCooldown = enemyType.AttackCooldown
		return
	}

	// Don't apply movement input during attack - let friction naturally slow down
}

func updateEnemyAnimation(enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, animData *components.AnimationData) {
	// Simple animation state based on movement and AI state
	var targetState cfg.StateID

	switch state.CurrentState {
	case cfg.StateAttackingPunch:
		targetState = cfg.Punch01 // Use punch animation for attacks
	case cfg.Hit:
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
