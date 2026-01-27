## Goal
Add "game juice" effects across the entire game to make combat, movement, and boomerang mechanics feel satisfying and responsive.

## Reference
See `game-feel-juice.md` for background on game feel concepts.

## Assets

We have visual effect sprites available:
- **SFX sprites:** `assets/images/spritesheets/sfx/` (explosions, impacts)
- **Jump dust:** `assets/images/spritesheets/player/jumpdust.png`
- **Landing dust:** `assets/images/spritesheets/player/landingdust.png`
- **Slide dust:** `assets/images/spritesheets/player/slidedust.png`

---

## 1. Movement Effects

### Jump Effects
- **Jump Dust:** Spawn `jumpdust.png` animation at player's feet when jumping
- **Squash/Stretch:** Stretch player sprite vertically on jump (scale Y ~1.1-1.2)

### Landing Effects
- **Landing Dust:** Spawn `landingdust.png` at player's feet on ground impact
- **Squash/Stretch:** Squash player sprite horizontally on land (scale X ~1.1-1.2, scale Y ~0.85)
- **Recovery:** Lerp back to normal scale over ~0.1s

### Ground Slide Effects
- **Slide Dust:** Spawn `slidedust.png` when player initiates a ground slide (press down while running)
- Spawn at player's feet at slide start

---

## 2. Combat Effects

### Player Punch
- **Hit Flash:** Flash enemy sprite white for 1-2 frames on hit
- **Hit Particles:** Spawn impact particles at hit location
- **Screen Shake:** Subtle shake on successful hit

### Enemy Attacks
- **Player Hit Flash:** Flash player sprite red/white when taking damage
- **Screen Shake:** Shake on player damage (slightly stronger than dealing damage)
- **Health Bar Shake:** Shake the player's health bar UI element

### Effect Positioning
- Position effects to avoid blocking enemy health bars (spawn slightly below center or offset to sides)

---

## 3. Boomerang Effects

### Charged Throw
- **Charge-Up:** Particles/glow on player while charging
- **Particle Trail:** Significant visible trail behind boomerang
- **Impact:** `explosion_short.png` on enemy hit
- **Screen Shake:** Noticeable shake on impact

### Quick Throw
- **Particle Trail:** Lighter/shorter trail
- **Impact:** `plasma.png` on enemy hit (lighter effect)
- **Screen Shake:** None or very subtle

### Universal
- **Flight Sound:** Loop with pitch modulated by velocity
- **Catch Sound:** Satisfying sound when player catches boomerang

---

## 4. Core Systems Required

### Screen Shake System
- Offset camera by random amount within intensity radius
- Decay over duration
- Support directional shake (away from impact point)

### Squash/Stretch System
- Apply scale multipliers to player sprite
- Lerp back to (1.0, 1.0) over time

### Sprite Flash System
- Temporarily replace sprite shader/color
- White flash for hits dealt, red flash for damage taken

---

## Implementation Tasks

### Phase 1: Core Systems
- [x] Implement screen shake system in camera
- [x] Implement sprite flash system (white/red flash)
- [x] Implement squash/stretch system for player

### Phase 2: Movement Effects
- [x] Load jump/landing/slide dust sprites
- [x] Spawn jump dust on player jump
- [x] Spawn landing dust on player land
- [x] Spawn slide dust on ground slide start
- [x] Add squash/stretch on jump and land

### Phase 3: Combat Effects
- [x] Add hit flash on enemy damage
- [x] Add hit particles on punch connect (HitExplosion, scaled by charge)
- [x] Add screen shake on punch connect
- [x] Add player damage flash
- ~~Add health bar shake on player damage~~ (skipped)

### Phase 4: Boomerang Effects
- [x] Add charge-up visual on player (level_up VFX at feet after 15 frames)
- [x] Add charge-up sound effect (boomerang_charge.wav)
- [x] Add muzzle flash on throw (gunshot_rifle effect)
- [x] Add impact effects (explosion vs plasma)
- [x] Wire up screen shake for boomerang hits
- [x] Add catch sound (already wired up)
- ~~Add flight loop sound~~ (skipped)

### Phase 5: Polish
- [x] Tune all timing values
- [x] Ensure effects don't block enemy health bars
- [x] Balance screen shake (not too intense)
- [x] Test feel across all interactions
- [x] Add slide sound effect

## Status: COMPLETE

---

## Files to Modify/Create

| File | Purpose |
|------|---------|
| `systems/effects.go` | Sprite flash, squash/stretch (create) |
| `systems/camera.go` | Screen shake implementation |
| `systems/particles.go` | Particle spawning and rendering (create) |
| `systems/player.go` | Trigger movement effects |
| `systems/combat.go` | Trigger combat effects |
| `systems/boomerang.go` | Trigger throw/hit/catch effects |
| `systems/ui.go` | Health bar shake |
| `components/effects.go` | Effect-related components (create) |
| `config/config.go` | Effect parameters |

---

## Suggested Config Values

```go
// Screen Shake
ScreenShake struct {
    PunchIntensity        float64 // 2.0 pixels
    PunchDuration         float64 // 0.08 seconds
    PlayerDamageIntensity float64 // 4.0 pixels
    PlayerDamageDuration  float64 // 0.12 seconds
    ChargedThrowIntensity float64 // 4.0 pixels
    ChargedThrowDuration  float64 // 0.1 seconds
}

// Squash/Stretch
SquashStretch struct {
    JumpStretchY float64 // 1.15
    JumpStretchX float64 // 0.9
    LandSquashY  float64 // 0.85
    LandSquashX  float64 // 1.15
    RecoveryTime float64 // 0.1 seconds
}

// Sprite Flash
SpriteFlash struct {
    HitFlashDuration    float64 // 0.05 seconds
    DamageFlashDuration float64 // 0.1 seconds
}
```
