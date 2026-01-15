# Boomerang Alpha

## Progress

```
Core Gameplay:  [██████████] 100% complete
Content/Polish: [██████░░░░] 57% complete (4/7)
```

---

## Core Gameplay (Complete)

The game engine, mechanics, and feel are fully implemented.

### 1. Pause Screen ✓
document: `1-pause.md`

### 2. Menu Screen ✓
document: `2-menu.md`

### 3. Basic Level ✓
document: `3-basic-level.md`

### 4-5. Audio (Music + Sound Effects) ✓
document: `4-5-audio.md`

### 6a. UI Improvements ✓
document: `6a-ui-improvements.md`

### 6b. Combat Improvements ✓
document: `6b-combat-improvements.md`

### 6c. Visual Juice ✓
document: `6c-visual-juice.md`

---

## Content & Polish (In Progress)

Level design, saves, settings, and release tasks.

### 7. Finish Level One
document: `7-finish-level-one.md`

Complete Level One with tutorial, new mechanics, and full level design:
- **Tutorial system:** Trigger zones show help text (auto-dismiss, once per life)
- **5-way boomerang aiming:** Forward, up, down, up-forward, down-forward
- **Ranged enemy:** Knife-throwing enemy type
- **Fire obstacles:** Pulsing (timed) and continuous fire hazards
- **Checkpoints:** Multiple spawn points, respawn at last activated
- **Level finish:** Fade transition to next level

### 8. Save Games and Checkpoint

Add ability to save games and resume from a checkpoint.

- gdata is already integrated (`systems/persistence.go`) for settings - extend for save games
- Add more checkpoints to level 1
- Wire up "Continue" menu option from Main Menu. Hide it if no save game is found.

### 9. Level Two
document: `9-level-two.md`

"The Highrise" - ascending a cyberpunk tower with increased difficulty:
- **Enemy variants:** Fast Grunt, Heavy Grunt, Rapid Thrower (config-driven, same assets)
- **10 sections:** Combat, platforming, obstacles, and mixed challenges
- **Boss fight:** "The Enforcer" with phase-based AI
- **Level select:** Wire up from Main Menu

### 10. Settings Screen ✓
document: `10-settings-screen.md`

Audio, display, and input settings accessible from Main Menu and Pause Menu.

- Music/SFX volume sliders (0-100%)
- Mute toggle
- Fullscreen toggle
- Resolution selector (720p, 900p, 1080p, 1440p)
- Input mode selector (Keyboard/Controller)
- Settings persistence to disk

### 11. Multi-OS Build, Release and itch.io Upload
document: `11-multi-os-build.md`

Build for Windows, Mac, Linux (via Docker), and Web (WASM) with Makefile targets. Deploy to itch.io using butler.

**Status:** Build targets complete. Remaining: test itch.io uploads via butler.

### 12. Controller Support ✓
document: `12-controller-support.md`

Xbox and PS5 controller support with standard button mappings and analog stick movement.

- Standard gamepad buttons (A/Cross, B/Circle, X/Square, Start/Options)
- D-pad navigation
- Left analog stick with deadzone (0.25)
- Works alongside keyboard (both always active)

### 13. Button Prompts (Optional)
document: `13-button-prompts.md`

UI button prompts that change based on input method (keyboard text vs controller icons).

**Dependencies:** Controller Support (#12), Settings Screen (#10)

---

## Misc Features

Small improvements and polish items.

### M1. Slide Improvements
document: `misc-slide-improvements.md`

- **Slide Kick:** Add ability to throw a kick attack from a slide
- **Slide Timing:** Improve slide input after landing from a jump (reduce delay so down-press registers more easily)
