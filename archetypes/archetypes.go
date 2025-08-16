package archetypes

import (
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/layers"
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
	)
	Enemy = newArchetype(
		tags.Enemy,
		components.Enemy,
		components.Object,
		components.Health,
		components.Animation,
		components.Physics,
		components.State,
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
	Level = newArchetype(
		components.Level,
	)
	Camera = newArchetype(
		components.Camera,
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
		layers.Default,
		append(a.components, cs...)...,
	))
	return e
}
