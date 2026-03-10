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

func UpdateEnemies(e *ecs.ECS) {
	playerEntry, _ := components.Player.First(e.World)
	var playerObject *resolv.Object
	if playerEntry != nil {
		playerObject = components.Object.Get(playerEntry).Object
	}

	// Pre-collect living enemy positions for O(n) separation (avoids O(n²) nested Each).
	enemyPositions := collectEnemyPositions(e)

	tags.Enemy.Each(e.World, func(entry *donburi.Entry) {
		if entry.HasComponent(components.Death) {
			anim := components.Animation.Get(entry)
			if anim != nil && anim.CurrentAnimation != nil {
				anim.CurrentAnimation.Update()
			}
			return
		}

		enemy := components.Enemy.Get(entry)
		if enemy.InvulnFrames > 0 {
			enemy.InvulnFrames--
		}

		if entry.HasComponent(components.HealthBar) {
			healthBar := components.HealthBar.Get(entry)
			healthBar.TimeToLive--
			if healthBar.TimeToLive <= 0 {
				donburi.Remove[components.HealthBarData](entry, components.HealthBar)
			}
		}

		updateEnemyAI(e, entry, playerObject, enemyPositions)
		updateEnemyAnimation(enemy, components.Physics.Get(entry), components.State.Get(entry), components.Animation.Get(entry))
	})
}

func updateEnemyAI(e *ecs.ECS, enemyEntry *donburi.Entry, playerObject *resolv.Object, enemyPositions []float64) {
	enemy := components.Enemy.Get(enemyEntry)
	physics := components.Physics.Get(enemyEntry)
	enemyObject := components.Object.Get(enemyEntry)
	state := components.State.Get(enemyEntry)
	state.StateTimer++

	if enemy.AttackCooldown > 0 {
		enemy.AttackCooldown--
	}
	if playerObject == nil {
		return
	}

	distanceToPlayer := math.Abs(playerObject.X - enemyObject.X)

	if enemy.TypeConfig != nil && enemy.TypeConfig.IsRanged {
		updateRangedEnemyAI(e, enemyEntry, enemy, physics, state, enemyObject.Object, playerObject, distanceToPlayer)
	} else {
		updateMeleeEnemyAI(e, enemyEntry, enemy, physics, state, enemyObject.Object, playerObject, distanceToPlayer)
	}

	// Apply separation to any enemy that is actively moving (patrol or chase).
	// Keyed on SpeedX so attack/hit/throw states — where the enemy must stay put — are unaffected.
	if physics.SpeedX != 0 {
		physics.SpeedX += computeSeparationX(enemyObject.X, enemyPositions, cfg.Enemy.SeparationRadius) * cfg.Enemy.SeparationForce
	}
}

func updateMeleeEnemyAI(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	verticalDistance := math.Abs(playerObject.Y - enemyObject.Y)
	if enemy.TypeConfig != nil && enemy.TypeConfig.MaxVerticalChase > 0 && verticalDistance > enemy.TypeConfig.MaxVerticalChase {
		if state.CurrentState == cfg.StateChase || state.CurrentState == cfg.StateAttackingPunch {
			state.CurrentState = cfg.StatePatrol
			state.StateTimer = 0
		}
		return
	}

	switch state.CurrentState {
	case cfg.StatePatrol:
		handlePatrolState(e, enemyEntry, enemy, physics, state, enemyObject, playerObject, distanceToPlayer)
	case cfg.StateChase:
		handleChaseState(enemyEntry, playerObject, distanceToPlayer)
	case cfg.StateAttackingPunch:
		handleAttackState(enemyEntry)
	case cfg.Hit:
		if state.StateTimer > enemy.TypeConfig.HitstunDuration {
			state.CurrentState = cfg.StateChase
			state.StateTimer = 0
		}
	}
}

func handlePatrolState(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	if distanceToPlayer <= enemy.ChaseRange {
		state.CurrentState = cfg.StateChase
		state.StateTimer = 0
		return
	}
	dispatchPatrol(e, enemyEntry, enemy, physics, state, enemyObject)
}

// dispatchPatrol routes to the appropriate patrol handler based on whether
// a custom path is configured. Used by both melee and ranged patrol states.
func dispatchPatrol(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject *resolv.Object) {
	if enemy.PatrolPathName != "" {
		handleCustomPatrol(e, enemyEntry, enemy, physics, state, enemyObject)
	} else {
		handleDefaultPatrol(enemy, physics, enemyObject)
	}
}

func handleCustomPatrol(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject *resolv.Object) {
	levelEntry, ok := components.Level.First(e.World)
	if !ok {
		handleDefaultPatrol(enemy, physics, enemyObject)
		return
	}

	patrolPath, exists := components.Level.Get(levelEntry).CurrentLevel.PatrolPaths[enemy.PatrolPathName]
	if !exists || len(patrolPath.Points) < 2 {
		handleDefaultPatrol(enemy, physics, enemyObject)
		return
	}

	startPoint, endPoint := patrolPath.Points[0], patrolPath.Points[1]
	if startPoint.X > endPoint.X {
		startPoint, endPoint = endPoint, startPoint
	}

	var targetPoint math2.Vec2
	if enemy.Direction.X > 0 {
		targetPoint = endPoint
	} else {
		targetPoint = startPoint
	}

	physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X

	// Proximity: flip before reaching target
	if math.Abs(targetPoint.X-enemyObject.X) < enemy.PatrolSpeed {
		enemy.Direction.X *= -1
		return
	}

	// Overshoot: flip if we've passed the target
	if enemy.Direction.X > 0 && enemyObject.X > targetPoint.X {
		enemy.Direction.X = -1
	} else if enemy.Direction.X < 0 && enemyObject.X < targetPoint.X {
		enemy.Direction.X = 1
	}
}

func handleDefaultPatrol(enemy *components.EnemyData, physics *components.PhysicsData, enemyObject *resolv.Object) {
	physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X
	if enemy.Direction.X > 0 && enemyObject.X >= enemy.PatrolRight {
		enemy.Direction.X = -1
	} else if enemy.Direction.X < 0 && enemyObject.X <= enemy.PatrolLeft {
		enemy.Direction.X = 1
	}
}

func handleChaseState(enemyEntry *donburi.Entry, playerObject *resolv.Object, distanceToPlayer float64) {
	enemy := components.Enemy.Get(enemyEntry)
	physics := components.Physics.Get(enemyEntry)
	state := components.State.Get(enemyEntry)
	enemyObject := components.Object.Get(enemyEntry)

	if distanceToPlayer <= enemy.AttackRange && enemy.AttackCooldown == 0 {
		state.CurrentState = cfg.StateAttackingPunch
		state.StateTimer = 0
		return
	}
	if distanceToPlayer > enemy.ChaseRange*cfg.Enemy.HysteresisMultiplier {
		state.CurrentState = cfg.StatePatrol
		state.StateTimer = 0
		return
	}

	enemy.Direction.X = math.Copysign(1, playerObject.X-enemyObject.X)
	if distanceToPlayer > enemy.StoppingDistance {
		physics.SpeedX = math.Copysign(enemy.ChaseSpeed, playerObject.X-enemyObject.X)
	}
}

func handleAttackState(enemyEntry *donburi.Entry) {
	enemy := components.Enemy.Get(enemyEntry)
	state := components.State.Get(enemyEntry)
	if state.StateTimer >= enemy.TypeConfig.AttackDuration {
		state.CurrentState = cfg.StateChase
		state.StateTimer = 0
		enemy.AttackCooldown = enemy.TypeConfig.AttackCooldown
	}
	// Movement during attack is intentionally omitted — friction handles deceleration.
}

func updateRangedEnemyAI(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	switch state.CurrentState {
	case cfg.StatePatrol, cfg.Idle:
		updateRangedPatrolState(e, enemyEntry, enemy, physics, state, enemyObject, playerObject, distanceToPlayer)
	case cfg.StateApproachEdge:
		handleApproachEdgeState(e, enemyEntry, enemy, physics, state, enemyObject, playerObject, distanceToPlayer)
	case cfg.Throw:
		handleThrowState(e, enemyEntry, enemy, state, enemyObject, playerObject)
	case cfg.Hit:
		if state.StateTimer > enemy.TypeConfig.HitstunDuration {
			state.CurrentState = cfg.StatePatrol
			state.StateTimer = 0
		}
		physics.SpeedX = 0
	}
}

func updateRangedPatrolState(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	if distanceToPlayer > enemy.TypeConfig.ThrowRange || enemy.AttackCooldown > 0 {
		dispatchPatrol(e, enemyEntry, enemy, physics, state, enemyObject)
		return
	}

	enemy.Direction.X = math.Copysign(1, playerObject.X-enemyObject.X)

	verticalDiff := playerObject.Y - enemyObject.Y
	if enemy.TypeConfig.MinVerticalToThrow > 0 && verticalDiff > enemy.TypeConfig.MinVerticalToThrow {
		state.CurrentState = cfg.StateApproachEdge
		state.StateTimer = 0
		return
	}

	state.CurrentState = cfg.Throw
	state.StateTimer = 0
	physics.SpeedX = 0
}

func handleThrowState(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, state *components.StateData, enemyObject, playerObject *resolv.Object) {
	if state.StateTimer == enemy.TypeConfig.ThrowWindupTime {
		targetX := playerObject.X + playerObject.W/2
		targetY := playerObject.Y + playerObject.H/2
		factory.CreateKnife(e, enemyEntry, targetX, targetY)
		PlaySFX(e, cfg.SoundBoomerangThrow)
	}
	if state.StateTimer >= enemy.TypeConfig.ThrowWindupTime+15 {
		state.CurrentState = cfg.StatePatrol
		state.StateTimer = 0
		enemy.AttackCooldown = enemy.TypeConfig.ThrowCooldown
	}
}

func handleApproachEdgeState(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64) {
	enemy.Direction.X = math.Copysign(1, playerObject.X-enemyObject.X)

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

// collectEnemyPositions gathers the X position of every living enemy into a slice.
// Called once per UpdateEnemies frame so the separation calculation is O(n).
func collectEnemyPositions(e *ecs.ECS) []float64 {
	var positions []float64
	tags.Enemy.Each(e.World, func(entry *donburi.Entry) {
		if !entry.HasComponent(components.Death) {
			positions = append(positions, components.Object.Get(entry).X)
		}
	})
	return positions
}

// computeSeparationX returns a lateral force pushing this enemy away from nearby enemies.
// Force magnitude is proportional to overlap (strongest when touching, zero at radius).
func computeSeparationX(selfX float64, enemyPositions []float64, radius float64) float64 {
	force := 0.0
	for _, otherX := range enemyPositions {
		dx := selfX - otherX
		if dist := math.Abs(dx); dist < radius && dist > 0 {
			force += math.Copysign(1, dx) * (radius-dist) / radius
		}
	}
	return force
}

func updateEnemyAnimation(enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, animData *components.AnimationData) {
	var targetState cfg.StateID
	switch state.CurrentState {
	case cfg.StateAttackingPunch:
		targetState = cfg.Punch01
	case cfg.Throw:
		targetState = cfg.Throw
	case cfg.Hit:
		targetState = cfg.Hit
	case cfg.StateApproachEdge:
		targetState = cfg.Walk
	default:
		switch {
		case physics.OnGround == nil:
			targetState = cfg.Jump
		case physics.SpeedX != 0:
			targetState = cfg.Running
		default:
			targetState = cfg.Idle
		}
	}

	animData.SetAnimation(targetState)
	if animData.CurrentAnimation != nil {
		animData.CurrentAnimation.Update()
	}
}
