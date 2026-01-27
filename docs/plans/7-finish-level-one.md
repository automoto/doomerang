## Goal
Complete Level One with a tutorial flow, directional boomerang aiming, a ranged enemy type, fire obstacles, checkpoints, and a level finish transition.

---

## 1. Tutorial System

### Design
Trigger zones in Tiled display help text when the player enters them. Text auto-dismisses after a delay.

### Behavior
1. Player enters trigger zone (Tiled object layer)
2. Text box appears with tip (e.g., "Press SPACE to jump")
3. Text auto-dismisses after configurable delay (e.g., 3-4 seconds)
4. Tip only shows once per life (resets on death/respawn)

### Tiled Integration
- Object layer: `tutorial_triggers`
- Properties per object:
  - `tip_id`: Unique identifier for tracking shown tips
  - `message`: The help text to display
- Duration configured globally in Go config (not per-trigger)

### Tutorial Flow (suggested order)
| Zone | Message |
|------|---------|
| Start | "Use ARROW KEYS to move" |
| First gap | "Press SPACE to jump" |
| First enemy | "Press Z to punch" |
| Tall wall | "Hold toward wall + jump to wall slide" |
| Boomerang pickup | "Press X to throw boomerang" |
| Aiming section | "Hold UP or DOWN while throwing to aim" |
| First fire | "Time your movement past the fire" |

---

## 2. Directional Boomerang Aiming

### Design
Player can aim boomerang throws in 5 directions based on input:

| Input | Throw Direction |
|-------|-----------------|
| None | Forward (facing direction) |
| Up | Straight up |
| Down | Straight down |
| Up + Forward/moving | Diagonal up-forward |
| Down + Forward/moving | Diagonal down-forward |

### Behavior
1. When player presses throw, check directional input
2. Calculate throw angle based on input combination
3. Boomerang travels in that direction
4. Return behavior unchanged (boomerang returns to player)

### Implementation Notes
- "Forward" means the direction the player is facing
- Diagonal detection: Up/Down held AND (moving forward OR forward key pressed)
- Angle values: 0° (right), 45° (up-right), 90° (up), etc.

---

## 3. Ranged Enemy (Knife Thrower)

### Design
A new enemy type that throws knives at the player from a distance. Configured in Go using the existing `EnemyTypeConfig` pattern - Tiled just specifies the type name.

### Behavior
- **Detection:** Activates when player enters range
- **Attack:** Throws knife toward player's position at time of throw
- **Cooldown:** Wait period between throws
- **Movement:** Stationary (doesn't patrol)

### Knife Projectile
- Travels in straight line toward player's position at time of throw
- Faster speed than player movement
- Destroyed on contact with: player, walls, boomerang, or player punch
- Deals damage on player hit

### Tiled Integration
- Same as other enemies - just specify `type: "KnifeThrower"`
- All behavior configured in Go config

### Sprites Needed
- Ranged enemy idle/throw animations
- Knife projectile sprite

---

## 4. Fire Obstacles

### 4a. Pulsing Fire
- **Behavior:** Alternates between active (dangerous) and inactive (safe)
- **Timing:** Configured in Go (not per-instance in Tiled)
- **Visual:** Full fire when active, embers/dim when inactive
- **Damage:** Hurts and knocks back player when active

### 4b. Continuous Fire
- **Behavior:** Always active, always dangerous
- **Visual:** Constant fire animation
- **Damage:** Hurts and knocks back player on contact

### Tiled Integration
- Object layer: `obstacles`
- Just specify `type: "fire_pulsing"` or `type: "fire_continuous"`
- All timing, damage, and knockback configured in Go config

---

## 5. Checkpoints / Spawn Points

### Design
Multiple checkpoints throughout the level. Player respawns at most recent checkpoint after death (if lives remain).

### Behavior
- **Activation:** Last-touched wins (no sequential enforcement required)
- **Respawn:** Death → respawn at last activated checkpoint
- **Out of lives:** Game over → restart level from beginning

### On Respawn
- **Health:** Full health restored
- **Enemies:** All enemies reset to initial state
- **Boomerang:** Player always respawns with boomerang
- **Tutorial tips:** Reset (player sees tips again)
- **Visual:** Checkpoint indicator changes when activated (flag raise, color change, glow)

### Tiled Integration
- Object layer: `checkpoints`
- Just position + optional `checkpoint_id` for debugging
- Activation radius configured in Go config

---

## 6. Level Finish

### Design
Finish zone at end of level triggers transition to next level.

### Behavior
1. Player enters finish zone
2. Screen fades out
3. Load Level Two (or menu if not ready)

### Tiled Integration
- Object in `triggers` layer with type "level_finish"

---

## 7. Level Design Guidelines

### Tutorial Section (Safe)
- Flat ground for movement practice
- Small gap (non-lethal fall) for jump practice
- Single melee enemy in open space for combat practice
- Tall wall requiring wall slide
- Boomerang pickup followed by ranged target

### Main Level (Challenging)
- Mix of melee and ranged enemies
- Introduce ranged enemy alone first, then combine with melee
- Pulsing fire sections (timing challenges)
- Continuous fire as barriers/hazards
- Vertical sections using wall slide
- Checkpoints before difficult sections
- Gradual difficulty increase toward end

### Enemy Placement Rules
- Don't place enemies directly next to hazards (unfair)
- Give player room to maneuver
- Ranged enemies work best on elevated platforms

---

## Implementation Tasks

### Tutorial System
- [ ] Create tutorial trigger entity/component
- [ ] Implement trigger zone detection
- [ ] Create text box UI element
- [ ] Implement auto-dismiss timer
- [ ] Track shown tips per life (reset on death)
- [ ] Load tutorial triggers from Tiled
- [ ] Add tutorial messages to config or Tiled properties

### Directional Aiming
- [x] Modify boomerang throw to read directional input
- [x] Implement 5-way angle calculation
- [x] Handle diagonal detection (up/down + forward)
- [x] Test all 5 throw directions

### Ranged Enemy
- [ ] Create knife projectile entity/component
- [ ] Implement projectile movement (straight line)
- [ ] Implement projectile collision (player, walls, boomerang)
- [ ] Create ranged enemy AI (detect, aim, throw, cooldown)
- [ ] Add ranged enemy archetype
- [ ] Load ranged enemies from Tiled
- [ ] Add sprites/animations

### Fire Obstacles
- [x] Create fire obstacle entity/component
- [x] Implement pulsing fire timer logic
- [x] Implement continuous fire (always active)
- [x] Add damage and knockback on contact
- [x] Add fire sprites/animations
- [x] Load fire obstacles from Tiled

### Checkpoints
- [x] Create checkpoint entity/component
- [x] Implement checkpoint activation on contact
- [x] Track last activated checkpoint
- [x] Modify respawn logic to use checkpoint
- [x] Reset tip tracking on respawn
- [ ] Add checkpoint visual feedback
- [x] Load checkpoints from Tiled

### Level Finish
- [ ] Create finish zone entity
- [ ] Detect player entering zone
- [ ] Implement fade out transition
- [ ] Load next level

### Level Design (Tiled)
- [ ] Create tutorial section at start
- [ ] Add tutorial trigger zones with messages
- [ ] Extend level with platforms and ground
- [ ] Add vertical wall-slide sections
- [ ] Place melee enemies throughout
- [ ] Place ranged enemies strategically
- [ ] Place pulsing fire (timing challenges)
- [ ] Place continuous fire (barriers)
- [ ] Place checkpoints before hard sections
- [ ] Add finish zone at end

---

## Files to Modify/Create

| File | Purpose |
|------|---------|
| `systems/tutorial.go` | Tutorial trigger and text display (create) |
| `systems/boomerang.go` | Directional aiming logic |
| `systems/projectile.go` | Knife projectile logic (create) |
| `systems/enemy_ranged.go` | Ranged enemy AI (create) |
| `systems/obstacles.go` | Fire obstacle logic (create) |
| `systems/checkpoint.go` | Checkpoint logic (create) |
| `systems/level.go` | Level finish detection |
| `systems/factory/projectile.go` | Projectile factory (create) |
| `systems/factory/obstacle.go` | Obstacle factory (create) |
| `components/tutorial.go` | Tutorial components (create) |
| `components/projectile.go` | Projectile component (create) |
| `components/checkpoint.go` | Checkpoint component (create) |
| `components/obstacle.go` | Obstacle component (create) |
| `archetypes/archetypes.go` | New entity archetypes |
| `config/config.go` | New config values |
| `assets/levels/level1.tmx` | Level design |

---

## Config Values

```go
// Tutorial
Tutorial struct {
    DefaultDuration float64 // 3.5 seconds
    FadeInTime      float64 // 0.2 seconds
    FadeOutTime     float64 // 0.3 seconds
}

// Boomerang Aiming
BoomerangAim struct {
    DiagonalAngle float64 // 45 degrees
}

// Ranged Enemy - extends existing EnemyTypeConfig
// Add these fields to EnemyTypeConfig in config/config.go:
//   IsRanged         bool
//   ProjectileSpeed  float64
//   ThrowCooldown    float64
//   ProjectileDamage int

// Fire Obstacles - follows EnemyTypeConfig pattern
FireTypeConfig struct {
    OnDuration  float64 // 0 = always on (continuous)
    OffDuration float64
    Damage      int
    Knockback   float64
}

Fire struct {
    Types map[string]FireTypeConfig
    // Pre-defined types:
    // "fire_pulsing":    {OnDuration: 2.0, OffDuration: 1.5, Damage: 1, Knockback: 250}
    // "fire_continuous": {OnDuration: 0, OffDuration: 0, Damage: 1, Knockback: 250}
}

// Checkpoints
Checkpoint struct {
    ActivationRadius float64 // 24 pixels
}

// Level Transition
LevelTransition struct {
    FadeOutDuration float64 // 0.5 seconds
}
```
