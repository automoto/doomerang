package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi/ecs"
)

func UpdateCamera(ecs *ecs.ECS) {
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)

	playerEntry, _ := tags.Player.First(ecs.World)
	playerObject := components.Object.Get(playerEntry)

	// Get level dimensions for camera bounds
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		return
	}
	levelData := components.Level.Get(levelEntry)
	if levelData.CurrentLevel == nil {
		return
	}

	// Calculate target camera position (following the player with smoothing)
	targetX := playerObject.X
	targetY := playerObject.Y

	// Calculate camera bounds based on screen and level dimensions
	screenWidth := float64(config.C.Width)
	screenHeight := float64(config.C.Height)
	levelWidth := float64(levelData.CurrentLevel.Width)
	levelHeight := float64(levelData.CurrentLevel.Height)

	// Camera bounds: ensure the level always fills the screen
	minCameraX := screenWidth / 2
	maxCameraX := levelWidth - screenWidth/2
	minCameraY := screenHeight / 2
	maxCameraY := levelHeight - screenHeight/2

	// Constrain target position to camera bounds
	targetX = math.Max(minCameraX, math.Min(maxCameraX, targetX))
	targetY = math.Max(minCameraY, math.Min(maxCameraY, targetY))

	// Center the camera on the constrained target position, with some smoothing.
	camera.Position.X += (targetX - camera.Position.X) * 0.1
	camera.Position.Y += (targetY - camera.Position.Y) * 0.1
}
