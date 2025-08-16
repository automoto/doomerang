package components

import (
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
)

type EnemyData struct {
	SpeedX         float64
	SpeedY         float64
	OnGround       *resolv.Object
	FacingRight    bool
	
	// AI state management
	CurrentState   string        // Current AI state (patrol, chase, attack)
	StateTimer     int           // Frame counter for state duration
	PatrolLeft     float64       // Left boundary for patrol
	PatrolRight    float64       // Right boundary for patrol
	PatrolSpeed    float64       // Speed while patrolling
	ChaseSpeed     float64       // Speed while chasing player
	AttackRange    float64       // Distance to start attacking
	ChaseRange     float64       // Distance to start chasing
	StoppingDistance float64    // Distance to stop before attacking
	
	// Combat
	AttackCooldown int           // Frames until can attack again
	InvulnFrames   int           // Invincibility frames after being hit
}

var Enemy = donburi.NewComponentType[EnemyData]()