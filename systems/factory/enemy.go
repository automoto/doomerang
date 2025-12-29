package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func init() {
	// Define enemy types
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
		TintColor:        cfg.White,
	}
	
	lightGuardType := cfg.EnemyTypeConfig{
		Name:             "LightGuard",
		Health:           40,
		PatrolSpeed:      3.0,
		ChaseSpeed:       3.5,
		AttackRange:      32.0,
		ChaseRange:       100.0,
		StoppingDistance: 24.0,
		AttackCooldown:   40,
		InvulnFrames:     10,
		AttackDuration:   20,
		HitstunDuration:  10,
		Damage:           8,
		KnockbackForce:   3.0,
		Gravity:          0.8,
		Friction:         0.25,
		MaxSpeed:         7.0,
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   14,
		CollisionHeight:  36,
		TintColor:        cfg.Yellow,
	}
	
	heavyGuardType := cfg.EnemyTypeConfig{
		Name:             "HeavyGuard",
		Health:           100,
		PatrolSpeed:      1.5,
		ChaseSpeed:       2.0,
		AttackRange:      40.0,
		ChaseRange:       60.0,
		StoppingDistance: 32.0,
		AttackCooldown:   90,
		InvulnFrames:     25,
		AttackDuration:   45,
		HitstunDuration:  25,
		Damage:           18,
		KnockbackForce:   8.0,
		Gravity:          0.7,
		Friction:         0.15,
		MaxSpeed:         4.0,
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   20,
		CollisionHeight:  44,
		TintColor:        cfg.Orange,
	}

	cfg.Enemy = cfg.EnemyConfig{
		Types: map[string]cfg.EnemyTypeConfig{
			"Guard":      guardType,
			"LightGuard": lightGuardType,
			"HeavyGuard": heavyGuardType,
		},
		HysteresisMultiplier:  1.5,
		DefaultPatrolDistance: 32.0,
	}
}

func CreateEnemy(ecs *ecs.ECS, x, y float64, patrolPath string, enemyTypeName string) *donburi.Entry {
	// Use the requested enemy type, default to "Guard" if not found
	enemyType, exists := cfg.Enemy.Types[enemyTypeName]
	if !exists {
		enemyTypeName = "Guard"
		enemyType = cfg.Enemy.Types[enemyTypeName] // Fallback to default
	}

	enemy := archetypes.Enemy.Spawn(ecs)

	// Create collision object
	obj := resolv.NewObject(x, y, float64(enemyType.CollisionWidth), float64(enemyType.CollisionHeight))
	components.Object.SetValue(enemy, components.ObjectData{Object: obj})
	obj.SetShape(resolv.NewRectangle(0, 0, float64(enemyType.CollisionWidth), float64(enemyType.CollisionHeight)))
	obj.AddTags("character")

	// Set enemy data with AI parameters from config
	enemyData := components.EnemyData{
		TypeName:         enemyTypeName, // Set the enemy type name
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
		CurrentState: cfg.StatePatrol,
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
	return CreateEnemy(ecs, 200, 128+float64(enemyType.FrameHeight-enemyType.CollisionHeight), "", "Guard")
}
