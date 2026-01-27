package components

import "github.com/yohamta/donburi"

type CheckpointData struct {
	CheckpointID float64
	Activated    bool
	SpawnX       float64 // Respawn position (center of checkpoint)
	SpawnY       float64
}

var Checkpoint = donburi.NewComponentType[CheckpointData]()

// ActiveCheckpointData is stored in LevelData to track the last activated checkpoint
type ActiveCheckpointData struct {
	SpawnX       float64
	SpawnY       float64
	CheckpointID float64
}
