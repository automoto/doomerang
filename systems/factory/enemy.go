package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const (
	enemyFrameWidth      = 96
	enemyFrameHeight     = 84
	enemyCollisionWidth  = 16
	enemyCollisionHeight = 40 // Fixed: matches actual character height
)

// AI state constants
const (
	enemyStatePatrol = "patrol"
	enemyStateChase  = "chase"
	enemyStateAttack = "attack"
)

func CreateEnemy(ecs *ecs.ECS, x, y float64) *donburi.Entry {
	enemy := archetypes.Enemy.Spawn(ecs)

	// Create collision object
	obj := resolv.NewObject(x, y, enemyCollisionWidth, enemyCollisionHeight)
	cfg.SetObject(enemy, obj)
	obj.SetShape(resolv.NewRectangle(0, 0, enemyCollisionWidth, enemyCollisionHeight))
	obj.AddTags("character")
	// Set enemy data with AI parameters
	components.Enemy.SetValue(enemy, components.EnemyData{
		Direction:      components.Vector{X: -1, Y: 0}, // Start facing left
		CurrentState:   enemyStatePatrol,
		StateTimer:     0,
		PatrolLeft:     x - 16,
		PatrolRight:    x + 16,
		PatrolSpeed:    1.0,
		ChaseSpeed:     1.5,    // Faster when chasing
		AttackRange:    24.0,   // Attack when player within 24 pixels
		ChaseRange:     80.0,   // Start chasing when player within 80 pixels
		StoppingDistance: 20.0, // Stop 20 pixels away from player
		AttackCooldown: 0,
		InvulnFrames:   0,
	})

	// Set health (enemies have less health than player)
	components.Health.SetValue(enemy, components.HealthData{
		Current: 60, // Less health than player (100)
		Max:     60,
	})

	// Use same animations as player
	animData := GeneratePlayerAnimations() // Reuse from player factory
	animData.CurrentAnimation = animData.Animations[cfg.Idle]
	components.Animation.Set(enemy, animData)

	return enemy
}

// CreateTestEnemy spawns a hardcoded enemy for testing
func CreateTestEnemy(ecs *ecs.ECS) *donburi.Entry {
	// Spawn enemy at position (200, 128) - to the right of player spawn
	return CreateEnemy(ecs, 200, 128+float64(enemyFrameHeight-enemyCollisionHeight))
}
