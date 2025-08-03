package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi/ecs"
)

func CreateCamera(ecs *ecs.ECS) {
	camera := archetypes.Camera.Spawn(ecs)
	components.Camera.Set(camera, &components.CameraData{})
}
