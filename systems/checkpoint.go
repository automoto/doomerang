package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateCheckpoints checks for player collision with checkpoints and activates them
func UpdateCheckpoints(ecs *ecs.ECS) {
	playerEntry, ok := tags.Player.First(ecs.World)
	if !ok {
		return
	}

	playerObj := components.Object.Get(playerEntry)

	// Check collision with checkpoints
	check := playerObj.Check(0, 0, tags.ResolvCheckpoint)
	if check == nil {
		return
	}

	checkpointObjs := check.ObjectsByTags(tags.ResolvCheckpoint)
	if len(checkpointObjs) == 0 {
		return
	}

	// Get the checkpoint entity from the resolv object
	checkpointResolv := checkpointObjs[0]
	checkpointEntry, ok := checkpointResolv.Data.(*donburi.Entry)
	if !ok || checkpointEntry == nil {
		return
	}

	checkpoint := components.Checkpoint.Get(checkpointEntry)

	// Only activate if not already activated
	if checkpoint.Activated {
		return
	}

	// Activate checkpoint
	checkpoint.Activated = true

	// Update level's active checkpoint
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		return
	}

	levelData := components.Level.Get(levelEntry)
	levelData.ActiveCheckpoint = &components.ActiveCheckpointData{
		SpawnX:       checkpoint.SpawnX,
		SpawnY:       checkpoint.SpawnY,
		CheckpointID: checkpoint.CheckpointID,
	}

	SaveGameProgress(levelData.LevelIndex, levelData.ActiveCheckpoint)
}
