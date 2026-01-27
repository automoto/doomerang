## Goal
Replace the placeholder level with a polished, short (~30 sec) level using the Retro Cyberpunk theme assets. Add dead zones that cost a life when fallen into.

## Design

### Visual Theme
- **Tiles**: `Street Tiles/Stylish Black` - sleek dark urban look
- **Background**: `Background-Night.png` with parallax city layers
- **Props**: Street lamps, neon signs for atmosphere

### Layout
- Short level (~30 seconds to complete)
- Multiple platforms with gaps requiring jumps
- Enemy placement for combat encounters
- Clear path from start to end

### Dead Zones
- **Trigger**: Invisible collision zones below platforms / at pit bottoms
- **Behavior**:
  1. Player loses 1 life
  2. Player respawns at level start (or last checkpoint if implemented)
  3. If lives = 0, trigger Game Over
- **Visual**: No visual indicator needed (falling off screen is obvious)

## Implementation Tasks

### 3.1 Update Level Art (Tiled work)
**status:** done

- Add new tileset named cyberpunk-tiles
- Add background image
- Update `ground-walls` layer in tiled with new locations.
- Remove name "ground" from the `ground-walls` objects. Just assume anything in that layer is a solid.

### 3.2 Code Changes for New Assets
The level loader (`assets/assets.go`) already supports:
- Multiple tilesets via go-tiled
- Image layers (currently disabled with `render="false"`)
- Layer compositing into single Background image

Code changes needed:
- [ ] Ensure new tileset paths resolve correctly (relative to .tmx file)
- [ ] Verify image layer renders (set `render="true"` in Tiled)
- [ ] Test that `go-tiled/render` handles the Collection of Images tileset type

Files to check:
- `assets/assets.go` (lines 201-226) - Level rendering logic
- `assets/levels/level1.tmx` - Level file
- `assets/levels/*.tsx` - Tileset definitions

### 3.3 Dead Zones (Code work)
- [ ] Create "DeadZone" object layer in Tiled with rectangle objects
- [ ] Add `tags.ResolvDeadZone` tag constant in `tags/tags.go`
- [ ] Update `assets/assets.go` to parse DeadZone objects (similar to ground-walls)
- [ ] Create dead zone collision objects in resolv.Space
- [ ] Handle dead zone collision in player physics system
- [ ] Add `Lives` field to player component (or create LivesData component)
- [ ] Implement life loss + respawn logic
- [ ] Add Game Over state when lives = 0 (transition to menu or game over screen)

## Files to Modify
- `assets/levels/level1.tmx` - Level file in Tiled
- `assets/levels/tilesets/*.tsx` - New tileset files
- `assets/assets.go` - Parse DeadZone objects from Tiled
- `tags/tags.go` - Add `ResolvDeadZone` tag
- `systems/player.go` or `systems/physics.go` - Dead zone collision handling
- `components/player.go` - Add lives tracking
- `config/config.go` - Add starting lives count, respawn settings

## Reference
- See `docs/plans/level-update-guide.md` for detailed Tiled workflow
