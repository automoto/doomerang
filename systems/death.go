package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateDeaths(ecs *ecs.ECS) {
	components.Death.Each(ecs.World, func(e *donburi.Entry) {
		death := components.Death.Get(e)
		death.Timer--
		if death.Timer <= 0 {
			// Get the space and object to remove the object from the physics world.
			spaceEntry, _ := components.Space.First(e.World)
			space := components.Space.Get(spaceEntry)
			if obj := components.Object.Get(e); obj != nil {
				space.Remove(obj.Object)
			}
			ecs.World.Remove(e.Entity())
		}
	})
}
