package config

import "github.com/hajimehoshi/ebiten/v2"

// ActionID represents a logical game action
type ActionID int

const (
	ActionNone ActionID = iota
	ActionMoveLeft
	ActionMoveRight
	ActionMoveUp
	ActionJump
	ActionAttack
	ActionCrouch
	ActionBoomerang
	ActionPause
	ActionMenuUp
	ActionMenuDown
	ActionMenuLeft
	ActionMenuRight
	ActionMenuSelect
	ActionMenuBack
	ActionCount // Must be last - used for array sizing
)

// InputBinding represents a single key or button binding for an action
type InputBinding struct {
	Keys                   []ebiten.Key
	StandardGamepadButtons []ebiten.StandardGamepadButton
}

// ControlSchemeID represents a keyboard control scheme preset
type ControlSchemeID int

const (
	ControlSchemeWASD      ControlSchemeID = iota // WASD movement, Space/J/K actions
	ControlSchemeArrowKeys                        // Arrow key movement, Z/X/C actions
)

// InputConfig holds all input mappings
type InputConfig struct {
	Bindings     map[ActionID]InputBinding
	ActiveScheme ControlSchemeID
	// Deadzone for analog stick input (0.0 to 1.0)
	AnalogDeadzone float64
}

// Input is the global input configuration
var Input InputConfig

func init() {
	Input = InputConfig{
		AnalogDeadzone: 0.25,
		Bindings: map[ActionID]InputBinding{
			// Pause and menu select/back are scheme-independent
			ActionPause: {
				Keys: []ebiten.Key{ebiten.KeyEscape, ebiten.KeyP},
				// Start / Options button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonCenterRight,
				},
			},
			ActionMenuSelect: {
				Keys: []ebiten.Key{ebiten.KeyEnter},
				// A / Cross button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonRightBottom,
				},
			},
			ActionMenuBack: {
				Keys: []ebiten.Key{ebiten.KeyEscape, ebiten.KeyBackspace},
				// B / Circle button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonRightRight,
				},
			},
		},
	}
	ApplyControlScheme(ControlSchemeWASD)
}

// ApplyControlScheme updates keyboard bindings for the given control scheme.
// Only keyboard keys change — all gamepad buttons remain identical across schemes.
// Menu navigation always works with both WASD and arrow keys.
func ApplyControlScheme(scheme ControlSchemeID) {
	Input.ActiveScheme = scheme

	// Gamepad buttons — identical regardless of keyboard scheme
	gpDpadLeft := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftLeft}
	gpDpadRight := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftRight}
	gpDpadUp := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftTop}
	gpDpadDown := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonLeftBottom}
	gpJump := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightBottom}   // A / Cross
	gpAttack := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightLeft}    // X / Square
	gpBoomerang := []ebiten.StandardGamepadButton{ebiten.StandardGamepadButtonRightRight} // B / Circle

	switch scheme {
	case ControlSchemeArrowKeys:
		// Movement on arrow keys, actions on Z/X/C
		Input.Bindings[ActionMoveLeft] = InputBinding{Keys: []ebiten.Key{ebiten.KeyLeft}, StandardGamepadButtons: gpDpadLeft}
		Input.Bindings[ActionMoveRight] = InputBinding{Keys: []ebiten.Key{ebiten.KeyRight}, StandardGamepadButtons: gpDpadRight}
		Input.Bindings[ActionMoveUp] = InputBinding{Keys: []ebiten.Key{ebiten.KeyUp}, StandardGamepadButtons: gpDpadUp}
		Input.Bindings[ActionCrouch] = InputBinding{Keys: []ebiten.Key{ebiten.KeyDown}, StandardGamepadButtons: gpDpadDown}
		Input.Bindings[ActionJump] = InputBinding{Keys: []ebiten.Key{ebiten.KeyZ}, StandardGamepadButtons: gpJump}
		Input.Bindings[ActionAttack] = InputBinding{Keys: []ebiten.Key{ebiten.KeyX}, StandardGamepadButtons: gpAttack}
		Input.Bindings[ActionBoomerang] = InputBinding{Keys: []ebiten.Key{ebiten.KeyC}, StandardGamepadButtons: gpBoomerang}

	default: // ControlSchemeWASD
		// Movement on WASD, actions on Space/J/K
		Input.Bindings[ActionMoveLeft] = InputBinding{Keys: []ebiten.Key{ebiten.KeyA}, StandardGamepadButtons: gpDpadLeft}
		Input.Bindings[ActionMoveRight] = InputBinding{Keys: []ebiten.Key{ebiten.KeyD}, StandardGamepadButtons: gpDpadRight}
		Input.Bindings[ActionMoveUp] = InputBinding{Keys: []ebiten.Key{ebiten.KeyW}, StandardGamepadButtons: gpDpadUp}
		Input.Bindings[ActionCrouch] = InputBinding{Keys: []ebiten.Key{ebiten.KeyS}, StandardGamepadButtons: gpDpadDown}
		Input.Bindings[ActionJump] = InputBinding{Keys: []ebiten.Key{ebiten.KeySpace}, StandardGamepadButtons: gpJump}
		Input.Bindings[ActionAttack] = InputBinding{Keys: []ebiten.Key{ebiten.KeyJ}, StandardGamepadButtons: gpAttack}
		Input.Bindings[ActionBoomerang] = InputBinding{Keys: []ebiten.Key{ebiten.KeyK}, StandardGamepadButtons: gpBoomerang}
	}

	// Menu navigation always supports both WASD and arrow keys (scheme-independent)
	Input.Bindings[ActionMenuUp] = InputBinding{
		Keys:                   []ebiten.Key{ebiten.KeyUp, ebiten.KeyW},
		StandardGamepadButtons: gpDpadUp,
	}
	Input.Bindings[ActionMenuDown] = InputBinding{
		Keys:                   []ebiten.Key{ebiten.KeyDown, ebiten.KeyS},
		StandardGamepadButtons: gpDpadDown,
	}
	Input.Bindings[ActionMenuLeft] = InputBinding{
		Keys:                   []ebiten.Key{ebiten.KeyLeft, ebiten.KeyA},
		StandardGamepadButtons: gpDpadLeft,
	}
	Input.Bindings[ActionMenuRight] = InputBinding{
		Keys:                   []ebiten.Key{ebiten.KeyRight, ebiten.KeyD},
		StandardGamepadButtons: gpDpadRight,
	}
}
