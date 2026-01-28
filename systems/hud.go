package systems

import (
	"image/color"

	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const (
	hudBarWidth  = 130
	hudBarHeight = 13
	hudMargin    = 10
	livesMargin  = 5
)

var heartIcon *ebiten.Image
var hudDrawOp = &ebiten.DrawImageOptions{}

// DrawHUD renders the player's health bar and lives counter in the top-left corner.
func DrawHUD(ecs *ecs.ECS, screen *ebiten.Image) {
	playerEntry, ok := components.Player.First(ecs.World)
	if !ok {
		return
	}
	hp := components.Health.Get(playerEntry)

	// Background (dark gray)
	vector.DrawFilledRect(screen,
		float32(hudMargin), float32(hudMargin),
		float32(hudBarWidth), float32(hudBarHeight),
		color.RGBA{40, 40, 40, 255}, false)

	// Current HP (green)
	ratio := float32(hp.Current) / float32(hp.Max)
	vector.DrawFilledRect(screen,
		float32(hudMargin), float32(hudMargin),
		float32(hudBarWidth)*ratio, float32(hudBarHeight),
		color.RGBA{40, 220, 40, 255}, false)

	// Draw lives counter
	drawLives(playerEntry, screen)
}

func drawLives(playerEntry *donburi.Entry, screen *ebiten.Image) {
	lives := components.Lives.Get(playerEntry)

	// Lazy load heart icon
	if heartIcon == nil {
		heartIcon = assets.GetIconImage("icon_heart.png")
	}

	heartWidth := heartIcon.Bounds().Dx()
	livesY := float64(hudMargin + hudBarHeight + livesMargin)

	for i := 0; i < lives.Lives; i++ {
		hudDrawOp.GeoM.Reset()
		hudDrawOp.GeoM.Translate(float64(hudMargin)+float64(i)*float64(heartWidth+livesMargin), livesY)
		screen.DrawImage(heartIcon, hudDrawOp)
	}
}
