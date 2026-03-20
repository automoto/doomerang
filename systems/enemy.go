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

type enemyPos struct{ X, Y float64 }

var enemyPositionsBuf []enemyPos

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

func updateEnemyAI(e *ecs.ECS, enemyEntry *donburi.Entry, playerObject *resolv.Object, enemyPositions []enemyPos) {
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
		updateMeleeEnemyAI(e, enemyEntry, enemy, physics, state, enemyObject.Object, playerObject, distanceToPlayer, enemyPositions)
	}

	if enemy.LedgeCooldown > 0 {
		enemy.LedgeCooldown--
	}

	// Flip patrol direction when another enemy is ahead within separation radius.
	// Only applies during patrol — chase/attack states should converge on the player.
	// Cooldown prevents oscillation when a patrol boundary and separation conflict.
	if enemy.SeparationCooldown > 0 {
		enemy.SeparationCooldown--
	} else if state.CurrentState == cfg.StatePatrol && enemyAheadOnSameLevel(enemyObject.X, enemyObject.Y, enemy.Direction.X, enemyPositions, cfg.Enemy.SeparationRadius, cfg.Enemy.SeparationYThreshold) {
		enemy.Direction.X *= -1
		physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X
		enemy.SeparationCooldown = cfg.Enemy.SeparationCooldown
	}
}

func updateMeleeEnemyAI(e *ecs.ECS, enemyEntry *donburi.Entry, enemy *components.EnemyData, physics *components.PhysicsData, state *components.StateData, enemyObject, playerObject *resolv.Object, distanceToPlayer float64, enemyPositions []enemyPos) {
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
		handleChaseState(enemyEntry, playerObject, distanceToPlayer, enemyPositions)
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
	if enemy.LedgeCooldown == 0 && isAtPlatformEdge(enemyObject, enemy.Direction.X) {
		enemy.Direction.X *= -1
		physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X
		enemy.LedgeCooldown = cfg.Enemy.LedgeCooldown
	}
}

func handleDefaultPatrol(enemy *components.EnemyData, physics *components.PhysicsData, enemyObject *resolv.Object) {
	physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X
	if enemy.Direction.X > 0 && enemyObject.X >= enemy.PatrolRight {
		enemy.Direction.X = -1
	} else if enemy.Direction.X < 0 && enemyObject.X <= enemy.PatrolLeft {
		enemy.Direction.X = 1
	}
	if enemy.LedgeCooldown == 0 && isAtPlatformEdge(enemyObject, enemy.Direction.X) {
		enemy.Direction.X *= -1
		physics.SpeedX = enemy.PatrolSpeed * enemy.Direction.X
		enemy.LedgeCooldown = cfg.Enemy.LedgeCooldown
	}
}

func handleChaseState(enemyEntry *donburi.Entry, playerObject *resolv.Object, distanceToPlayer float64, enemyPositions []enemyPos) {
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

	if isAtPlatformEdge(enemyObject.Object, enemy.Direction.X) {
		physics.SpeedX = 0
	}

	// Rear enemy drifts backward instead of freezing when another enemy is ahead.
	if enemyAheadOnSameLevel(enemyObject.X, enemyObject.Y, enemy.Direction.X, enemyPositions, cfg.Enemy.SeparationRadius, cfg.Enemy.SeparationYThreshold) {
		physics.SpeedX = -enemy.Direction.X * cfg.Enemy.ChaseBackoffSpeed
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

// collectEnemyPositions gathers the position of every living enemy into a reusable buffer.
// Called once per UpdateEnemies frame so the separation calculation is O(n).
func collectEnemyPositions(e *ecs.ECS) []enemyPos {
	enemyPositionsBuf = enemyPositionsBuf[:0]
	tags.Enemy.Each(e.World, func(entry *donburi.Entry) {
		if !entry.HasComponent(components.Death) {
			obj := components.Object.Get(entry)
			enemyPositionsBuf = append(enemyPositionsBuf, enemyPos{X: obj.X, Y: obj.Y})
		}
	})
	return enemyPositionsBuf
}

// enemyAheadOnSameLevel returns true if any other enemy is within radius ahead of
// the current movement direction and within yThreshold vertically,
// signalling that patrol should reverse or chase should stop.
// Note: positions includes self, but self-match is harmless because dx == 0
// fails the dx*directionX > 0 check (strict inequality).
func enemyAheadOnSameLevel(selfX, selfY, directionX float64, positions []enemyPos, radius, yThreshold float64) bool {
	for _, other := range positions {
		if math.Abs(other.Y-selfY) > yThreshold {
			continue
		}
		dx := other.X - selfX
		if dx*directionX > 0 && math.Abs(dx) < radius {
			return true
		}
	}
	return false
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
