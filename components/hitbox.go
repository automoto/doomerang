package components

import (
	"github.com/yohamta/donburi"
)

type HitboxData struct {
	OwnerEntity    *donburi.Entry          // The entity that created this hitbox (player/enemy)
	Damage         int                     // Damage this hitbox deals
	KnockbackForce float64                 // Knockback strength
	LifeTime       int                     // Frames this hitbox lasts
	HitEntities    map[*donburi.Entry]bool // Entities already hit (prevent multiple hits)
	AttackType     string                  // "punch" or "kick" for different hitbox sizes
	ChargeRatio    float64                 // 0.0 = quick attack, 1.0 = fully charged
}

var Hitbox = donburi.NewComponentType[HitboxData]()
