package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// CreateFinishLine creates a finish line entity with collision detection
func CreateFinishLine(ecs *ecs.ECS, x, y, w, h float64) *donburi.Entry {
	finishLine := archetypes.FinishLine.Spawn(ecs)

	obj := resolv.NewObject(x, y, w, h, tags.ResolvFinishLine)
	obj.SetShape(resolv.NewRectangle(0, 0, w, h))
	obj.Data = finishLine

	components.Object.SetValue(finishLine, components.ObjectData{Object: obj})
	components.FinishLine.SetValue(finishLine, components.FinishLineData{
		Activated: false,
	})

	// Add to physics space
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		components.Space.Get(spaceEntry).Add(obj)
	}

	return finishLine
}
