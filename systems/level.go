package systems

import (
    "image/color"

    "github.com/automoto/doomerang/components"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/vector"
    "github.com/yohamta/donburi/ecs"
)

func DrawLevel(ecs *ecs.ECS, screen *ebiten.Image) {
    // Get camera
    cameraEntry, _ := components.Camera.First(ecs.World)
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

    // Draw the ground objects from the loaded level
    for _, path := range levelData.CurrentLevel.Paths {
        if len(path.Points) >= 2 {
            p1 := path.Points[0]
            p2 := path.Points[1]
            drawColor := color.RGBA{60, 60, 60, 255}
            // This debug draw doesn't need to be camera-aware, as it's for debugging.
            vector.DrawFilledRect(screen, float32(p1.X), float32(p1.Y), float32(p2.X-p1.X), float32(p2.Y-p1.Y), drawColor, false)
        }
    }
}
