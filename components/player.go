package components

import (
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
)

type PlayerData struct {
	SpeedX         float64
	SpeedY         float64
	OnGround       *resolv.Object
	WallSliding    *resolv.Object
	FacingRight    bool
	IgnorePlatform *resolv.Object

	// New state management
	CurrentState string // Current state (from config/states.go)
	StateTimer   int    // Frame counter for state duration
	ComboCounter int    // For tracking punch/kick sequences
}

var Player = donburi.NewComponentType[PlayerData]()
