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

	// Draw collision grid
	spaceEntry, ok := components.Space.First(ecs.World)
	if ok {
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

	// Draw level collision objects
	levelEntry, ok := components.Level.First(ecs.World)
	if ok {
		levelData := components.Level.Get(levelEntry)
		if levelData.CurrentLevel != nil {
			for _, path := range levelData.CurrentLevel.Paths {
				if len(path.Points) >= 2 {
					p1 := path.Points[0]
					p2 := path.Points[1]
					// Apply camera transformation to collision rectangles
					rectX := float32(p1.X + camX)
					rectY := float32(p1.Y + camY)
					rectW := float32(p2.X - p1.X)
					rectH := float32(p2.Y - p1.Y)
					drawColor := color.RGBA{60, 60, 60, 128} // Semi-transparent grey
					ebitenutil.DrawRect(screen, float64(rectX), float64(rectY), float64(rectW), float64(rectH), drawColor)
				}
			}
		}
	}

	// Draw all collision objects in the space (Entities)
	if ok { // reusing spaceEntry check from above
		space := components.Space.Get(spaceEntry)
		for _, obj := range space.Objects() {
			// Apply camera offset
			x := obj.X + camX
			y := obj.Y + camY

			// Determine color based on tags
			c := color.RGBA{0, 255, 255, 100} // Cyan default
			if obj.HasTags("solid") {
				c = color.RGBA{100, 100, 100, 100} // Grey
			} else if obj.HasTags("Player") {
				c = color.RGBA{0, 0, 255, 100} // Blue
			} else if obj.HasTags("Enemy") {
				c = color.RGBA{255, 0, 0, 100} // Red
			} else if obj.HasTags("Boomerang") {
				c = color.RGBA{0, 255, 0, 200} // Green, more opaque
			}

			// Draw outline
			ebitenutil.DrawRect(screen, x, y, obj.W, 1, c)         // Top
			ebitenutil.DrawRect(screen, x, y+obj.H-1, obj.W, 1, c) // Bottom
			ebitenutil.DrawRect(screen, x, y, 1, obj.H, c)         // Left
			ebitenutil.DrawRect(screen, x+obj.W-1, y, 1, obj.H, c) // Right
		}
	}
}
