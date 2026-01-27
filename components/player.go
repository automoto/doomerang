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
	ChargeVFX           *donburi.Entry // VFX shown while charging boomerang
	LastSafeX           float64        // Last position where player was safely grounded
	LastSafeY           float64
}

var Player = donburi.NewComponentType[PlayerData]()
