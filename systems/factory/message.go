package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// CreateMessagePoint creates a message point entity from Tiled data
func CreateMessagePoint(ecs *ecs.ECS, x, y float64, messageID float64) *donburi.Entry {
	entry := archetypes.MessagePoint.Spawn(ecs)

	components.MessagePoint.SetValue(entry, components.MessagePointData{
		MessageID: messageID,
		X:         x,
		Y:         y,
	})

	return entry
}
