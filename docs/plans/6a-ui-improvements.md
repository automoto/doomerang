## Goal
Fix UI issues and add missing UI elements to improve player feedback and display quality.

## Design

### Lives Counter UI
Display the player's remaining lives on screen so they know how many attempts they have left.

**Requirements:**
- Show current lives count (e.g., heart icons or "x3" text)
- Position in corner of screen (top-left recommended, away from other HUD elements)
- Update immediately when player loses a life
- Visual feedback when life is lost (brief flash or animation)

**Suggested Implementation:**
- Add to existing HUD system (`systems/hud.go` if exists)
- Use simple sprite icons or text rendering
- Read lives from player component

### Screen Resize Fix
Fix issue where resizing the window shows black bars and zooms out too much.

**Current Problem:**
- When screen is resized, game shows black bars
- Camera zooms out incorrectly
- Aspect ratio handling is broken

**Requirements:**
- Maintain proper aspect ratio when window is resized
- No excessive black bars (or use letterboxing correctly)
- Camera should not zoom out unexpectedly
- Game should remain playable at different window sizes

**Suggested Investigation:**
- Check `ebiten.SetWindowResizingMode()` settings
- Review camera viewport calculations
- Check how screen dimensions are used in render calculations

## Implementation Tasks

### Lives Counter
- [ ] Add lives display to HUD
- [ ] Position in top-left corner (or appropriate location)
- [ ] Update display when lives change
- [ ] Add visual feedback on life loss (optional flash/animation)

### Screen Resize
- [ ] Investigate current resize behavior
- [ ] Fix aspect ratio calculations
- [ ] Fix camera viewport on resize
- [ ] Test at multiple window sizes
- [ ] Ensure game remains playable after resize

## Files to Modify
- `systems/hud.go` - Add lives display
- `systems/camera.go` - Fix viewport calculations
- `main.go` or game init - Window resize settings
- `config/config.go` - Screen/viewport configuration
