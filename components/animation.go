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
	CachedFrames     map[config.StateID]map[int]*ebiten.Image // Pre-calculated subimages keyed by sheet index
	CurrentSheet     config.StateID
	FrameWidth       int
	FrameHeight      int
	Animations       map[config.StateID]*animations.Animation
}

func (a *AnimationData) SetAnimation(state config.StateID) {
	if a.CurrentSheet == state && (a.CurrentAnimation != nil || a.Animations[state] == nil) {
		return
	}

	anim, ok := a.Animations[state]
	if ok {
		if a.CurrentAnimation != anim {
			a.CurrentAnimation = anim
			a.CurrentSheet = state
			a.CurrentAnimation.Restart()
			a.CurrentAnimation.Looped = false
		}
	} else {
		// No animation for this state, clear current
		a.CurrentAnimation = nil
		a.CurrentSheet = state
	}
}

var Animation = donburi.NewComponentType[AnimationData]()
