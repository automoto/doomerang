## Goal

Add a Settings screen accessible from both the Main Menu and Pause Menu, allowing players to adjust audio, display, and input settings.

## Design

### Settings Menu Structure

```
SETTINGS
─────────────────
Music Volume    [████░░░░░░] 50%
SFX Volume      [██████████] 100%
Mute            [ ] Off
─────────────────
Fullscreen      [ ] Off
Resolution      1280 x 720
─────────────────
Input           Keyboard
─────────────────
< Back
```

### Settings Options

| Setting | Type | Values | Default |
|---------|------|--------|---------|
| Music Volume | Slider | 0%, 25%, 50%, 75%, 100% | 100% |
| SFX Volume | Slider | 0%, 25%, 50%, 75%, 100% | 100% |
| Mute | Toggle | On/Off | Off |
| Fullscreen | Toggle | On/Off | Off |
| Resolution | Selector | See list below | 1280x720 |
| Input | Selector | Keyboard/Controller | Keyboard |

### Resolution Options

Resolutions that scale cleanly with 16:9 aspect ratio:
- 1280 x 720 (720p) - default
- 1600 x 900
- 1920 x 1080 (1080p)
- 2560 x 1440 (1440p)

Note: In fullscreen mode, resolution selector is hidden (uses native resolution).

### Navigation

- **Up/Down**: Move between settings
- **Left/Right**: Adjust value (volume slider, toggle, selector)
- **Select/Enter**: Toggle for checkboxes, or activate "Back"
- **Escape**: Return to previous menu (same as "Back")

### Entry Points

1. **Main Menu → Settings**: Opens settings as overlay or new scene
2. **Pause Menu → Settings**: Opens settings as sub-menu within pause overlay

### Audio Integration

Existing functions in `systems/audio.go`:
```go
SetMusicVolume(e *ecs.ECS, volume float64)  // 0.0 - 1.0
SetSFXVolume(e *ecs.ECS, volume float64)    // 0.0 - 1.0
```

When mute is enabled, set both to 0. Store original values to restore on unmute.

### Display Integration

Ebitengine APIs:
```go
ebiten.SetFullscreen(fullscreen bool)
ebiten.IsFullscreen() bool
ebiten.SetWindowSize(width, height int)
```

### Settings Persistence

Use [gdata](https://github.com/quasilyte/gdata) library for cross-platform settings storage.

```go
import "github.com/quasilyte/gdata"

// Initialize once at startup
m := gdata.Manager{}
m.Open(gdata.Config{
    AppName: "doomerang",
})

// Save settings
m.SaveItem("settings", settingsBytes)

// Load settings
data := m.LoadItem("settings")
```

Benefits:
- Cross-platform (Windows, Mac, Linux, Web/WASM)
- Uses appropriate OS config directories automatically
- Same library will be used for Save Games (#8)

## Implementation Tasks

### Phase 1: Settings Component & State
- [ ] Create `components/settings.go` with `SettingsData` struct
- [ ] Add `SettingsOption` enum for menu navigation
- [ ] Add settings fields: MusicVolume, SFXVolume, Muted, Fullscreen, ResolutionIndex, InputMode
- [ ] Create `GetOrCreateSettings()` helper

### Phase 2: Settings System
- [ ] Create `systems/settings.go` with Update and Draw functions
- [ ] Implement menu navigation (up/down between options)
- [ ] Implement value adjustment (left/right for sliders/selectors)
- [ ] Handle "Back" option to return to previous menu
- [ ] Play menu navigation/select sounds

### Phase 3: Wire Up Audio Settings
- [ ] Connect Music Volume slider to `SetMusicVolume()`
- [ ] Connect SFX Volume slider to `SetSFXVolume()`
- [ ] Implement Mute toggle (store/restore volumes)
- [ ] Play preview sound when adjusting SFX volume

### Phase 4: Wire Up Display Settings
- [ ] Connect Fullscreen toggle to `ebiten.SetFullscreen()`
- [ ] Implement Resolution selector with `ebiten.SetWindowSize()`
- [ ] Hide Resolution option when Fullscreen is enabled
- [ ] Handle edge case: switching from fullscreen restores previous resolution

### Phase 5: Entry Points
- [ ] Update `systems/menu.go` - MainMenuSettings opens settings
- [ ] Update `systems/pause.go` - MenuSettings opens settings
- [ ] Add state tracking: which menu opened settings (for "Back" navigation)
- [ ] Handle Escape key to close settings

### Phase 6: Settings Persistence (using gdata)
- [ ] Add `github.com/quasilyte/gdata` dependency
- [ ] Create `systems/storage.go` with gdata.Manager initialization
- [ ] Implement `SaveSettings()` function (serialize to JSON, save via gdata)
- [ ] Implement `LoadSettings()` function (load via gdata, deserialize JSON)
- [ ] Initialize gdata and call LoadSettings on game start
- [ ] Call SaveSettings when leaving settings screen

### Phase 7: Input Mode Setting (for Controller Support #12)
- [ ] Add Input selector: Keyboard / Controller
- [ ] Store InputMode in settings
- [ ] Wire up to input system (Controller Support #12 will use this)

## Files to Create

- `components/settings.go` - SettingsData component, SettingsOption enum
- `systems/settings.go` - UpdateSettings, DrawSettings functions
- `systems/storage.go` - gdata.Manager wrapper, SaveSettings, LoadSettings

## Files to Modify

- `config/config.go` - Add SettingsConfig with defaults, resolution list
- `systems/menu.go` - Wire up MainMenuSettings
- `systems/pause.go` - Wire up MenuSettings
- `main.go` - Call LoadSettings on startup
- `config/input.go` - Add ActionMenuLeft, ActionMenuRight for value adjustment

## Config Additions

```go
// In config/config.go
type SettingsConfig struct {
    Resolutions []Resolution
    DefaultResolutionIndex int
    DefaultMusicVolume float64
    DefaultSFXVolume float64
}

type Resolution struct {
    Width  int
    Height int
    Label  string  // "1280 x 720"
}

var Settings SettingsConfig
```

## Input Actions to Add

```go
// In config/input.go
ActionMenuLeft   // Left arrow, A key
ActionMenuRight  // Right arrow, D key
```

## Dependencies

- None (standalone feature)
- Enables: Controller Support (#12), Button Prompts (#13)

## Testing

1. Open Settings from Main Menu, verify navigation works
2. Open Settings from Pause Menu, verify navigation works
3. Adjust Music Volume, verify audio changes in real-time
4. Adjust SFX Volume, play sound to confirm
5. Enable Mute, verify all audio stops
6. Disable Mute, verify volumes restored
7. Enable Fullscreen, verify window goes fullscreen
8. Change Resolution, verify window resizes
9. Close and reopen game, verify settings persisted
10. Test Escape key returns to previous menu

## References

- [gdata](https://github.com/quasilyte/gdata) - Cross-platform game data storage
- [Ebitengine Fullscreen](https://ebitengine.org/en/documents/api.html)
