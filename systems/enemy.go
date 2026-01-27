package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems/factory"
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

	// Use ranged AI for ranged enemies
	if enemy.TypeConfig != nil && enemy.TypeConfig.IsRanged {
		updateRangedEnemyAI(ecs, enemyEntry, enemy, physics, state, enemyObject.Object, playerObject, distanceToPlayer)
		return
	}

	// For melee enemies, skip chase/attack if player is on a different vertical level
	verticalDistance := math.Abs(playerObject.Y - enemyObject.Y)
	if enemy.TypeConfig != nil && verticalDistance > enemy.TypeConfig.MaxVerticalChase && enemy.TypeConfig.MaxVerticalChase > 0 {
		if state.CurrentState == cfg.StateChase || state.CurrentState == cfg.StateAttackingPunch {
			state.CurrentState = cfg.StatePatrol
			state.StateTimer = 0
		}
		return
	}

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
		if state.StateTimer > enemy.TypeConfig.HitstunDuration {
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
	// Hitbox creation is handled by combat_hitbox.go

	// Attack animation duration (simplified - using timer)
	if state.StateTimer >= enemy.TypeConfig.AttackDuration {
		// Attack finished
		state.CurrentState = cfg.StateChase
		state.StateTimer = 0
		enemy.AttackCooldown = enemy.TypeConfig.AttackCooldown
		return
	}

	// Don't apply movement input during attack - let friction naturally slow down
}

func updateRangedEnemyAI(ecs *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	switch state.CurrentState {
	case cfg.StatePatrol, cfg.Idle:
		updateRangedPatrolState(ecs, enemyEntry, enemy, physics, state, enemyObject, playerObject, distanceToPlayer)

	case cfg.StateApproachEdge:
		handleApproachEdgeState(ecs, enemyEntry, enemy, physics, state, enemyObject, playerObject, distanceToPlayer)

	case cfg.Throw:
		handleThrowState(ecs, enemyEntry, enemy, state, enemyObject, playerObject)

	case cfg.Hit:
		if state.StateTimer > enemy.TypeConfig.HitstunDuration {
			state.CurrentState = cfg.StatePatrol
			state.StateTimer = 0
		}
		physics.SpeedX = 0
	}
}

func updateRangedPatrolState(ecs *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	if distanceToPlayer > enemy.TypeConfig.ThrowRange || enemy.AttackCooldown > 0 {
		if enemy.PatrolPathName != "" {
			handleCustomPatrol(ecs, enemyEntry, enemy, physics, state, enemyObject)
		} else {
			handleDefaultPatrol(enemy, physics, enemyObject)
		}
		return
	}

	if playerObject.X > enemyObject.X {
		enemy.Direction.X = 1
	} else {
		enemy.Direction.X = -1
	}

	verticalDiff := playerObject.Y - enemyObject.Y
	if verticalDiff > enemy.TypeConfig.MinVerticalToThrow && enemy.TypeConfig.MinVerticalToThrow > 0 {
		state.CurrentState = cfg.StateApproachEdge
		state.StateTimer = 0
		return
	}

	state.CurrentState = cfg.Throw
	state.StateTimer = 0
	physics.SpeedX = 0
}

// handleThrowState handles the throwing animation and knife creation
func handleThrowState(ecs *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, state *components.StateData, enemyObject, playerObject *resolv.Object) {
	// Throw knife at windup frame
	if state.StateTimer == enemy.TypeConfig.ThrowWindupTime {
		// Target player's current position
		targetX := playerObject.X + playerObject.W/2
		targetY := playerObject.Y + playerObject.H/2
		factory.CreateKnife(ecs, enemyEntry, targetX, targetY)
		PlaySFX(ecs, cfg.SoundBoomerangThrow) // Reuse throw sound for now
	}

	// Animation complete (windup + throw animation)
	if state.StateTimer >= enemy.TypeConfig.ThrowWindupTime+15 {
		state.CurrentState = cfg.StatePatrol
		state.StateTimer = 0
		enemy.AttackCooldown = enemy.TypeConfig.ThrowCooldown
	}
}

func handleApproachEdgeState(ecs *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	if playerObject.X > enemyObject.X {
		enemy.Direction.X = 1
	} else {
		enemy.Direction.X = -1
	}

	verticalDiff := playerObject.Y - enemyObject.Y
	if distanceToPlayer > enemy.TypeConfig.ThrowRange || verticalDiff <= enemy.TypeConfig.MinVerticalToThrow {
		state.CurrentState = cfg.StatePatrol
		state.StateTimer = 0
		return
	}

	if !isAtPlatformEdge(enemyObject, enemy.Direction.X) {
		physics.SpeedX = enemy.TypeConfig.EdgeApproachSpeed * enemy.Direction.X
		return
	}

	physics.SpeedX = 0
	if distanceToPlayer <= enemy.TypeConfig.EdgeThrowDistance && enemy.AttackCooldown == 0 {
		state.CurrentState = cfg.Throw
		state.StateTimer = 0
	}
}

func isAtPlatformEdge(obj *resolv.Object, direction float64) bool {
	return obj.Check(8.0*direction, obj.H+4.0, "solid", "platform") == nil
}

func updateEnemyAnimation(enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, animData *components.AnimationData) {
	// Simple animation state based on movement and AI state
	var targetState cfg.StateID

	switch state.CurrentState {
	case cfg.StateAttackingPunch:
		targetState = cfg.Punch01 // Use punch animation for attacks
	case cfg.Throw:
		targetState = cfg.Throw // Use throw animation for ranged attacks
	case cfg.Hit:
		targetState = cfg.Hit
	case cfg.StateApproachEdge:
		targetState = cfg.Walk // Use walk animation when approaching edge
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
	animData.SetAnimation(targetState)

	if animData.CurrentAnimation != nil {
		animData.CurrentAnimation.Update()
	}
}
