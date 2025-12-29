package factory

import (
	"fmt"

	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/assets/animations"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
)

// GenerateAnimations creates an AnimationData component based on the character key
// (e.g., "player", "guard") which maps to a set of animation definitions in config.
func GenerateAnimations(key string, frameWidth, frameHeight int) *components.AnimationData {
	// Get definitions for this key
	defs, ok := cfg.CharacterAnimations[key]
	if !ok {
		// Fallback or panic? For now, let's panic to catch configuration errors early.
		panic(fmt.Sprintf("No animation definitions found for key: %s", key))
	}

	animData := &components.AnimationData{
		SpriteSheets: make(map[cfg.StateID]*ebiten.Image),
		Animations:   make(map[cfg.StateID]*animations.Animation),
		FrameWidth:   frameWidth,
		FrameHeight:  frameHeight,
		CurrentSheet: cfg.Idle, // Default state
	}

	for state, def := range defs {
		// Load sprite from: assets/images/spritesheets/<key>/<state>.png
		// We assume the key corresponds to the directory name in assets/images/spritesheets/
		sprite := assets.GetSheet(key, state)
		animData.SpriteSheets[state] = sprite

		// Create Animation Object
		animData.Animations[state] = animations.NewAnimation(def.First, def.Last, def.Step, def.Speed)
	}

	return animData
}
