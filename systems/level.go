package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi/ecs"
)

func DrawLevel(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return // No camera yet
	}
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(width)/2-camera.Position.X, float64(height)/2-camera.Position.Y)

	// Draw the level background
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		return
	}

	levelData := components.Level.Get(levelEntry)
	if levelData.CurrentLevel == nil {
		return
	}

	// Draw the background from the loaded level
	if levelData.CurrentLevel.Background != nil {
		screen.DrawImage(levelData.CurrentLevel.Background, opts)
	}

	// Ground objects are now drawn in the debug system when F1 debug mode is enabled
}
