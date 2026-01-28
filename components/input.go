package components

import (
	cfg "github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
)

// InputMethod represents the type of input device being used
type InputMethod int

const (
	InputKeyboard InputMethod = iota
	InputXbox
	InputPlayStation
)

// ActionState represents the temporal state of an action
type ActionState struct {
	Pressed      bool // Currently held down
	JustPressed  bool // Pressed this frame
	JustReleased bool // Released this frame
}

// InputData stores the current and previous frame's pressed state for all actions.
// JustPressed/JustReleased are computed on-demand by comparing frames.
type InputData struct {
	Current         [cfg.ActionCount]bool // Current frame's Pressed state
	Previous        [cfg.ActionCount]bool // Previous frame's Pressed state
	LastInputMethod InputMethod           // Most recently used input method
}

var Input = donburi.NewComponentType[InputData]()
