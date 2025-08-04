package components

import "github.com/yohamta/donburi"

type HealthData struct {
	Current int
	Max     int
}

var Health = donburi.NewComponentType[HealthData]()
