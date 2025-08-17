// components/melee.go
package components

import "github.com/yohamta/donburi"

type MeleeAttackData struct {
	ComboStep   int     // 0: idle, 1: punch, 2: kick
	ChargeTime  float64 // Time in seconds the attack button has been held
	IsCharging  bool
	IsAttacking bool
}

var MeleeAttack = donburi.NewComponentType[MeleeAttackData]()