package systems

import (
	"image"

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

// renderContext holds camera and culling bounds for render functions.
type renderContext struct {
	camera        *components.CameraData
	screenWidth   int
	screenHeight  int
	minX, maxX    float64
	minY, maxY    float64
}

const cullPadding = 64.0

func getRenderContext(ecs *ecs.ECS, screen *ebiten.Image) (renderContext, bool) {
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return renderContext{}, false
	}

	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()
	halfW, halfH := float64(width)/2, float64(height)/2

	return renderContext{
		camera:       camera,
		screenWidth:  width,
		screenHeight: height,
		minX:         camera.Position.X - halfW - cullPadding,
		maxX:         camera.Position.X + halfW + cullPadding,
		minY:         camera.Position.Y - halfH - cullPadding,
		maxY:         camera.Position.Y + halfH + cullPadding,
	}, true
}

func (rc *renderContext) isOutsideViewport(o *components.ObjectData) bool {
	return o.X+o.W < rc.minX || o.X > rc.maxX || o.Y+o.H < rc.minY || o.Y > rc.maxY
}

func (rc *renderContext) cameraOffsetX() float64 {
	return float64(rc.screenWidth)/2 - rc.camera.Position.X
}

func (rc *renderContext) cameraOffsetY() float64 {
	return float64(rc.screenHeight)/2 - rc.camera.Position.Y
}

// applyDirectionFlip flips the sprite horizontally if facing left.
func applyDirectionFlip(directionX float64) {
	if directionX < 0 {
		drawOp.GeoM.Scale(-1, 1)
	}
}

// applyInvulnFlicker applies a red tint flicker effect when invulnerable.
func applyInvulnFlicker(invulnFrames int) {
	if invulnFrames > 0 && invulnFrames%4 < 2 {
		drawOp.ColorScale.Scale(1, 0.5, 0.5, 0.8)
	}
}

// DrawAnimated renders entities with an Animation component based on their current frame and state.
func DrawAnimated(ecs *ecs.ECS, screen *ebiten.Image) {
	rc, ok := getRenderContext(ecs, screen)
	if !ok {
		return
	}

	components.Animation.Each(ecs.World, func(e *donburi.Entry) {
		o := components.Object.Get(e)
		if rc.isOutsideViewport(o) {
			return
		}

		animData := components.Animation.Get(e)
		if animData.CurrentAnimation == nil {
			drawAnimatedFallback(e, o, &rc, screen)
			return
		}

		img := getAnimationFrame(animData)
		if img == nil {
			return
		}

		drawAnimatedSprite(e, o, animData, img, &rc, screen)
	})
}

func getAnimationFrame(animData *components.AnimationData) *ebiten.Image {
	frame := animData.CurrentAnimation.Frame()

	if frames, ok := animData.CachedFrames[animData.CurrentSheet]; ok {
		if img := frames[frame]; img != nil {
			return img
		}
	}

	// Fallback to runtime slicing if not cached (safety)
	sheet := animData.SpriteSheets[animData.CurrentSheet]
	if sheet == nil {
		return nil
	}

	sx := frame * animData.FrameWidth
	srcRect := image.Rect(sx, 0, sx+animData.FrameWidth, animData.FrameHeight)
	return sheet.SubImage(srcRect).(*ebiten.Image)
}

func drawAnimatedSprite(e *donburi.Entry, o *components.ObjectData, animData *components.AnimationData, img *ebiten.Image, rc *renderContext, screen *ebiten.Image) {
	drawOp.GeoM.Reset()
	drawOp.ColorScale.Reset()

	// Anchor the sprite at its bottom-center so that the feet line up with the collision box.
	drawOp.GeoM.Translate(-float64(animData.FrameWidth)/2, -float64(animData.FrameHeight))

	// Flip sprite and apply effects based on entity type
	if e.HasComponent(components.Player) {
		player := components.Player.Get(e)
		applyDirectionFlip(player.Direction.X)
		applyInvulnFlicker(player.InvulnFrames)
	} else if e.HasComponent(components.Enemy) {
		enemy := components.Enemy.Get(e)
		applyDirectionFlip(enemy.Direction.X)
		applyInvulnFlicker(enemy.InvulnFrames)
		drawOp.ColorScale.ScaleWithColorScale(enemy.TintColor)
	}

	// Position sprite at bottom-center of collision box, then apply camera offset
	drawOp.GeoM.Translate(o.X+o.W/2, o.Y+o.H)
	drawOp.GeoM.Translate(rc.cameraOffsetX(), rc.cameraOffsetY())

	screen.DrawImage(img, drawOp)
}

func drawAnimatedFallback(e *donburi.Entry, o *components.ObjectData, rc *renderContext, screen *ebiten.Image) {
	entityColor := cfg.Debug.EntityDefault

	if e.HasComponent(components.Player) {
		entityColor = cfg.Blue
		if physics := components.Physics.Get(e); physics.OnGround == nil {
			entityColor = cfg.Purple
		}
	} else if e.HasComponent(components.Enemy) {
		entityColor = cfg.LightRed
		if physics := components.Physics.Get(e); physics.OnGround == nil {
			entityColor = cfg.Magenta
		}
	}

	screenX := rc.cameraOffsetX() + o.X
	screenY := rc.cameraOffsetY() + o.Y
	vector.DrawFilledRect(screen, float32(screenX), float32(screenY), float32(o.W), float32(o.H), entityColor, false)
}

const (
	healthBarWidth   = 32.0
	healthBarHeight  = 4.0
	healthBarPadding = 4.0
)

func DrawHealthBars(ecs *ecs.ECS, screen *ebiten.Image) {
	rc, ok := getRenderContext(ecs, screen)
	if !ok {
		return
	}

	components.HealthBar.Each(ecs.World, func(e *donburi.Entry) {
		if !e.HasComponent(components.Health) {
			return
		}

		o := components.Object.Get(e)
		if rc.isOutsideViewport(o) {
			return
		}

		hp := components.Health.Get(e)
		healthPercentage := float64(hp.Current) / float64(hp.Max)

		// Position the bar above the entity's collision box
		barX := o.X + (o.W-healthBarWidth)/2
		barY := o.Y - healthBarHeight - healthBarPadding

		drawX := float32(barX + rc.cameraOffsetX())
		drawY := float32(barY + rc.cameraOffsetY())

		// Draw background (red) and foreground (green)
		vector.DrawFilledRect(screen, drawX, drawY, healthBarWidth, healthBarHeight, cfg.Red, false)
		vector.DrawFilledRect(screen, drawX, drawY, float32(healthBarWidth*healthPercentage), healthBarHeight, cfg.Green, false)
	})
}

func DrawSprites(ecs *ecs.ECS, screen *ebiten.Image) {
	rc, ok := getRenderContext(ecs, screen)
	if !ok {
		return
	}

	components.Sprite.Each(ecs.World, func(e *donburi.Entry) {
		o := components.Object.Get(e)
		if rc.isOutsideViewport(o) {
			return
		}

		sprite := components.Sprite.Get(e)

		drawOp.GeoM.Reset()
		drawOp.ColorScale.Reset()

		// Translate to pivot, rotate, then position at hitbox center
		drawOp.GeoM.Translate(-sprite.PivotX, -sprite.PivotY)
		drawOp.GeoM.Rotate(sprite.Rotation)
		drawOp.GeoM.Translate(o.X+o.W/2, o.Y+o.H/2)
		drawOp.GeoM.Translate(rc.cameraOffsetX(), rc.cameraOffsetY())

		screen.DrawImage(sprite.Image, drawOp)
	})
}
