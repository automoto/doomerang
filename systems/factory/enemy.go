package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

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
	obj.AddTags("character", tags.ResolvEnemy)
	obj.Data = enemy

	// Set enemy data with AI parameters from config
	enemyData := components.EnemyData{
		TypeName:         enemyTypeName,                  // Set the enemy type name
		TypeConfig:       &enemyType,                     // Cache the config reference
		Direction:        components.Vector{X: -1, Y: 0}, // Start facing left
		PatrolSpeed:      enemyType.PatrolSpeed,
		ChaseSpeed:       enemyType.ChaseSpeed,
		AttackRange:      enemyType.AttackRange,
		ChaseRange:       enemyType.ChaseRange,
		StoppingDistance: enemyType.StoppingDistance,
		AttackCooldown:   0,
		InvulnFrames:     0,
	}

	// Pre-calculate and cache color tint
	enemyData.TintColor.Reset()
	if enemyType.TintColor.R != 255 || enemyType.TintColor.G != 255 || enemyType.TintColor.B != 255 || enemyType.TintColor.A != 255 {
		r := float32(enemyType.TintColor.R) / 255.0
		g := float32(enemyType.TintColor.G) / 255.0
		b := float32(enemyType.TintColor.B) / 255.0
		a := float32(enemyType.TintColor.A) / 255.0
		enemyData.TintColor.Scale(r, g, b, a)
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
		CurrentState:  cfg.StatePatrol,
		PreviousState: cfg.StateNone,
		StateTimer:    0,
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
	animData := GenerateAnimations(enemyType.SpriteSheetKey, enemyType.FrameWidth, enemyType.FrameHeight)
	animData.CurrentAnimation = animData.Animations[cfg.Idle]
	components.Animation.Set(enemy, animData)

	// Initialize Flash component (permanently attached to avoid archetype thrashing)
	components.Flash.SetValue(enemy, components.FlashData{
		Duration: 0,
		R: 1, G: 1, B: 1,
	})

	return enemy
}

// CreateTestEnemy spawns a hardcoded enemy for testing
func CreateTestEnemy(ecs *ecs.ECS) *donburi.Entry {
	enemyType := cfg.Enemy.Types["Guard"]
	return CreateEnemy(ecs, 200, 128+float64(enemyType.FrameHeight-enemyType.CollisionHeight), "", "Guard")
}
