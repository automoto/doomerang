package components

import "github.com/yohamta/donburi"

type HealthData struct {
	Current int
	Max     int
}

type HealthBarData struct {
	// TimeToLive is the number of frames the health bar should be visible.
	TimeToLive int
}

var Health = donburi.NewComponentType[HealthData]()
var HealthBar = donburi.NewComponentType[HealthBarData]()
