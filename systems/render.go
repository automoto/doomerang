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

var (
	drawOp = &ebiten.DrawImageOptions{}
)

func DrawAnimated(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	components.Animation.Each(ecs.World, func(e *donburi.Entry) {
		o := components.Object.Get(e)
		animData := components.Animation.Get(e)

		if animData.CurrentAnimation != nil && animData.SpriteSheets[animData.CurrentSheet] != nil {
			// Calculate the source rectangle for the current frame.
			frame := animData.CurrentAnimation.Frame()
			sx := frame * animData.FrameWidth
			sy := 0
			srcRect := image.Rect(sx, sy, sx+animData.FrameWidth, sy+animData.FrameHeight)

			// Reset draw options.
			drawOp.GeoM.Reset()
			drawOp.ColorM.Reset()

			// Anchor the sprite at its bottom-center so that the feet line up with the
			// bottom of the collision box.
			drawOp.GeoM.Translate(-float64(animData.FrameWidth)/2, -float64(animData.FrameHeight))

			// Flip the sprite if facing left.
			if e.HasComponent(components.Player) {
				player := components.Player.Get(e)
				if player.Direction.X < 0 {
					drawOp.GeoM.Scale(-1, 1)
				}
			}
			if e.HasComponent(components.Enemy) {
				enemy := components.Enemy.Get(e)
				if enemy.Direction.X < 0 {
					drawOp.GeoM.Scale(-1, 1)
				}
			}

			// Move the sprite so that its bottom-center aligns with the bottom-center
			// of the (smaller) collision box.
			drawOp.GeoM.Translate(o.X+o.W/2, o.Y+o.H)

			// Apply the camera translation.
			drawOp.GeoM.Translate(float64(width)/2-camera.Position.X, float64(height)/2-camera.Position.Y)

			// Flicker effect if invulnerable
			if e.HasComponent(components.Enemy) {
				enemy := components.Enemy.Get(e)
				if enemy.InvulnFrames > 0 && enemy.InvulnFrames%4 < 2 {
					drawOp.ColorM.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
				}
			}
			if e.HasComponent(components.Player) {
				player := components.Player.Get(e)
				if player.InvulnFrames > 0 && player.InvulnFrames%4 < 2 {
					drawOp.ColorM.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
				}
			}

			// Apply enemy type color tinting
			if e.HasComponent(components.Enemy) {
				enemy := components.Enemy.Get(e)
				enemyType, exists := cfg.Enemy.Types[enemy.TypeName]
				// Check if tint is not white (default)
				if exists && (enemyType.TintColor.R != 255 || enemyType.TintColor.G != 255 || enemyType.TintColor.B != 255 || enemyType.TintColor.A != 255) {
					// Apply color tint (normalize RGBA values to 0-1 range)
					r := float64(enemyType.TintColor.R) / 255.0
					g := float64(enemyType.TintColor.G) / 255.0
					b := float64(enemyType.TintColor.B) / 255.0
					a := float64(enemyType.TintColor.A) / 255.0
					drawOp.ColorM.Scale(r, g, b, a)
				}
			}

			// Draw the current frame.
			screen.DrawImage(animData.SpriteSheets[animData.CurrentSheet].SubImage(srcRect).(*ebiten.Image), drawOp)
		} else {
			// Fallback to rectangle if no animation is available
			var entityColor color.Color
			if e.HasComponent(components.Player) {
				physics := components.Physics.Get(e)
				entityColor = cfg.Blue
				if physics.OnGround == nil {
					entityColor = cfg.Purple
				}
			} else if e.HasComponent(components.Enemy) {
				physics := components.Physics.Get(e)
				entityColor = cfg.LightRed
				if physics.OnGround == nil {
					entityColor = cfg.Magenta
				}
			}

			// Calculate screen position for debug rect
			screenX := float64(width)/2 - camera.Position.X + o.X
			screenY := float64(height)/2 - camera.Position.Y + o.Y

			// This debug draw doesn't need to be camera-aware, as it's for debugging.
			vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(o.W), float32(o.H), entityColor, false)
		}
	})
}

func DrawHealthBars(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera for position translation
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Iterate over entities with Health and HealthBar components
	components.HealthBar.Each(ecs.World, func(e *donburi.Entry) {
		if !e.HasComponent(components.Health) {
			return
		}
		hp := components.Health.Get(e)
		o := components.Object.Get(e)

		// Health bar dimensions and position
		barWidth := 32.0
		barHeight := 4.0
		// Position the bar above the entity's collision box
		barX := o.X + (o.W-barWidth)/2
		barY := o.Y - barHeight - 4 // 4 pixels of padding

		// Calculate health percentage
		healthPercentage := float64(hp.Current) / float64(hp.Max)

		// Apply camera translation
		drawX := barX + float64(width)/2 - camera.Position.X
		drawY := barY + float64(height)/2 - camera.Position.Y

		// Draw the background of the health bar (red)
		vector.DrawFilledRect(screen, float32(drawX), float32(drawY), float32(barWidth), float32(barHeight), cfg.Red, false)

		// Draw the foreground of the health bar (green)
		vector.DrawFilledRect(screen, float32(drawX), float32(drawY), float32(barWidth*healthPercentage), float32(barHeight), cfg.Green, false)
	})
}

func DrawSprites(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	components.Sprite.Each(ecs.World, func(e *donburi.Entry) {
		sprite := components.Sprite.Get(e)
		o := components.Object.Get(e)

		drawOp.GeoM.Reset()
		drawOp.ColorM.Reset()

		// Translate to pivot (center of sprite)
		drawOp.GeoM.Translate(-sprite.PivotX, -sprite.PivotY)

		// Rotate
		drawOp.GeoM.Rotate(sprite.Rotation)

		// Translate to object position + center offset
		// Assuming o.X, o.Y is top-left of hitbox.
		// We want to draw sprite centered on hitbox center.
		centerX := o.X + o.W/2
		centerY := o.Y + o.H/2
		drawOp.GeoM.Translate(centerX, centerY)

		// Camera
		drawOp.GeoM.Translate(float64(width)/2-camera.Position.X, float64(height)/2-camera.Position.Y)

		screen.DrawImage(sprite.Image, drawOp)
	})
}
