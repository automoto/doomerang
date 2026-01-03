package components

import (
	cfg "github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
)

// ActionState represents the temporal state of an action
type ActionState struct {
	Pressed      bool // Currently held down
	JustPressed  bool // Pressed this frame
	JustReleased bool // Released this frame
}

// InputData stores the current state of all logical actions
type InputData struct {
	Actions map[cfg.ActionID]ActionState
}

var Input = donburi.NewComponentType[InputData]()
