package components

import (
	"github.com/yohamta/donburi"
)

// TODO: update this effect to be better looking
type HitboxData struct {
	OwnerEntity    *donburi.Entry          // The entity that created this hitbox (player/enemy)
	Damage         int                     // Damage this hitbox deals
	KnockbackForce float64                 // Knockback strength
	LifeTime       int                     // Frames this hitbox lasts
	HitEntities    map[*donburi.Entry]bool // Entities already hit (prevent multiple hits)
	AttackType     string                  // "punch" or "kick" for different hitbox sizes
}

var Hitbox = donburi.NewComponentType[HitboxData]()
