package components

import (
	"github.com/yohamta/donburi"
)

type KnifeData struct {
	Owner  *donburi.Entry
	Damage int
	Speed  float64
}

var Knife = donburi.NewComponentType[KnifeData]()
