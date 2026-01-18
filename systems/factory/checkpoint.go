package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// CreateCheckpoint creates a checkpoint entity with collision detection
func CreateCheckpoint(ecs *ecs.ECS, x, y, w, h, checkpointID float64) *donburi.Entry {
	checkpoint := archetypes.Checkpoint.Spawn(ecs)

	obj := resolv.NewObject(x, y, w, h, tags.ResolvCheckpoint)
	obj.SetShape(resolv.NewRectangle(0, 0, w, h))
	obj.Data = checkpoint

	components.Object.SetValue(checkpoint, components.ObjectData{Object: obj})

	// Calculate spawn position at center of checkpoint
	spawnX := x + w/2
	spawnY := y + h/2

	components.Checkpoint.SetValue(checkpoint, components.CheckpointData{
		CheckpointID: checkpointID,
		Activated:    false,
		SpawnX:       spawnX,
		SpawnY:       spawnY,
	})

	// Add to physics space
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		components.Space.Get(spaceEntry).Add(obj)
	}

	return checkpoint
}
