package components

import "github.com/yohamta/donburi"

type DamageEventData struct {
	Amount     int
	KnockbackX float64
	KnockbackY float64
}

var DamageEvent = donburi.NewComponentType[DamageEventData]()
