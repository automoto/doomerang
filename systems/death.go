package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi/ecs"
)

// UpdateDeaths decrements death timers and removes entities when the timer
// expires.
func UpdateDeaths(ecs *ecs.ECS) {
	for e := range components.Death.Iter(ecs.World) {
		d := components.Death.Get(e)
		d.Timer--
		if d.Timer <= 0 {
			// Remove the entity from the world.
			e.Remove()
		}
	}
}
