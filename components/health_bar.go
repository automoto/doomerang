package components

import "github.com/yohamta/donburi"

type HealthBarData struct {
	// TimeToLive is the number of frames the health bar should be visible.
	TimeToLive int
}

var HealthBar = donburi.NewComponentType[HealthBarData]()