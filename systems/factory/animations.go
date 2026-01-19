package factory

import (
	"fmt"
	"image"

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
		CachedFrames: make(map[cfg.StateID]map[int]*ebiten.Image),
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

		// Pre-calculate frames
		frames := make(map[int]*ebiten.Image)
		step := def.Step
		if step <= 0 {
			step = 1
		}

		for sheetIndex := def.First; sheetIndex <= def.Last; sheetIndex += step {
			sx := sheetIndex * frameWidth
			sy := 0
			srcRect := image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)
			// Use the global frame cache to avoid creating duplicate images
			frames[sheetIndex] = assets.GetFrame(key, state, sheetIndex, srcRect)
		}
		animData.CachedFrames[state] = frames
	}

	return animData
}
