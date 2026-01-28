package components

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
)

type CameraData struct {
	Position   math.Vec2
	LookAheadX float64 // Current smoothed X offset for look-ahead
}

var Camera = donburi.NewComponentType[CameraData]()
