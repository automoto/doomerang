package factory

import (
	"image"
	"math"

	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/assets/animations"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// VFX frame dimensions
var vfxFrameSizes = map[cfg.StateID]struct{ W, H int }{
	// Dust effects (96x84 like player sprites)
	cfg.StateJumpDust:  {96, 84},
	cfg.StateLandDust:  {96, 84},
	cfg.StateSlideDust: {96, 84},
	// SFX effects
	cfg.StateExplosionShort: {57, 56},
	cfg.StatePlasma:         {32, 43},
	cfg.StateGunshot:        {46, 26},
	cfg.HitExplosion:        {47, 57},
	cfg.ChargeUp:            {102, 135},
}

// VFX directories
var vfxDirs = map[cfg.StateID]string{
	cfg.StateJumpDust:       "player",
	cfg.StateLandDust:       "player",
	cfg.StateSlideDust:      "player",
	cfg.StateExplosionShort: "sfx",
	cfg.StatePlasma:         "sfx",
	cfg.StateGunshot:        "sfx",
	cfg.HitExplosion:        "sfx",
	cfg.ChargeUp:            "sfx",
}

// SpawnVFX creates a visual effect entity at the given position (bottom-center anchored, for dust effects)
func SpawnVFX(ecs *ecs.ECS, x, y float64, effectType cfg.StateID) {
	spawnVFXInternal(ecs, x, y, effectType, false)
}

// SpawnVFXCentered creates a visual effect centered at the given position (for impact effects)
func SpawnVFXCentered(ecs *ecs.ECS, x, y float64, effectType cfg.StateID) {
	spawnVFXInternalScaled(ecs, x, y, effectType, true, 1.0)
}

// SpawnVFXCenteredScaled creates a scaled visual effect centered at the given position
func SpawnVFXCenteredScaled(ecs *ecs.ECS, x, y float64, effectType cfg.StateID, scale float64) {
	spawnVFXInternalScaled(ecs, x, y, effectType, true, scale)
}

// spawnVFXInternal handles VFX spawning with configurable centering (no scaling)
func spawnVFXInternal(ecs *ecs.ECS, x, y float64, effectType cfg.StateID, centered bool) {
	spawnVFXInternalScaled(ecs, x, y, effectType, centered, 1.0)
}

// spawnVFXInternalScaled handles VFX spawning with configurable centering and scaling
func spawnVFXInternalScaled(ecs *ecs.ECS, x, y float64, effectType cfg.StateID, centered bool, scale float64) {
	size, ok := vfxFrameSizes[effectType]
	if !ok {
		return // Unknown effect type
	}

	dir := vfxDirs[effectType]

	// Create entity (with VFXScale component if scale != 1.0)
	var entry *donburi.Entry
	if scale != 1.0 {
		entry = archetypes.VFXEffect.Spawn(ecs, components.VFXScale)
		components.VFXScale.Set(entry, &components.VFXScaleData{Scale: scale})
	} else {
		entry = archetypes.VFXEffect.Spawn(ecs)
	}

	// Get space for physics object
	spaceEntry, ok := components.Space.First(ecs.World)
	if !ok {
		return
	}
	space := components.Space.Get(spaceEntry)

	// Create physics object (non-colliding, just for position)
	// Centered: true center at (x, y) for impact effects
	// Not centered: bottom-center at (x, y) for dust effects at feet
	var objY float64
	if centered {
		objY = y - float64(size.H)/2
	} else {
		objY = y - float64(size.H)
	}
	obj := resolv.NewObject(x-float64(size.W)/2, objY, float64(size.W), float64(size.H))
	obj.Data = entry
	space.Add(obj)
	components.Object.Set(entry, &components.ObjectData{Object: obj})

	// Create animation data for this effect
	animData := createVFXAnimation(dir, effectType, size.W, size.H)
	components.Animation.Set(entry, animData)
	animData.SetAnimation(effectType)

	// Set auto-destroy on animation loop
	components.AutoDestroy.Set(entry, &components.AutoDestroyData{
		FramesRemaining:   -1,
		DestroyOnAnimLoop: true,
	})
}

// SpawnVFXWithRotation creates a visual effect with rotation (for directional effects like gunshot)
func SpawnVFXWithRotation(ecs *ecs.ECS, x, y, rotation float64, effectType cfg.StateID) {
	size, ok := vfxFrameSizes[effectType]
	if !ok {
		return
	}

	dir := vfxDirs[effectType]

	entry := archetypes.VFXEffect.Spawn(ecs)

	spaceEntry, ok := components.Space.First(ecs.World)
	if !ok {
		return
	}
	space := components.Space.Get(spaceEntry)

	obj := resolv.NewObject(x-float64(size.W)/2, y-float64(size.H)/2, float64(size.W), float64(size.H))
	obj.Data = entry
	space.Add(obj)
	components.Object.Set(entry, &components.ObjectData{Object: obj})

	animData := createVFXAnimation(dir, effectType, size.W, size.H)
	components.Animation.Set(entry, animData)
	animData.SetAnimation(effectType)

	// Add sprite component for rotation support
	// Note: VFX with rotation use Sprite component instead of Animation rendering
	entry.AddComponent(components.Sprite)
	// Get first frame as sprite
	if frames, ok := animData.CachedFrames[effectType]; ok {
		if img, ok := frames[0]; ok {
			components.Sprite.Set(entry, &components.SpriteData{
				Image:    img,
				Rotation: rotation,
				PivotX:   float64(size.W) / 2,
				PivotY:   float64(size.H) / 2,
			})
		}
	}

	components.AutoDestroy.Set(entry, &components.AutoDestroyData{
		FramesRemaining:   -1,
		DestroyOnAnimLoop: true,
	})
}

// createVFXAnimation creates animation data for a VFX effect
func createVFXAnimation(dir string, effectType cfg.StateID, frameWidth, frameHeight int) *components.AnimationData {
	// Get animation definition
	var defs map[cfg.StateID]cfg.AnimationDef
	if dir == "player" {
		defs = cfg.CharacterAnimations["player"]
	} else {
		defs = cfg.CharacterAnimations["sfx"]
	}

	def, ok := defs[effectType]
	if !ok {
		return nil
	}

	animData := &components.AnimationData{
		SpriteSheets: make(map[cfg.StateID]*ebiten.Image),
		Animations:   make(map[cfg.StateID]*animations.Animation),
		CachedFrames: make(map[cfg.StateID]map[int]*ebiten.Image),
		FrameWidth:   frameWidth,
		FrameHeight:  frameHeight,
		CurrentSheet: effectType,
	}

	// Load sprite sheet
	sprite := assets.GetSheet(dir, effectType)
	animData.SpriteSheets[effectType] = sprite

	// Create animation
	anim := animations.NewAnimation(def.First, def.Last, def.Step, def.Speed)
	animData.Animations[effectType] = anim

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
	animData.CachedFrames[effectType] = frames

	return animData
}

// Helper functions for common effect spawning

// SpawnJumpDust spawns jump dust at player's feet
func SpawnJumpDust(ecs *ecs.ECS, x, y float64) {
	SpawnVFX(ecs, x, y, cfg.StateJumpDust)
}

// SpawnLandDust spawns landing dust at player's feet
func SpawnLandDust(ecs *ecs.ECS, x, y float64) {
	SpawnVFX(ecs, x, y, cfg.StateLandDust)
}

// SpawnSlideDust spawns slide dust at player's feet
func SpawnSlideDust(ecs *ecs.ECS, x, y float64) {
	SpawnVFX(ecs, x, y, cfg.StateSlideDust)
}

// SpawnExplosion spawns explosion effect centered at position with optional scale
// scale: 1.0 = full size, 0.5 = half size, etc.
func SpawnExplosion(ecs *ecs.ECS, x, y, scale float64) {
	SpawnVFXCenteredScaled(ecs, x, y, cfg.StateExplosionShort, scale)
}

// SpawnPlasma spawns plasma effect centered at position
func SpawnPlasma(ecs *ecs.ECS, x, y float64) {
	SpawnVFXCentered(ecs, x, y, cfg.StatePlasma)
}

// SpawnGunshot spawns gunshot effect with direction
func SpawnGunshot(ecs *ecs.ECS, x, y, directionX float64) {
	rotation := 0.0
	if directionX < 0 {
		rotation = math.Pi // Flip 180 degrees if facing left
	}
	SpawnVFXWithRotation(ecs, x, y, rotation, cfg.StateGunshot)
}

// SpawnHitExplosion spawns scaled hit explosion effect centered at position
func SpawnHitExplosion(ecs *ecs.ECS, x, y, scale float64) {
	SpawnVFXCenteredScaled(ecs, x, y, cfg.HitExplosion, scale)
}

// SpawnChargeVFX spawns a looping charge-up VFX at the player's feet and returns the entry
func SpawnChargeVFX(ecs *ecs.ECS, x, y float64) *donburi.Entry {
	size := vfxFrameSizes[cfg.ChargeUp]
	dir := vfxDirs[cfg.ChargeUp]

	entry := archetypes.VFXEffect.Spawn(ecs)

	spaceEntry, ok := components.Space.First(ecs.World)
	if !ok {
		return nil
	}
	space := components.Space.Get(spaceEntry)

	// Position at bottom-center: x,y is player's feet, so center horizontally and extend upward
	obj := resolv.NewObject(x-float64(size.W)/2, y-float64(size.H), float64(size.W), float64(size.H))
	obj.Data = entry
	space.Add(obj)
	components.Object.Set(entry, &components.ObjectData{Object: obj})

	animData := createVFXAnimation(dir, cfg.ChargeUp, size.W, size.H)
	components.Animation.Set(entry, animData)
	animData.SetAnimation(cfg.ChargeUp)

	// Don't auto-destroy - this VFX loops until manually destroyed
	components.AutoDestroy.Set(entry, &components.AutoDestroyData{
		FramesRemaining:   -1,
		DestroyOnAnimLoop: false,
	})

	return entry
}

// DestroyChargeVFX removes a charge VFX entity
func DestroyChargeVFX(ecs *ecs.ECS, entry *donburi.Entry) {
	if entry == nil || !entry.Valid() {
		return
	}
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		obj := components.Object.Get(entry)
		if obj != nil && obj.Object != nil {
			components.Space.Get(spaceEntry).Remove(obj.Object)
		}
	}
	ecs.World.Remove(entry.Entity())
}

// UpdateChargeVFXPosition updates the position of a charge VFX to follow a target
func UpdateChargeVFXPosition(entry *donburi.Entry, x, y float64) {
	if entry == nil || !entry.Valid() {
		return
	}
	obj := components.Object.Get(entry)
	if obj != nil && obj.Object != nil {
		size := vfxFrameSizes[cfg.ChargeUp]
		// Position at bottom-center: x,y is player's feet
		obj.X = x - float64(size.W)/2
		obj.Y = y - float64(size.H)
	}
}
