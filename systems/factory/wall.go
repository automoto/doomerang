package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func CreateWall(ecs *ecs.ECS, obj *resolv.Object) *donburi.Entry {
	wall := archetypes.Wall.Spawn(ecs)
	components.Object.Set(wall, obj)
	return wall
}
