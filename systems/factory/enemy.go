package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// AI state constants
const (
	enemyStatePatrol = "patrol"
	enemyStateChase  = "chase"
	enemyStateAttack = "attack"
)

func init() {
	// Define a default "Guard" enemy type
	guardType := cfg.EnemyTypeConfig{
		Name:             "Guard",
		Health:           60,
		PatrolSpeed:      2.0,
		ChaseSpeed:       2.5,
		AttackRange:      36.0,
		ChaseRange:       80.0,
		StoppingDistance: 28.0,
		AttackCooldown:   60,
		InvulnFrames:     15,
		AttackDuration:   30,
		HitstunDuration:  15,
		Damage:           10,
		KnockbackForce:   5.0,
		Gravity:          0.75,
		Friction:         0.2,
		MaxSpeed:         6.0,
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   16,
		CollisionHeight:  40,
	}

	cfg.Enemy = cfg.EnemyConfig{
		Types: map[string]cfg.EnemyTypeConfig{
			"Guard": guardType,
		},
		HysteresisMultiplier:  1.5,
		DefaultPatrolDistance: 32.0,
	}
}

func CreateEnemy(ecs *ecs.ECS, x, y float64, patrolPath string) *donburi.Entry {
	// Use the default "Guard" type for now
	enemyType := cfg.Enemy.Types["Guard"]

	enemy := archetypes.Enemy.Spawn(ecs)

	// Create collision object
	obj := resolv.NewObject(x, y, float64(enemyType.CollisionWidth), float64(enemyType.CollisionHeight))
	cfg.SetObject(enemy, obj)
	obj.SetShape(resolv.NewRectangle(0, 0, float64(enemyType.CollisionWidth), float64(enemyType.CollisionHeight)))
	obj.AddTags("character")

	// Set enemy data with AI parameters from config
	enemyData := components.EnemyData{
		Direction:        components.Vector{X: -1, Y: 0}, // Start facing left
		PatrolSpeed:      enemyType.PatrolSpeed,
		ChaseSpeed:       enemyType.ChaseSpeed,
		AttackRange:      enemyType.AttackRange,
		ChaseRange:       enemyType.ChaseRange,
		StoppingDistance: enemyType.StoppingDistance,
		AttackCooldown:   0,
		InvulnFrames:     0,
	}

	// Set patrol boundaries based on whether we have a custom patrol path
	if patrolPath != "" {
		// Custom patrol path will be handled in the AI system
		enemyData.PatrolPathName = patrolPath
		// Initialize default patrol boundaries (will be overridden by custom path)
		enemyData.PatrolLeft = x
		enemyData.PatrolRight = x
	} else {
		// Default patrol behavior (back and forth from current position)
		enemyData.PatrolLeft = x - cfg.Enemy.DefaultPatrolDistance
		enemyData.PatrolRight = x + cfg.Enemy.DefaultPatrolDistance
	}

	components.Enemy.SetValue(enemy, enemyData)
	components.State.SetValue(enemy, components.StateData{
		CurrentState: "patrol",
		StateTimer:   0,
	})
	components.Physics.SetValue(enemy, components.PhysicsData{
		Gravity:  enemyType.Gravity,
		Friction: enemyType.Friction,
		MaxSpeed: enemyType.MaxSpeed,
	})

	// Set health from config
	components.Health.SetValue(enemy, components.HealthData{
		Current: enemyType.Health,
		Max:     enemyType.Health,
	})

	// Use same animations as player
	animData := GeneratePlayerAnimations() // Reuse from player factory
	animData.CurrentAnimation = animData.Animations[cfg.Idle]
	components.Animation.Set(enemy, animData)

	return enemy
}

// CreateTestEnemy spawns a hardcoded enemy for testing
func CreateTestEnemy(ecs *ecs.ECS) *donburi.Entry {
	enemyType := cfg.Enemy.Types["Guard"]
	return CreateEnemy(ecs, 200, 128+float64(enemyType.FrameHeight-enemyType.CollisionHeight), "")
}
