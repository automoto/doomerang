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
	BaseWidth      float64 // Full-size hitbox width
	BaseHeight     float64 // Full-size hitbox height
	SpriteCenterX  float64 // Fixed sprite center X (for rendering)
	SpriteCenterY  float64 // Fixed sprite center Y (for rendering)
	AnchorX        float64 // Fixed anchor point X (for hitbox positioning)
	AnchorY        float64 // Fixed anchor point Y (for hitbox positioning)
	HitboxPhases   []cfg.FireHitboxPhase // Cached from config (nil = static hitbox)
}

var Fire = donburi.NewComponentType[FireData]()
