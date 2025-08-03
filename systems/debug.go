package systems

import (
    "image/color"

    "github.com/automoto/doomerang/components"
    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
    "github.com/yohamta/donburi/ecs"
)

func DrawDebug(ecs *ecs.ECS, screen *ebiten.Image) {
    settings := GetOrCreateSettings(ecs)
    if !settings.Debug {
        return
    }

    // Get camera for world-space rendering.
    cameraEntry, _ := components.Camera.First(ecs.World)
    camera := components.Camera.Get(cameraEntry)
    width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
    camX := float64(width)/2 - camera.Position.X
    camY := float64(height)/2 - camera.Position.Y

    spaceEntry, ok := components.Space.First(ecs.World)
    if !ok {
        return
    }
    space := components.Space.Get(spaceEntry)

    for y := 0; y < space.Height(); y++ {
        for x := 0; x < space.Width(); x++ {
            cell := space.Cell(x, y)

            cw := float64(space.CellWidth)
            ch := float64(space.CellHeight)
            // Apply camera offset to the cell's position.
            cx := float64(cell.X)*cw + camX
            cy := float64(cell.Y)*ch + camY

            drawColor := color.RGBA{20, 20, 20, 255}

            if cell.Occupied() {
                drawColor = color.RGBA{255, 255, 0, 255}
            }

            // Draw the grid lines directly to the screen in world space.
            ebitenutil.DrawLine(screen, cx, cy, cx+cw, cy, drawColor)
            ebitenutil.DrawLine(screen, cx+cw, cy, cx+cw, cy+ch, drawColor)
            ebitenutil.DrawLine(screen, cx+cw, cy+ch, cx, cy+ch, drawColor)
            ebitenutil.DrawLine(screen, cx, cy+ch, cx, cy, drawColor)
        }
    }
}
