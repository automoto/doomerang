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

// InputConfig holds all input mappings
type InputConfig struct {
	Bindings map[ActionID]InputBinding
	// Deadzone for analog stick input (0.0 to 1.0)
	AnalogDeadzone float64
}

// Input is the global input configuration
var Input InputConfig

func init() {
	Input = InputConfig{
		AnalogDeadzone: 0.25,
		Bindings: map[ActionID]InputBinding{
			ActionMoveLeft: {
				Keys: []ebiten.Key{ebiten.KeyLeft, ebiten.KeyA},
				// D-pad Left (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftLeft,
				},
			},
			ActionMoveRight: {
				Keys: []ebiten.Key{ebiten.KeyRight, ebiten.KeyD},
				// D-pad Right (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftRight,
				},
			},
			ActionMoveUp: {
				Keys: []ebiten.Key{ebiten.KeyUp},
				// D-pad Up (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftTop,
				},
			},
			ActionJump: {
				Keys: []ebiten.Key{ebiten.KeyX, ebiten.KeyW},
				// A / Cross button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonRightBottom,
				},
			},
			ActionAttack: {
				Keys: []ebiten.Key{ebiten.KeyZ},
				// X / Square button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonRightLeft,
				},
			},
			ActionCrouch: {
				Keys: []ebiten.Key{ebiten.KeyDown, ebiten.KeyS},
				// D-pad Down (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftBottom,
				},
			},
			ActionBoomerang: {
				Keys: []ebiten.Key{ebiten.KeySpace},
				// B / Circle button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonRightRight,
				},
			},
			ActionPause: {
				Keys: []ebiten.Key{ebiten.KeyEscape, ebiten.KeyP},
				// Start / Options button
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonCenterRight,
				},
			},
			ActionMenuUp: {
				Keys: []ebiten.Key{ebiten.KeyUp, ebiten.KeyW},
				// D-pad Up (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftTop,
				},
			},
			ActionMenuDown: {
				Keys: []ebiten.Key{ebiten.KeyDown, ebiten.KeyS},
				// D-pad Down (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftBottom,
				},
			},
			ActionMenuLeft: {
				Keys: []ebiten.Key{ebiten.KeyLeft, ebiten.KeyA},
				// D-pad Left (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftLeft,
				},
			},
			ActionMenuRight: {
				Keys: []ebiten.Key{ebiten.KeyRight, ebiten.KeyD},
				// D-pad Right (analog stick handled separately)
				StandardGamepadButtons: []ebiten.StandardGamepadButton{
					ebiten.StandardGamepadButtonLeftRight,
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
}
