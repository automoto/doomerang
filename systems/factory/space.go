package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func CreateSpace(ecs *ecs.ECS, width, height, cellWidth, cellHeight int) *donburi.Entry {
	space := archetypes.Space.Spawn(ecs)
	spaceData := resolv.NewSpace(width, height, cellWidth, cellHeight)
	components.Space.Set(space, spaceData)
	return space
}
