package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func CreateWall(ecs *ecs.ECS, x, y, w, h float64) *donburi.Entry {
	wall := archetypes.Wall.Spawn(ecs)

	// Create collision object
	obj := resolv.NewObject(x, y, w, h, tags.ResolvSolid)
	obj.SetShape(resolv.NewRectangle(0, 0, w, h))
	obj.Data = wall // Link for O(1) lookup

	components.Object.SetValue(wall, components.ObjectData{Object: obj})

	// Add to space if it exists
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		components.Space.Get(spaceEntry).Add(obj)
	}

	return wall
}

// CreateSlopeWall creates a slope tile for ramp collision
// Uses rectangular bounds for detection, surface height is calculated mathematically
func CreateSlopeWall(ecs *ecs.ECS, x, y, w, h float64, slopeType string) *donburi.Entry {
	wall := archetypes.Wall.Spawn(ecs)

	// Create collision object with "ramp" tag and slope type tag
	// Use rectangle bounds - actual slope surface is calculated in collision code
	obj := resolv.NewObject(x, y, w, h, tags.ResolvRamp, slopeType)
	obj.SetShape(resolv.NewRectangle(0, 0, w, h))
	obj.Data = wall

	components.Object.SetValue(wall, components.ObjectData{Object: obj})

	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		components.Space.Get(spaceEntry).Add(obj)
	}

	return wall
}
