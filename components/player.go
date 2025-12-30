package components

import (
	"github.com/yohamta/donburi"
)

type PlayerData struct {
	Direction           Vector
	ComboCounter        int // For tracking punch/kick sequences
	InvulnFrames        int // Invulnerability frames timer
	BoomerangChargeTime int
	ActiveBoomerang     *donburi.Entry
}

var Player = donburi.NewComponentType[PlayerData]()
