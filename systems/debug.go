package systems

import (
	"image/color"

	"github.com/automoto/doomerang/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

func DrawDebug(ecs *ecs.ECS, screen *ebiten.Image) {
	settings := GetOrCreateSettings(ecs)
	if !settings.Debug {
		return
	}

	// Get camera for world-space rendering.
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return // No camera yet
	}
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	camX := float64(width)/2 - camera.Position.X
	camY := float64(height)/2 - camera.Position.Y

	// Draw collision grid (Disabled, showing objects only)
	spaceEntry, ok := components.Space.First(ecs.World)
	if ok {
		space := components.Space.Get(spaceEntry)

		// Viewport in world coordinates
		viewX := camera.Position.X - float64(width)/2
		viewY := camera.Position.Y - float64(height)/2
		viewW := float64(width)
		viewH := float64(height)

		// Draw all collision objects in the space (Entities)
		// Viewport for culling (already calculated above)

		for _, obj := range space.Objects() {
			// Cull objects outside viewport
			if obj.X+obj.W < viewX || obj.X > viewX+viewW || obj.Y+obj.H < viewY || obj.Y > viewY+viewH {
				continue
			}

			// Apply camera offset
			x := obj.X + camX
			y := obj.Y + camY

			// Determine color based on tags
			c := color.RGBA{0, 255, 255, 255} // Cyan default
			if obj.HasTags("solid") {
				c = color.RGBA{100, 100, 100, 255} // Grey
			} else if obj.HasTags("Player") {
				c = color.RGBA{0, 0, 255, 255} // Blue
			} else if obj.HasTags("Enemy") {
				c = color.RGBA{255, 0, 0, 255} // Red
			} else if obj.HasTags("Boomerang") {
				c = color.RGBA{0, 255, 0, 255} // Green
			}

			// Draw outline
			vector.DrawFilledRect(screen, float32(x), float32(y), float32(obj.W), 1, c, false)         // Top
			vector.DrawFilledRect(screen, float32(x), float32(y+obj.H-1), float32(obj.W), 1, c, false) // Bottom
			vector.DrawFilledRect(screen, float32(x), float32(y), 1, float32(obj.H), c, false)         // Left
			vector.DrawFilledRect(screen, float32(x+obj.W-1), float32(y), 1, float32(obj.H), c, false) // Right
		}
	}
}
