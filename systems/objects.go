package systems

import (
	"github.com/automoto/doomerang/components"
	dresolv "github.com/automoto/doomerang/resolv"
	"github.com/yohamta/donburi/ecs"
)

func UpdateObjects(ecs *ecs.ECS) {
	for e := range components.Object.Iter(ecs.World) {
		obj := dresolv.GetObject(e)
		obj.Update()
	}
}
