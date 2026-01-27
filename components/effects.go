package components

import "github.com/yohamta/donburi"

// ScreenShakeData tracks active screen shake effect on the camera
type ScreenShakeData struct {
	Intensity float64 // max offset in pixels
	Duration  int     // frames remaining
	Elapsed   int     // frames elapsed (for oscillation)
}

var ScreenShake = donburi.NewComponentType[ScreenShakeData]()

// FlashData tracks sprite flash effect (hit flash, damage flash)
type FlashData struct {
	Duration int     // frames remaining
	R, G, B  float32 // color multipliers (1,1,1 = white, 1,0.5,0.5 = red tint)
}

var Flash = donburi.NewComponentType[FlashData]()

// SquashStretchData tracks sprite scale deformation for jump/land feel
type SquashStretchData struct {
	ScaleX, ScaleY   float64 // current scale
	TargetX, TargetY float64 // lerp target (usually 1.0, 1.0)
	LerpSpeed        float64 // how fast to return to normal
}

var SquashStretch = donburi.NewComponentType[SquashStretchData]()

// AutoDestroyData marks entities that should be destroyed after a duration or animation
type AutoDestroyData struct {
	FramesRemaining   int  // frames until destruction (-1 = use animation)
	DestroyOnAnimLoop bool // destroy when animation loops
}

var AutoDestroy = donburi.NewComponentType[AutoDestroyData]()

// VFXScaleData stores a fixed scale for VFX entities (doesn't lerp like SquashStretch)
type VFXScaleData struct {
	Scale float64 // uniform scale (1.0 = normal size)
}

var VFXScale = donburi.NewComponentType[VFXScaleData]()
