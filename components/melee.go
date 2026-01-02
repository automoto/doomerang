// components/melee.go
package components

import "github.com/yohamta/donburi"

type MeleeAttackData struct {
	ComboStep        int     // 0: idle, 1: punch, 2: kick
	ChargeTime       float64 // Time in seconds the attack button has been held
	IsCharging       bool
	IsAttacking      bool
	ActiveHitbox     *donburi.Entry // Direct reference to the active hitbox
	HasSpawnedHitbox bool           // Prevents multiple hitboxes per attack cycle
}

var MeleeAttack = donburi.NewComponentType[MeleeAttackData]()
