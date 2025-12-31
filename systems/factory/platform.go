package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/solarlune/resolv"
	"github.com/tanema/gween"
	"github.com/tanema/gween/ease"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func CreatePlatform(ecs *ecs.ECS, object *resolv.Object) *donburi.Entry {
	platform := archetypes.Platform.Spawn(ecs)
	object.Data = platform
	components.Object.SetValue(platform, components.ObjectData{Object: object})

	return platform
}

func CreateFloatingPlatform(ecs *ecs.ECS, object *resolv.Object) *donburi.Entry {
	platform := archetypes.FloatingPlatform.Spawn(ecs)
	object.Data = platform
	components.Object.SetValue(platform, components.ObjectData{Object: object})

	// The floating platform moves using a *gween.Sequence sequence of tweens, moving it back and forth.
	tw := gween.NewSequence()
	obj := components.Object.Get(platform)
	tw.Add(
		gween.New(float32(obj.Y), float32(obj.Y-128), 2, ease.Linear),
		gween.New(float32(obj.Y-128), float32(obj.Y), 2, ease.Linear),
	)
	components.Tween.Set(platform, tw)

	return platform
}
