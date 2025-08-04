package components

import "github.com/yohamta/donburi"

type DamageEventData struct {
	Amount int
}

var DamageEvent = donburi.NewComponentType[DamageEventData]()
