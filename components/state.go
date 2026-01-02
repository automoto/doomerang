package components

import (
	"github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
)

type StateData struct {
	CurrentState  config.StateID
	PreviousState config.StateID
	StateTimer    int
}

var State = donburi.NewComponentType[StateData]()

type IdleState struct{}
type RunningState struct{}
type JumpingState struct{}
type FallingState struct{}
type WallSlidingState struct{}
type AttackingState struct{}
type CrouchingState struct{}
type StunnedState struct{}

var Idle = donburi.NewComponentType[IdleState]()
var Running = donburi.NewComponentType[RunningState]()
var Jumping = donburi.NewComponentType[JumpingState]()
var Falling = donburi.NewComponentType[FallingState]()
var WallSliding = donburi.NewComponentType[WallSlidingState]()
var Attacking = donburi.NewComponentType[AttackingState]()
var Crouching = donburi.NewComponentType[CrouchingState]()
var Stunned = donburi.NewComponentType[StunnedState]()
