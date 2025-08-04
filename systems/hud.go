package systems

import (
	"image/color"

	"github.com/automoto/doomerang/components"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

const (
	hudBarWidth  = 200
	hudBarHeight = 20
	hudMargin    = 10
)

// DrawHUD renders the player's health bar in the top-left corner.
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

	// Current HP (red)
	ratio := float32(hp.Current) / float32(hp.Max)
	vector.DrawFilledRect(screen,
		float32(hudMargin), float32(hudMargin),
		float32(hudBarWidth)*ratio, float32(hudBarHeight),
		color.RGBA{220, 40, 40, 255}, false)
}
