package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateFinishLine checks for player collision with finish line and triggers level complete
func UpdateFinishLine(ecs *ecs.ECS) {
	// Skip if level is already complete
	levelComplete := GetOrCreateLevelComplete(ecs)
	if levelComplete.IsComplete {
		return
	}

	playerEntry, ok := tags.Player.First(ecs.World)
	if !ok {
		return
	}

	playerObj := components.Object.Get(playerEntry)

	// Check collision with finish line
	check := playerObj.Check(0, 0, tags.ResolvFinishLine)
	if check == nil {
		return
	}

	finishLineObjs := check.ObjectsByTags(tags.ResolvFinishLine)
	if len(finishLineObjs) == 0 {
		return
	}

	// Get the finish line entity from the resolv object
	finishLineResolv := finishLineObjs[0]
	finishLineEntry, ok := finishLineResolv.Data.(*donburi.Entry)
	if !ok || finishLineEntry == nil {
		return
	}

	finishLine := components.FinishLine.Get(finishLineEntry)

	// Only activate if not already activated
	if finishLine.Activated {
		return
	}

	// Activate finish line and trigger level complete
	finishLine.Activated = true
	levelComplete.IsComplete = true

	// Pause music when level completes
	PauseMusic(ecs)
}
