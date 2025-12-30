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
	HitEnemies       []*donburi.Entry
	Damage           int
}

var Boomerang = donburi.NewComponentType[BoomerangData]()
