## Goal

Display contextual button prompts in the UI that change based on the player's selected input method (keyboard vs controller).

## Design

### Prompt Display
- **Keyboard mode**: Text-based hints like "X to Jump", "Z to Attack"
- **Controller mode**: Button icons (Xbox A/B/X/Y or PlayStation Cross/Circle/Square/Triangle)

### Prompt Locations
- Menu screens (e.g., "A Select", "B Back")
- Pause screen controls
- Tutorial hints (if added)
- Any in-game contextual prompts

### Icon Sets
Need sprites for both controller families:

**Xbox:**
- A (green), B (red), X (blue), Y (yellow)
- D-pad directions
- Start/Menu button

**PlayStation:**
- Cross, Circle, Square, Triangle
- D-pad directions
- Options button

### Input Mode Detection
- Read from game settings (set in Settings Screen #10)
- `InputMode` enum: `Keyboard`, `ControllerXbox`, `ControllerPlayStation`
- Could auto-detect controller type when in controller mode

## Implementation Tasks

- [ ] Create button icon sprites for Xbox (A, B, X, Y, D-pad, Start)
- [ ] Create button icon sprites for PlayStation (Cross, Circle, Square, Triangle, D-pad, Options)
- [ ] Add prompt rendering helper function that takes an action and returns appropriate icon/text
- [ ] Update menu screens to use prompt helper for hints
- [ ] Update pause screen to show controller prompts
- [ ] Add controller type detection to show correct icon family (Xbox vs PS)

## Files to Modify

**Assets:**
- `assets/ui/` - New button icon sprites

**Systems:**
- `systems/menu.go` - Use prompt helper for menu hints
- `systems/pause.go` - Use prompt helper for pause screen
- New `systems/prompts.go` or helper in `systems/ui.go` - Prompt rendering logic

**Config:**
- `config/input.go` - May need to extend `InputMode` to distinguish Xbox vs PlayStation

## Dependencies

- **Controller Support (#12)**: Needs input mode setting to exist
- **Settings Screen (#10)**: For input mode toggle

## Testing

1. Set input mode to Keyboard, verify text prompts appear
2. Set input mode to Controller with Xbox, verify Xbox icons appear
3. Set input mode to Controller with PS5, verify PlayStation icons appear
4. Navigate all menus and verify prompts are consistent
5. Test pause screen prompts

## References

- Free button prompt assets: [Xelu's Controller Prompts](https://thoseawesomeguys.com/prompts/)
- [Kenney Input Prompts](https://kenney.nl/assets/input-prompts)
