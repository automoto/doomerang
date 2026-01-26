package components

import (
	cfg "github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
)

type FireData struct {
	FireType       string  // "fire_pulsing" or "fire_continuous"
	Active         bool    // Currently dangerous?
	Damage         int     // Cached from config
	KnockbackForce float64 // Cached from config
	Direction      string  // "up", "down", "left", "right"
	BaseWidth      float64 // Full-size hitbox width (for pulsing fire scaling)
	BaseHeight     float64 // Full-size hitbox height (for pulsing fire scaling)
	OriginX        float64 // Tiled point X - base edge where fire emanates from
	OriginY        float64 // Tiled point Y - base edge where fire emanates from
	FrameWidth     float64 // Sprite frame width (for calculating sprite center)
	SpriteCenterX  float64 // Pre-calculated sprite center X
	SpriteCenterY  float64 // Pre-calculated sprite center Y
	HitboxPhases   []cfg.FireHitboxPhase // Cached from config (nil = static hitbox)
}

var Fire = donburi.NewComponentType[FireData]()

// FireSpriteCenter calculates the sprite center position from origin, direction, and frame width.
// This is the single source of truth for sprite center calculation.
func FireSpriteCenter(originX, originY, frameWidth float64, direction string) (x, y float64) {
	hw := frameWidth / 2
	switch direction {
	case "left":
		return originX - hw, originY
	case "up":
		return originX, originY - hw
	case "down":
		return originX, originY + hw
	default: // "right"
		return originX + hw, originY
	}
}
