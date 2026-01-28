package archetypes

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

var (
	Platform = newArchetype(
		tags.Platform,
		components.Object,
	)
	FloatingPlatform = newArchetype(
		tags.FloatingPlatform,
		components.Object,
		components.Tween,
	)
	Player = newArchetype(
		tags.Player,
		components.Player,
		components.Object,
		components.Health,
		components.Animation,
		components.Physics,
		components.State,
		components.MeleeAttack,
		components.Lives,
		components.Flash,
	)
	Enemy = newArchetype(
		tags.Enemy,
		components.Enemy,
		components.Object,
		components.Health,
		components.Animation,
		components.Physics,
		components.State,
		components.Flash,
	)
	Hitbox = newArchetype(
		tags.Hitbox,
		components.Hitbox,
		components.Object,
	)
	Space = newArchetype(
		components.Space,
	)
	Wall = newArchetype(
		tags.Wall,
		components.Object,
	)
	Boomerang = newArchetype(
		tags.Boomerang,
		components.Boomerang,
		components.Object,
		components.Sprite,
		components.Physics,
	)
	Knife = newArchetype(
		tags.Knife,
		components.Knife,
		components.Object,
		components.Sprite,
		components.Physics,
	)
	Level = newArchetype(
		components.Level,
	)
	Camera = newArchetype(
		components.Camera,
	)
	VFXEffect = newArchetype(
		components.Object,
		components.Animation,
		components.AutoDestroy,
	)
	Checkpoint = newArchetype(
		tags.Checkpoint,
		components.Checkpoint,
		components.Object,
	)
	Fire = newArchetype(
		tags.Fire,
		components.Fire,
		components.Object,
		components.Animation,
	)
	MessagePoint = newArchetype(
		components.MessagePoint,
	)
	FinishLine = newArchetype(
		tags.FinishLine,
		components.FinishLine,
		components.Object,
	)
)

type archetype struct {
	components []donburi.IComponentType
}

func newArchetype(cs ...donburi.IComponentType) *archetype {
	return &archetype{
		components: cs,
	}
}

func (a *archetype) Spawn(ecs *ecs.ECS, cs ...donburi.IComponentType) *donburi.Entry {
	e := ecs.World.Entry(ecs.Create(
		cfg.Default,
		append(a.components, cs...)...,
	))
	return e
}
