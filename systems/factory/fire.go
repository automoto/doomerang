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
// x, y is the BASE EDGE where fire emanates from. Fire extends outward in the specified direction.
func CreateFire(ecs *ecs.ECS, x, y float64, fireType, direction string) *donburi.Entry {
	fire := archetypes.Fire.Spawn(ecs)

	// Get config for this fire type
	fireCfg, ok := cfg.Fire.Types[fireType]
	if !ok {
		// Default to continuous if unknown type
		fireCfg = cfg.Fire.Types["fire_continuous"]
	}

	// Hitbox scale factor (default to 1.0 if not set)
	hitboxScale := fireCfg.HitboxScale
	if hitboxScale == 0 {
		hitboxScale = 1.0
	}

	// Hitbox dimensions (swap width/height for vertical directions)
	hitboxW := float64(fireCfg.FrameWidth) * hitboxScale
	hitboxH := float64(fireCfg.FrameHeight) * hitboxScale
	if direction == "up" || direction == "down" {
		hitboxW, hitboxH = hitboxH, hitboxW
	}

	// Calculate sprite center and hitbox position
	frameW := float64(fireCfg.FrameWidth)
	spriteCenterX, spriteCenterY := components.FireSpriteCenter(x, y, frameW, direction)
	hitboxX := spriteCenterX - hitboxW/2
	hitboxY := spriteCenterY - hitboxH/2

	// Create resolv collision object
	obj := resolv.NewObject(hitboxX, hitboxY, hitboxW, hitboxH, tags.ResolvFire)
	obj.SetShape(resolv.NewRectangle(0, 0, hitboxW, hitboxH))
	obj.Data = fire

	components.Object.SetValue(fire, components.ObjectData{Object: obj})

	components.Fire.SetValue(fire, components.FireData{
		FireType:       fireType,
		Active:         true,
		Damage:         fireCfg.Damage,
		KnockbackForce: fireCfg.KnockbackForce,
		Direction:      direction,
		BaseWidth:      hitboxW,
		BaseHeight:     hitboxH,
		OriginX:        x,
		OriginY:        y,
		FrameWidth:     frameW,
		SpriteCenterX:  spriteCenterX,
		SpriteCenterY:  spriteCenterY,
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

