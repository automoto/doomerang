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

func CreatePlayer(ecs *ecs.ECS, x, y float64) *donburi.Entry {
	player := archetypes.Player.Spawn(ecs)

	obj := resolv.NewObject(x, y, float64(cfg.Player.CollisionWidth), float64(cfg.Player.CollisionHeight))
	components.Object.SetValue(player, components.ObjectData{Object: obj})
	obj.AddTags("character", tags.ResolvPlayer)
	obj.Data = player
	components.Player.SetValue(player, components.PlayerData{
		Direction:    components.Vector{X: 1, Y: 0},
		ComboCounter: 0,
		InvulnFrames: 0,
	})
	components.State.SetValue(player, components.StateData{
		CurrentState:  cfg.Idle,
		PreviousState: cfg.StateNone,
		StateTimer:    0,
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
	components.Lives.SetValue(player, components.LivesData{
		Lives:    cfg.Player.StartingLives,
		MaxLives: cfg.Player.StartingLives,
	})

	obj.SetShape(resolv.NewRectangle(0, 0, float64(cfg.Player.CollisionWidth), float64(cfg.Player.CollisionHeight)))

	// Load sprite sheets
	animData := GenerateAnimations("player", cfg.Player.FrameWidth, cfg.Player.FrameHeight)
	animData.CurrentAnimation = animData.Animations[cfg.Idle]
	components.Animation.Set(player, animData)

	// Initialize Flash component (permanently attached to avoid archetype thrashing)
	components.Flash.SetValue(player, components.FlashData{
		Duration: 0,
		R: 1, G: 1, B: 1,
	})

	return player
}
