package components

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type SpriteData struct {
	Image    *ebiten.Image
	Rotation float64
	PivotX   float64
	PivotY   float64
}

var Sprite = donburi.NewComponentType[SpriteData]()
