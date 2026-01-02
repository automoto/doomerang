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

// Viewport culling significantly improves performance by skipping the expensive matrix
// calculations and draw calls for entities that are currently off-screen.
// A small padding is used to prevent sprites from popping in/out at the edges.

// DrawAnimated renders entities with an Animation component based on their current frame and state.
func DrawAnimated(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return // No camera yet
	}
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Culling bounds
	padding := 64.0
	minX := camera.Position.X - float64(width)/2 - padding
	maxX := camera.Position.X + float64(width)/2 + padding
	minY := camera.Position.Y - float64(height)/2 - padding
	maxY := camera.Position.Y + float64(height)/2 + padding

	components.Animation.Each(ecs.World, func(e *donburi.Entry) {
		o := components.Object.Get(e)
		isPlayer := e.HasComponent(components.Player)
		isEnemy := e.HasComponent(components.Enemy)

		// Viewport Culling
		if o.X+o.W < minX || o.X > maxX || o.Y+o.H < minY || o.Y > maxY {
			return
		}

		animData := components.Animation.Get(e)

		if animData.CurrentAnimation != nil {
			// Get the current frame index (sheet index)
			frame := animData.CurrentAnimation.Frame()

			var img *ebiten.Image
			if frames, ok := animData.CachedFrames[animData.CurrentSheet]; ok {
				img = frames[frame]
			}

			// Fallback to runtime slicing if not cached (safety)
			if img == nil && animData.SpriteSheets[animData.CurrentSheet] != nil {
				sx := frame * animData.FrameWidth
				sy := 0
				srcRect := image.Rect(sx, sy, sx+animData.FrameWidth, sy+animData.FrameHeight)
				img = animData.SpriteSheets[animData.CurrentSheet].SubImage(srcRect).(*ebiten.Image)
			}

			if img != nil {
				// Reset draw options.
				drawOp.GeoM.Reset()
				drawOp.ColorScale.Reset()

				// Anchor the sprite at its bottom-center so that the feet line up with the
				// bottom of the collision box.
				drawOp.GeoM.Translate(-float64(animData.FrameWidth)/2, -float64(animData.FrameHeight))

				// Flip the sprite if facing left.
				if isPlayer {
					player := components.Player.Get(e)
					if player.Direction.X < 0 {
						drawOp.GeoM.Scale(-1, 1)
					}
				}
				if isEnemy {
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
				if isEnemy {
					enemy := components.Enemy.Get(e)
					if enemy.InvulnFrames > 0 && enemy.InvulnFrames%4 < 2 {
						drawOp.ColorScale.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
					}
				}
				if isPlayer {
					player := components.Player.Get(e)
					if player.InvulnFrames > 0 && player.InvulnFrames%4 < 2 {
						drawOp.ColorScale.Scale(1, 0.5, 0.5, 0.8) // Red tint and semi-transparent
					}
				}

				// Apply enemy type color tinting
				if isEnemy {
					enemy := components.Enemy.Get(e)
					drawOp.ColorScale.ScaleWithColorScale(enemy.TintColor)
				}

				// Draw the current frame.
				screen.DrawImage(img, drawOp)
			}
		} else {
			// Fallback to rectangle if no animation is available
			var entityColor color.Color
			if isPlayer {
				physics := components.Physics.Get(e)
				entityColor = cfg.Blue
				if physics.OnGround == nil {
					entityColor = cfg.Purple
				}
			} else if isEnemy {
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
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return // No camera yet
	}
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Culling bounds
	padding := 64.0
	minX := camera.Position.X - float64(width)/2 - padding
	maxX := camera.Position.X + float64(width)/2 + padding
	minY := camera.Position.Y - float64(height)/2 - padding
	maxY := camera.Position.Y + float64(height)/2 + padding

	// Iterate over entities with Health and HealthBar components
	components.HealthBar.Each(ecs.World, func(e *donburi.Entry) {
		if !e.HasComponent(components.Health) {
			return
		}
		o := components.Object.Get(e)

		// Viewport Culling
		if o.X+o.W < minX || o.X > maxX || o.Y+o.H < minY || o.Y > maxY {
			return
		}

		hp := components.Health.Get(e)

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
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return // No camera yet
	}
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Culling bounds
	padding := 64.0
	minX := camera.Position.X - float64(width)/2 - padding
	maxX := camera.Position.X + float64(width)/2 + padding
	minY := camera.Position.Y - float64(height)/2 - padding
	maxY := camera.Position.Y + float64(height)/2 + padding

	components.Sprite.Each(ecs.World, func(e *donburi.Entry) {
		o := components.Object.Get(e)

		// Viewport Culling
		if o.X+o.W < minX || o.X > maxX || o.Y+o.H < minY || o.Y > maxY {
			return
		}

		sprite := components.Sprite.Get(e)

		drawOp.GeoM.Reset()
		drawOp.ColorScale.Reset()

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
