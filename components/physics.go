package components

import (
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
)

// Vector represents a 2D vector.
type Vector struct {
	X, Y float64
}

type PhysicsData struct {
	SpeedX         float64
	SpeedY         float64
	AccelX         float64
	Gravity        float64
	Friction       float64
	AttackFriction float64
	MaxSpeed       float64
	OnGround       *resolv.Object
	WallSliding    *resolv.Object
	IgnorePlatform *resolv.Object
}

var Physics = donburi.NewComponentType[PhysicsData]()
