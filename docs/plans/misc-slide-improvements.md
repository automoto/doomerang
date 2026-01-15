## Goal
Improve the slide mechanic with a kick attack and better input responsiveness after landing.

---

## 1. Slide Kick

### Design
- Player can press attack button during a slide to perform a kick
- Kick has same damage and knockback as regular punch
- Uses a kick animation (or reuses punch animation if no kick sprite exists)
- Slide continues after kick (doesn't cancel the slide)

### Behavior
1. Player is in slide state
2. Player presses attack action
3. Kick hitbox activates, deals damage like punch
4. Slide momentum continues

---

## 2. Slide Timing Improvement

### Problem
After landing from a jump, there's a small delay before the player can initiate a slide by pressing down. This makes the movement feel unresponsive.

### Solution
- Reduce or eliminate the landing recovery frames that block slide input
- Allow slide input to be buffered during landing animation
- Alternatively: allow slide to trigger immediately on land if down is held

---

## Implementation Tasks

### Slide Kick
- [ ] Add slide kick state or extend slide state to handle attack
- [ ] Create kick hitbox (same properties as punch)
- [ ] Wire up attack action during slide to trigger kick
- [ ] Add kick animation (or reuse existing)
- [ ] Add kick sound effect

### Slide Timing
- [ ] Identify where landing delay is enforced
- [ ] Reduce/remove frames that block slide input after landing
- [ ] Test slide responsiveness after jump landing

---

## Files to Modify

| File | Purpose |
|------|---------|
| `systems/player.go` | Slide state logic, attack during slide |
| `systems/combat.go` | Kick hitbox and damage |
| `config/config.go` | Kick damage values (if separate from punch) |
| `config/states.go` | Add SlideKick state if needed |

---

## Config Values

```go
// If kick needs separate values (otherwise reuse Punch values)
SlideKick struct {
    Damage   int     // same as PunchDamage
    Knockback float64 // same as PunchKnockback
}

// Landing
Landing struct {
    SlideBufferFrames int // frames to buffer slide input during landing
}
```
