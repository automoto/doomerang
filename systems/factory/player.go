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
	cfg.Player = cfg.PlayerConfig{
		// Movement
		JumpSpeed:    15.0,
		Acceleration: 0.75,
		AttackAccel:  0.1,
		MaxSpeed:     6.0,

		// Combat
		Health:       100,
		InvulnFrames: 30,

		// Physics
		Gravity:        0.75,
		Friction:       0.5,
		AttackFriction: 0.2,

		// Dimensions
		FrameWidth:      96,
		FrameHeight:     84,
		CollisionWidth:  16,
		CollisionHeight: 40,
	}
}

func CreatePlayer(ecs *ecs.ECS, x, y float64) *donburi.Entry {
	player := archetypes.Player.Spawn(ecs)

	obj := resolv.NewObject(x, y, float64(cfg.Player.CollisionWidth), float64(cfg.Player.CollisionHeight))
	components.Object.SetValue(player, components.ObjectData{Object: obj})
	obj.AddTags("character", "Player")
	components.Player.SetValue(player, components.PlayerData{
		Direction:    components.Vector{X: 1, Y: 0},
		ComboCounter: 0,
	})
	components.State.SetValue(player, components.StateData{
		CurrentState: cfg.Idle,
		StateTimer:   0,
	})
	components.Physics.SetValue(player, components.PhysicsData{
		Gravity:        cfg.Player.Gravity,
		Friction:       cfg.Player.Friction,
		AttackFriction: cfg.Player.AttackFriction,
		MaxSpeed:       cfg.Player.MaxSpeed,
	})
	components.Health.SetValue(player, components.HealthData{
		Current: cfg.Player.Health,
		Max:     cfg.Player.Health,
	})

	obj.SetShape(resolv.NewRectangle(0, 0, float64(cfg.Player.CollisionWidth), float64(cfg.Player.CollisionHeight)))

	// Load sprite sheets
	animData := GenerateAnimations("player", cfg.Player.FrameWidth, cfg.Player.FrameHeight)
	animData.CurrentAnimation = animData.Animations[cfg.Idle]
	components.Animation.Set(player, animData)

	return player
}

