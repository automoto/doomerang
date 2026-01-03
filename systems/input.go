package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi/ecs"
)

// UpdateInput polls raw input and updates the InputComponent.
// Must run BEFORE UpdatePlayer in the system order.
func UpdateInput(ecs *ecs.ECS) {
	input := getOrCreateInput(ecs)

	for actionID, binding := range cfg.Input.Bindings {
		state := computeActionState(binding)
		input.Actions[actionID] = state
	}
}

// computeActionState checks all bindings for an action and returns combined state
func computeActionState(binding cfg.InputBinding) components.ActionState {
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

	// Check gamepad buttons
	for _, gpb := range binding.GamepadButtons {
		if ebiten.IsGamepadButtonPressed(gpb.GamepadID, gpb.Button) {
			state.Pressed = true
		}
		if inpututil.IsGamepadButtonJustPressed(gpb.GamepadID, gpb.Button) {
			state.JustPressed = true
		}
		if inpututil.IsGamepadButtonJustReleased(gpb.GamepadID, gpb.Button) {
			state.JustReleased = true
		}
	}

	return state
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
