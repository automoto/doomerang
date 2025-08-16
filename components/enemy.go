package components

import (
	"github.com/yohamta/donburi"
)

type EnemyData struct {
	Direction Vector

	// AI state management
	PatrolLeft       float64 // Left boundary for patrol
	PatrolRight      float64 // Right boundary for patrol
	PatrolSpeed      float64 // Speed while patrolling
	ChaseSpeed       float64 // Speed while chasing player
	AttackRange      float64 // Distance to start attacking
	ChaseRange       float64 // Distance to start chasing
	StoppingDistance float64 // Distance to stop before attacking

	// Combat
	AttackCooldown int // Frames until can attack again
	InvulnFrames   int // Invincibility frames after being hit
}

var Enemy = donburi.NewComponentType[EnemyData]()