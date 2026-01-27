## Goal

Add full Xbox and PS5 controller support with standard button mappings and analog stick movement.

## Design

### Supported Controllers
- Xbox controllers (360, One, Series X|S)
- PS5 DualSense
- Generic gamepads using standard mapping

### Button Mapping

| Action | Xbox | PlayStation | Notes |
|--------|------|-------------|-------|
| Jump | A | Cross | |
| Attack | X | Square | |
| Boomerang | B | Circle | |
| Crouch/Slide | D-pad Down | D-pad Down | Also left stick down |
| Pause | Start/Menu | Options | |
| Menu Up | D-pad Up | D-pad Up | Also left stick |
| Menu Down | D-pad Down | D-pad Down | Also left stick |
| Menu Select | A | Cross | |

### Movement Input
- **D-pad**: 4-way directional (left/right for movement, up/down for menus)
- **Left analog stick**:
  - Horizontal axis for left/right movement (deadzone threshold ~0.2)
  - Vertical axis for menu navigation and crouch

### Input Mode Selection
- Manual toggle in Settings screen: "Input: Keyboard / Controller"
- Default: Keyboard
- Keyboard always works even in controller mode (fallback)

### Technical Notes

**Ebitengine Gamepad API:**
- `ebiten.AppendGamepadIDs()` - detect connected controllers
- `ebiten.IsStandardGamepadLayoutAvailable()` - check for standard mapping
- `ebiten.StandardGamepadButton` - cross-platform button constants
- `ebiten.StandardGamepadAxis` - analog stick axes

**Standard Gamepad Buttons (Ebitengine):**
```go
StandardGamepadButtonRightBottom    // A / Cross
StandardGamepadButtonRightRight     // B / Circle
StandardGamepadButtonRightLeft      // X / Square
StandardGamepadButtonRightTop       // Y / Triangle
StandardGamepadButtonCenterRight    // Start / Options
StandardGamepadButtonLeftTop        // D-pad Up
StandardGamepadButtonLeftBottom     // D-pad Down
StandardGamepadButtonLeftLeft       // D-pad Left
StandardGamepadButtonLeftRight      // D-pad Right
```

**Analog Stick Handling:**
```go
StandardGamepadAxisLeftStickHorizontal  // -1.0 (left) to 1.0 (right)
StandardGamepadAxisLeftStickVertical    // -1.0 (up) to 1.0 (down)
```

## Implementation Tasks

### Phase 1: Core Controller Input
- [ ] Update `config/input.go` to use `StandardGamepadButton` instead of raw button IDs
- [ ] Add gamepad bindings for all actions (Jump, Attack, Boomerang, Crouch, Pause)
- [ ] Add gamepad bindings for menu actions (MenuUp, MenuDown, MenuSelect)
- [ ] Update `systems/input.go` to check `IsStandardGamepadLayoutAvailable()`

### Phase 2: Analog Stick Support
- [ ] Add analog stick reading for horizontal movement (left/right)
- [ ] Add analog stick reading for vertical input (crouch, menu nav)
- [ ] Implement deadzone threshold (~0.2) to prevent drift
- [ ] Update `ActionState` or add axis values to `InputData` component

### Phase 3: Input Mode Setting
- [ ] Add `InputMode` enum to config (Keyboard, Controller)
- [ ] Add input mode to game settings/state
- [ ] Wire up setting in Settings screen (depends on #10)
- [ ] Store preference (depends on #8 save system, or use simple config)

## Files to Modify

**Config:**
- `config/input.go` - Add standard gamepad button bindings, input mode enum

**Systems:**
- `systems/input.go` - Add analog stick support, standard layout detection

**Components:**
- `components/input.go` - May need axis values for analog input

**Settings (when #10 is done):**
- Settings screen - Input mode toggle

## Dependencies

- **Settings Screen (#10)**: Required for the manual input mode toggle
- Can implement Phases 1-2 independently
- Phase 3 fully wired when Settings exists

## Testing

1. Connect Xbox controller, verify all buttons work
2. Connect PS5 controller, verify standard mapping works
3. Test analog stick movement with various deadzone values
4. Verify D-pad works for both movement and menus
5. Test switching between keyboard and controller in settings

## References

- [Ebitengine Gamepad Docs](https://ebitengine.org/en/documents/input.html#Gamepad)
- [Standard Gamepad Layout](https://w3c.github.io/gamepad/#remapping)
