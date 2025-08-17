package systems

import (
	"image"
	"image/color"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func DrawAnimated(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	components.Animation.Each(ecs.World, func(e *donburi.Entry) {
		o := cfg.GetObject(e)
		animData := components.Animation.Get(e)

		if animData.CurrentAnimation != nil && animData.SpriteSheets[animData.CurrentSheet] != nil {
			// Calculate the source rectangle for the current frame.
			frame := animData.CurrentAnimation.Frame()
			sx := frame * animData.FrameWidth
			sy := 0
			srcRect := image.Rect(sx, sy, sx+animData.FrameWidth, sy+animData.FrameHeight)

			// Create draw options.
			op := &ebiten.DrawImageOptions{}

			// Anchor the sprite at its bottom-center so that the feet line up with the
			// bottom of the collision box.
			op.GeoM.Translate(-float64(animData.FrameWidth)/2, -float64(animData.FrameHeight))

			// Flip the sprite if facing left.
			if e.HasComponent(components.Player) {
				player := components.Player.Get(e)
				if player.Direction.X < 0 {
					op.GeoM.Scale(-1, 1)
				}
			}
			if e.HasComponent(components.Enemy) {
				enemy := components.Enemy.Get(e)
				if enemy.Direction.X < 0 {
					op.GeoM.Scale(-1, 1)
				}
			}

			// Move the sprite so that its bottom-center aligns with the bottom-center
			// of the (smaller) collision box.
			op.GeoM.Translate(o.X+o.W/2, o.Y+o.H)

			// Apply the camera translation.
			op.GeoM.Translate(float64(width)/2-camera.Position.X, float64(height)/2-camera.Position.Y)

			// Flicker effect if invulnerable
			if e.HasComponent(components.Enemy) {
				enemy := components.Enemy.Get(e)
				if enemy.InvulnFrames > 0 && enemy.InvulnFrames%4 < 2 {
					op.ColorM.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
				}
			}
			if e.HasComponent(components.Player) {
				player := components.Player.Get(e)
				if player.InvulnFrames > 0 && player.InvulnFrames%4 < 2 {
					op.ColorM.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
				}
			}

			// Draw the current frame.
			screen.DrawImage(animData.SpriteSheets[animData.CurrentSheet].SubImage(srcRect).(*ebiten.Image), op)
		} else {
			// Fallback to rectangle if no animation is available
			var entityColor color.Color
			if e.HasComponent(components.Player) {
				physics := components.Physics.Get(e)
				entityColor = color.RGBA{0, 255, 60, 255}
				if physics.OnGround == nil {
					entityColor = color.RGBA{200, 0, 200, 255}
				}
			} else if e.HasComponent(components.Enemy) {
				physics := components.Physics.Get(e)
				entityColor = color.RGBA{255, 60, 60, 255}
				if physics.OnGround == nil {
					entityColor = color.RGBA{255, 0, 255, 255}
				}
			}
			// This debug draw doesn't need to be camera-aware, as it's for debugging.
			vector.DrawFilledRect(screen, float32(o.X), float32(o.Y), float32(o.W), float32(o.H), entityColor, false)
		}
	})
}
