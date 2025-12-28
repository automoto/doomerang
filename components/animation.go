package components

import (
	"github.com/automoto/doomerang/assets/animations"
	"github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
)

type AnimationData struct {
	CurrentAnimation *animations.Animation
	SpriteSheets     map[config.StateID]*ebiten.Image
	CurrentSheet     config.StateID
	FrameWidth       int
	FrameHeight      int
	Animations       map[config.StateID]*animations.Animation
}

func (a *AnimationData) SetAnimation(state config.StateID) {
	if anim, ok := a.Animations[state]; ok {
		if a.CurrentAnimation != anim {
			a.CurrentAnimation = anim
			a.CurrentSheet = state
			a.CurrentAnimation.Restart()
			a.CurrentAnimation.Looped = false
		}
	}
}

var Animation = donburi.NewComponentType[AnimationData]()
