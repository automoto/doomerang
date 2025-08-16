package components

import (
	"github.com/yohamta/donburi"
)

type PlayerData struct {
	Direction    Vector
	ComboCounter int // For tracking punch/kick sequences
}

var Player = donburi.NewComponentType[PlayerData]()
