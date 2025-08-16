package components

import (
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
)

type PhysicsData struct {
	SpeedX         float64
	SpeedY         float64
	AccelX         float64
	Gravity        float64
	Friction       float64
	MaxSpeed       float64
	OnGround       *resolv.Object
	WallSliding    *resolv.Object
	IgnorePlatform *resolv.Object
}

var Physics = donburi.NewComponentType[PhysicsData]()
