package components

import (
	"github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type EnemyData struct {
	TypeName   string                  // "Guard", "LightGuard", "HeavyGuard" etc...
	TypeConfig *config.EnemyTypeConfig // Cached reference to type configuration
	Direction  Vector
	TintColor  ebiten.ColorScale // Cached color tint from enemy type config

	// AI state management
	PatrolLeft       float64 // Left boundary for patrol
	PatrolRight      float64 // Right boundary for patrol
	PatrolSpeed      float64 // Speed while patrolling
	PatrolPathName   string  // Name of custom patrol path (if any)
	ChaseSpeed       float64 // Speed while chasing player
	AttackRange      float64 // Distance to start attacking
	ChaseRange       float64 // Distance to start chasing
	StoppingDistance float64 // Distance to stop before attacking

	// Combat
	AttackCooldown int            // Frames until can attack again
	InvulnFrames   int            // Invincibility frames after being hit
	ActiveHitbox   *donburi.Entry // Direct reference to the active hitbox
}

var Enemy = donburi.NewComponentType[EnemyData]()
