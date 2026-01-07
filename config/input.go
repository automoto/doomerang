package config

import "github.com/hajimehoshi/ebiten/v2"

// ActionID represents a logical game action
type ActionID int

const (
	ActionNone ActionID = iota
	ActionMoveLeft
	ActionMoveRight
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
)

// GamepadBinding represents a gamepad button on a specific gamepad
type GamepadBinding struct {
	GamepadID ebiten.GamepadID
	Button    ebiten.GamepadButton
}

// InputBinding represents a single key or button binding for an action
type InputBinding struct {
	Keys           []ebiten.Key
	GamepadButtons []GamepadBinding
}

// InputConfig holds all input mappings
type InputConfig struct {
	Bindings map[ActionID]InputBinding
}

// Input is the global input configuration
var Input InputConfig

func init() {
	Input = InputConfig{
		Bindings: map[ActionID]InputBinding{
			ActionMoveLeft: {
				Keys: []ebiten.Key{ebiten.KeyLeft, ebiten.KeyA},
			},
			ActionMoveRight: {
				Keys: []ebiten.Key{ebiten.KeyRight, ebiten.KeyD},
			},
			ActionJump: {
				Keys: []ebiten.Key{ebiten.KeyX, ebiten.KeyW},
				GamepadButtons: []GamepadBinding{
					{GamepadID: 0, Button: ebiten.GamepadButton0},
					{GamepadID: 1, Button: ebiten.GamepadButton0},
				},
			},
			ActionAttack: {
				Keys: []ebiten.Key{ebiten.KeyZ},
			},
			ActionCrouch: {
				Keys: []ebiten.Key{ebiten.KeyDown, ebiten.KeyS},
			},
			ActionBoomerang: {
				Keys: []ebiten.Key{ebiten.KeySpace},
			},
			ActionPause: {
				Keys: []ebiten.Key{ebiten.KeyEscape, ebiten.KeyP},
			},
			ActionMenuUp: {
				Keys: []ebiten.Key{ebiten.KeyUp, ebiten.KeyW},
			},
			ActionMenuDown: {
				Keys: []ebiten.Key{ebiten.KeyDown, ebiten.KeyS},
			},
			ActionMenuLeft: {
				Keys: []ebiten.Key{ebiten.KeyLeft, ebiten.KeyA},
			},
			ActionMenuRight: {
				Keys: []ebiten.Key{ebiten.KeyRight, ebiten.KeyD},
			},
			ActionMenuSelect: {
				Keys: []ebiten.Key{ebiten.KeyEnter},
			},
		},
	}
}
