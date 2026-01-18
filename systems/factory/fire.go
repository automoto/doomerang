package factory

import (
	"image"

	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/assets/animations"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// CreateFire creates a fire obstacle entity with collision detection and animation
// x, y is the point where fire emanates FROM. Hitbox extends in the direction specified.
func CreateFire(ecs *ecs.ECS, x, y float64, fireType, direction string) *donburi.Entry {
	fire := archetypes.Fire.Spawn(ecs)

	// Get config for this fire type
	fireCfg, ok := cfg.Fire.Types[fireType]
	if !ok {
		// Default to continuous if unknown type
		fireCfg = cfg.Fire.Types["fire_continuous"]
	}

	// Use sprite dimensions from config for positioning
	// For up/down, swap width and height (rotated 90°)
	spriteW := float64(fireCfg.FrameWidth)
	spriteH := float64(fireCfg.FrameHeight)

	// Calculate hitbox position using full sprite dimensions
	// The point (x, y) marks where fire emanates FROM
	var hitboxX, hitboxY, w, h float64
	switch direction {
	case "left":
		w, h = spriteW, spriteH
		hitboxX, hitboxY = x-w, y-h/2
	case "up":
		w, h = spriteH, spriteW // Swap for vertical
		hitboxX, hitboxY = x-w/2, y-h
	case "down":
		w, h = spriteH, spriteW // Swap for vertical
		hitboxX, hitboxY = x-w/2, y
	default: // "right"
		w, h = spriteW, spriteH
		hitboxX, hitboxY = x, y-h/2
	}

	// Calculate sprite center (used for rendering)
	spriteCenterX := hitboxX + w/2
	spriteCenterY := hitboxY + h/2

	// Calculate anchor point for dynamic hitbox positioning
	// Anchor is the fixed point that doesn't move during scaling
	var anchorX, anchorY float64
	switch direction {
	case "left":
		anchorX, anchorY = hitboxX+w, hitboxY+h
	case "up":
		anchorX, anchorY = spriteCenterX, hitboxY+h
	case "down":
		anchorX, anchorY = spriteCenterX, hitboxY
	default: // "right"
		anchorX, anchorY = hitboxX, hitboxY+h
	}

	// Create resolv collision object
	obj := resolv.NewObject(hitboxX, hitboxY, w, h, tags.ResolvFire)
	obj.SetShape(resolv.NewRectangle(0, 0, w, h))
	obj.Data = fire

	components.Object.SetValue(fire, components.ObjectData{Object: obj})

	// Determine initial active state
	active := true // Both types start active

	components.Fire.SetValue(fire, components.FireData{
		FireType:       fireType,
		Active:         active,
		Damage:         fireCfg.Damage,
		KnockbackForce: fireCfg.KnockbackForce,
		Direction:      direction,
		BaseWidth:      w,
		BaseHeight:     h,
		SpriteCenterX:  spriteCenterX,
		SpriteCenterY:  spriteCenterY,
		AnchorX:        anchorX,
		AnchorY:        anchorY,
		HitboxPhases:   fireCfg.HitboxPhases,
	})

	// Set up animation
	animData := createFireAnimation(fireCfg.State, fireCfg.FrameWidth, fireCfg.FrameHeight)
	components.Animation.SetValue(fire, *animData)
	// Must get component after SetValue to call SetAnimation on the actual stored data
	components.Animation.Get(fire).SetAnimation(fireCfg.State)

	// Add to physics space
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		components.Space.Get(spaceEntry).Add(obj)
	}

	return fire
}

// createFireAnimation creates animation data for a fire obstacle
func createFireAnimation(state cfg.StateID, frameWidth, frameHeight int) *components.AnimationData {
	// Get animation definition
	defs := cfg.CharacterAnimations["obstacle"]
	def, ok := defs[state]
	if !ok {
		return nil
	}

	animData := &components.AnimationData{
		SpriteSheets: make(map[cfg.StateID]*ebiten.Image),
		Animations:   make(map[cfg.StateID]*animations.Animation),
		CachedFrames: make(map[cfg.StateID]map[int]*ebiten.Image),
		FrameWidth:   frameWidth,
		FrameHeight:  frameHeight,
		CurrentSheet: state,
	}

	// Load sprite sheet
	sprite := assets.GetSheet("obstacle", state)
	animData.SpriteSheets[state] = sprite

	// Create animation
	anim := animations.NewAnimation(def.First, def.Last, def.Step, def.Speed)
	animData.Animations[state] = anim

	// Pre-calculate frames
	frames := make(map[int]*ebiten.Image)
	step := def.Step
	if step <= 0 {
		step = 1
	}
	for i := def.First; i <= def.Last; i += step {
		sx := i * frameWidth
		sy := 0
		srcRect := image.Rect(sx, sy, sx+frameWidth, sy+frameHeight)
		frames[i] = sprite.SubImage(srcRect).(*ebiten.Image)
	}
	animData.CachedFrames[state] = frames

	return animData
}
