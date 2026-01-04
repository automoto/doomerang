package factory

import (
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi/ecs"
)

// CreateDeadZone creates an invisible collision zone that triggers death when touched
func CreateDeadZone(ecs *ecs.ECS, x, y, w, h float64) *resolv.Object {
	obj := resolv.NewObject(x, y, w, h, tags.ResolvDeadZone)
	obj.SetShape(resolv.NewRectangle(0, 0, w, h))

	// Add to space if it exists
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		components.Space.Get(spaceEntry).Add(obj)
	}

	return obj
}
