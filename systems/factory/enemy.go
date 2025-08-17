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
		Direction:        components.Vector{X: -1, Y: 0}, // Start facing left
		PatrolLeft:       x - 16,
		PatrolRight:      x + 16,
		PatrolSpeed:      2.0,
		ChaseSpeed:       2.5,  // Faster when chasing
		AttackRange:      36.0, // Attack when player within 32 pixels
		ChaseRange:       80.0, // Start chasing when player within 80 pixels
		StoppingDistance: 28.0, // Stop 24 pixels away from player
		AttackCooldown:   0,
		InvulnFrames:     0,
	})
	components.State.SetValue(enemy, components.StateData{
		CurrentState: "patrol",
		StateTimer:   0,
	})
	components.Physics.SetValue(enemy, components.PhysicsData{
		Gravity:  0.75,
		Friction: 0.2,
		MaxSpeed: 6.0,
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
