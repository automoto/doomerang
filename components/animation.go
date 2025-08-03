package components

import (
	"github.com/automoto/doomerang/assets/animations"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type AnimationData struct {
	CurrentAnimation *animations.Animation
	SpriteSheets     map[string]*ebiten.Image
	CurrentSheet     string
	FrameWidth       int
	FrameHeight      int
	Animations       map[string]*animations.Animation
}

func (a *AnimationData) SetAnimation(name string) {
	if anim, ok := a.Animations[name]; ok {
		if a.CurrentAnimation != anim {
			a.CurrentAnimation = anim
			a.CurrentSheet = name
			a.CurrentAnimation.Restart()
		}
	}
}

var Animation = donburi.NewComponentType[AnimationData]()
