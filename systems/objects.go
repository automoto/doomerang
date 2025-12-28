package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi/ecs"
)

func UpdateObjects(ecs *ecs.ECS) {
	for e := range components.Object.Iter(ecs.World) {
		obj := components.Object.Get(e)
		obj.Update()
	}
}
