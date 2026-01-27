package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
)

// Reusable slice for gamepad IDs to avoid allocations
var gamepadIDs []ebiten.GamepadID

// UpdateInput polls raw input and updates the InputComponent.
// Must run BEFORE UpdatePlayer in the system order.
func UpdateInput(ecs *ecs.ECS) {
	input := getOrCreateInput(ecs)

	// Get connected gamepads
	gamepadIDs = ebiten.AppendGamepadIDs(gamepadIDs[:0])

	// Read analog stick values for movement (with deadzone)
	analogLeft, analogRight, analogUp, analogDown := getAnalogStickState(gamepadIDs)

	for actionID, binding := range cfg.Input.Bindings {
		state := computeActionState(binding, gamepadIDs)

		// Add analog stick input for movement actions
		switch actionID {
		case cfg.ActionMoveLeft, cfg.ActionMenuLeft:
			if analogLeft {
				state.Pressed = true
			}
		case cfg.ActionMoveRight, cfg.ActionMenuRight:
			if analogRight {
				state.Pressed = true
			}
		case cfg.ActionCrouch, cfg.ActionMenuDown:
			if analogDown {
				state.Pressed = true
			}
		case cfg.ActionMenuUp:
			if analogUp {
				state.Pressed = true
			}
		}

		input.Actions[actionID] = state
	}
}

// computeActionState checks all bindings for an action and returns combined state
func computeActionState(binding cfg.InputBinding, gamepads []ebiten.GamepadID) components.ActionState {
	var state components.ActionState

	// Check keyboard keys
	for _, key := range binding.Keys {
		if ebiten.IsKeyPressed(key) {
			state.Pressed = true
		}
		if inpututil.IsKeyJustPressed(key) {
			state.JustPressed = true
		}
		if inpututil.IsKeyJustReleased(key) {
			state.JustReleased = true
		}
	}

	// Check standard gamepad buttons on all connected gamepads
	for _, gpID := range gamepads {
		// Only use standard layout if available
		if !ebiten.IsStandardGamepadLayoutAvailable(gpID) {
			continue
		}

		for _, btn := range binding.StandardGamepadButtons {
			if ebiten.IsStandardGamepadButtonPressed(gpID, btn) {
				state.Pressed = true
			}
			if inpututil.IsStandardGamepadButtonJustPressed(gpID, btn) {
				state.JustPressed = true
			}
			if inpututil.IsStandardGamepadButtonJustReleased(gpID, btn) {
				state.JustReleased = true
			}
		}
	}

	return state
}

// getAnalogStickState reads the left analog stick from all gamepads
// Returns directional states based on deadzone threshold
func getAnalogStickState(gamepads []ebiten.GamepadID) (left, right, up, down bool) {
	deadzone := cfg.Input.AnalogDeadzone

	for _, gpID := range gamepads {
		if !ebiten.IsStandardGamepadLayoutAvailable(gpID) {
			continue
		}

		// Read left stick axes
		horizontal := ebiten.StandardGamepadAxisValue(gpID, ebiten.StandardGamepadAxisLeftStickHorizontal)
		vertical := ebiten.StandardGamepadAxisValue(gpID, ebiten.StandardGamepadAxisLeftStickVertical)

		// Apply deadzone
		if horizontal < -deadzone {
			left = true
		}
		if horizontal > deadzone {
			right = true
		}
		if vertical < -deadzone {
			up = true
		}
		if vertical > deadzone {
			down = true
		}
	}

	return
}

// getOrCreateInput returns the singleton Input component, creating if needed
func getOrCreateInput(ecs *ecs.ECS) *components.InputData {
	entry, ok := components.Input.First(ecs.World)
	if !ok {
		entry = ecs.World.Entry(ecs.World.Create(components.Input))
		components.Input.SetValue(entry, components.InputData{
			Actions: make(map[cfg.ActionID]components.ActionState),
		})
	}
	return components.Input.Get(entry)
}
