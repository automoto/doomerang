package components

import (
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/features/math"
)

type CameraData struct {
	Position math.Vec2
}

var Camera = donburi.NewComponentType[CameraData]()
