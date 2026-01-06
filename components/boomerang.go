package components

import (
	"github.com/yohamta/donburi"
)

type BoomerangState int

const (
	BoomerangOutbound BoomerangState = iota
	BoomerangInbound
)

type BoomerangData struct {
	Owner            *donburi.Entry
	State            BoomerangState
	DistanceTraveled float64
	MaxRange         float64
	PierceDistance   float64
	HitEnemies       map[*donburi.Entry]struct{}
	Damage           int
	ChargeRatio      float64 // 0.0 = quick throw, 1.0 = fully charged
}

var Boomerang = donburi.NewComponentType[BoomerangData]()
