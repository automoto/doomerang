## Goal
Create Level 2 with increased difficulty using Level 1 mechanics, add simple enemy variants, and include a boss fight.

---

## Setting: "The Highrise"

Ascending a cyberpunk tower - rooftops, scaffolding, neon signs, ventilation shafts. Same visual style as Level 1, different layout.

---

## Enemy Variants

New enemy types that reuse existing assets and code with minimal changes.

### Fast Grunt
- **Base:** Melee enemy
- **Change:** Faster movement speed, slightly less HP
- **Visual:** Same sprite, different color tint (blue/cyan)
- **Implementation:** Same AI, tweak config values

### Heavy Grunt
- **Base:** Melee enemy
- **Change:** Slower movement, more HP (2-3 hits to kill)
- **Visual:** Same sprite, different color tint (red/orange)
- **Implementation:** Same AI, tweak config values

### Rapid Thrower
- **Base:** Ranged enemy (knife thrower)
- **Change:** Faster throw rate, shorter detection range
- **Visual:** Same sprite, different color tint
- **Implementation:** Same AI, tweak config values

### Implementation Approach
- Add `EnemyVariant` property in Tiled (e.g., "fast", "heavy", "rapid")
- Factory reads variant and applies config overrides
- One config struct per variant with speed/HP/damage multipliers

```go
// Example config
EnemyVariants struct {
    Fast  EnemyModifier // SpeedMult: 1.5, HPMult: 0.75
    Heavy EnemyModifier // SpeedMult: 0.6, HPMult: 2.0
    Rapid EnemyModifier // CooldownMult: 0.5, RangeMult: 0.7
}
```

---

## Boss Fight: "The Enforcer"

A larger, tougher enemy at the end of Level 2.

### Design Goals
- Feel like a real fight, not just a damage sponge
- Reuse existing attack patterns where possible
- Predictable patterns player can learn

### Boss Behavior
| Phase | HP Range | Behavior |
|-------|----------|----------|
| **Phase 1** | 100-60% | Melee attacks, slow movement, telegraphed swings |
| **Phase 2** | 60-30% | Adds knife throws between melee attacks |
| **Phase 3** | 30-0% | Faster movement, shorter cooldowns |

### Attack Patterns
1. **Melee Combo:** 2-3 punch sequence (reuse melee enemy attack logic)
2. **Knife Barrage:** Throws 3 knives in spread pattern (reuse projectile code)
3. **Charge:** Rushes toward player, vulnerable if misses (simple state machine)

### Visual
- Larger sprite (1.5-2x scale of normal enemy, or unique boss sprite)
- Health bar at top of screen (or large bar above boss)
- Flash/shake on hit for feedback

### Arena Design
- Flat arena with walls on sides
- No pits (fair fight, can't cheese with knockback)
- Fire hazards at edges activate in Phase 3 (optional pressure)

### Implementation Scope
- New boss entity with phase-based AI
- Reuses: melee attack, projectile system, hit flash, screen shake
- New: phase transitions, spread shot pattern, charge attack

---

## Level Sections

| # | Section | Type | Challenge | Enemies/Obstacles |
|---|---------|------|-----------|-------------------|
| 1 | **Rooftop Entry** | Combat | Warm-up, more enemies than L1 end | 3-4 melee grunts |
| 2 | **Scaffold Climb** | Platforming | Tight wall-slide sequences | Minimal enemies |
| 3 | **Vent Shaft** | Obstacle | Fire timing gauntlet | Pulsing + continuous fire |
| 4 | **Neon Plaza** | Combat | Elevation, aim up at ranged | Ranged on platforms, melee below |
| 5 | **Crane Crossing** | Platforming | Long gaps, precise jumps | Fire hazards on platforms |
| 6 | **Generator Room** | Mixed | Fight while avoiding hazards | Heavy Grunts + pulsing fire |
| 7 | **Antenna Tower** | Vertical | Extended wall-slide climb | Fast Grunts on ledges |
| 8 | **Security Hub** | Combat | Enemy gauntlet before boss | Mix of all enemy types |
| 9 | **Rooftop Arena** | Boss | Boss fight | The Enforcer |
| 10 | **Exit** | Finish | Victory | Level complete trigger |

---

## Difficulty Progression

```
Section:  1    2    3    4    5    6    7    8    9    10
Difficulty: ██  ██░  ███  ████  ███  █████  ████  ██████  ████████  ░
           Intro     Building          Peak        BOSS      Done

Checkpoints: Start, after 3, after 5, after 7, before 9 (boss)
```

---

## Implementation Tasks

### Enemy Variants
- [ ] Add EnemyVariant property support in Tiled loader
- [ ] Create variant config structs (Fast, Heavy, Rapid)
- [ ] Modify enemy factory to apply variant modifiers
- [ ] Add color tint support for enemy sprites
- [ ] Place variants in Level 2 Tiled map

### Boss Fight
- [ ] Create boss entity and component
- [ ] Implement phase-based AI state machine
- [ ] Add melee combo attack (reuse hit detection)
- [ ] Add knife spread attack (reuse projectile system)
- [ ] Add charge attack with recovery state
- [ ] Create boss health bar UI
- [ ] Add phase transition effects (flash, pause)
- [ ] Design boss arena in Tiled
- [ ] Add boss music/sound cues (optional)

### Level Design (Tiled)
- [ ] Create level2.tmx
- [ ] Build Section 1: Rooftop Entry
- [ ] Build Section 2: Scaffold Climb
- [ ] Build Section 3: Vent Shaft
- [ ] Build Section 4: Neon Plaza
- [ ] Build Section 5: Crane Crossing
- [ ] Build Section 6: Generator Room
- [ ] Build Section 7: Antenna Tower
- [ ] Build Section 8: Security Hub
- [ ] Build Section 9: Boss Arena
- [ ] Place checkpoints
- [ ] Place tutorial tips (if needed for new variants)
- [ ] Add finish zone

### Integration
- [ ] Wire up Level 2 load from Level 1 finish
- [ ] Add Level 2 to level select (if implemented)
- [ ] Victory screen or ending after boss defeat

---

## Files to Modify/Create

| File | Purpose |
|------|---------|
| `config/config.go` | Enemy variant configs, boss config |
| `systems/factory/enemy.go` | Variant modifier application |
| `systems/enemy.go` | Color tint support |
| `systems/boss.go` | Boss AI and phases (create) |
| `systems/ui.go` | Boss health bar |
| `components/boss.go` | Boss component (create) |
| `archetypes/archetypes.go` | Boss archetype |
| `assets/levels/level2.tmx` | Level 2 map (create) |

---

## Config Values

```go
// Enemy Variants
EnemyVariants struct {
    Fast struct {
        SpeedMult  float64 // 1.5
        HPMult     float64 // 0.75
        DamageMult float64 // 1.0
    }
    Heavy struct {
        SpeedMult  float64 // 0.6
        HPMult     float64 // 2.0
        DamageMult float64 // 1.25
    }
    RapidThrower struct {
        CooldownMult float64 // 0.5
        RangeMult    float64 // 0.7
    }
}

// Boss
Boss struct {
    MaxHP            int     // 100
    Phase2Threshold  float64 // 0.6 (60% HP)
    Phase3Threshold  float64 // 0.3 (30% HP)
    MeleeComboCount  int     // 3
    KnifeSpreadCount int     // 3
    KnifeSpreadAngle float64 // 30 degrees
    ChargeSpeed      float64 // 400
    ChargeRecovery   float64 // 1.0 seconds
}
```

---

## Notes

- Enemy variants are config-driven, minimal new code
- Boss is the main new implementation work
- Level design is the bulk of the effort (Tiled work)
- Can ship Level 2 without boss initially, add boss as polish
