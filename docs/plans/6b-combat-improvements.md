## Goal
Fix and improve combat mechanics to make gameplay feel more fair, responsive, and polished.

## Design

### 1. Hitbox & Damage Normalization
Make the hitboxes the same for the punch as the kick as well as the damage. These two strikes should be the same from a gameplay perspective.

**Requirements:**
- Punch and kick have identical hitbox size
- Punch and kick deal identical damage
- Both attacks should feel interchangeable mechanically

### 2. Knockback Tuning
The combat knockback effects need to be tuned. Right now it seems to work randomly when an enemy will get knocked backwards and enemies cant be knocked back over a ledge.

**Requirements:**
- Knockback should be consistent and predictable
- Enemies should be able to be knocked back over ledges
- Knockback direction and distance should feel satisfying

**Suggested Investigation:**
- Check knockback velocity calculations
- Check if collision detection prevents knockback over edges
- Review knockback conditions (when does it trigger vs not)

### 3. Invuln Frames
Invuln frames need to be slightly extended for a player or we need to modify the stunned effect on them. Right now they can get stuck being stunned when being attacked by multiple enemies and cant escape until they die.

**Requirements:**
- Player has enough invulnerability time to recover
- Player cannot be stun-locked to death by multiple enemies
- Combat should feel fair even when surrounded

**Options:**
- Extend invuln frame duration
- Add knockback to player when hit (creates separation)
- Limit how often player can be stunned
- Add brief super-armor after being hit

### 4. Ducking/Sliding Movement
**Problem 1:** Right now players continue to move forward when ducking, stuck in the ducking pose with no animation after. If they duck when moving forward, they should slow down eventually if they keep holding duck.

**Problem 2:** Need slide ability when running forward and pressing duck.

**Slide Requirements:**
- Trigger: Press duck while running forward
- Animation: Use the last 4 frames of the kick03 animation
- Rotation: Rotate the character correctly to look like they are sliding on the ground
- Speed: Fast during initial slide, slow down gradually if holding duck
- Recovery: Stand back up once they let go of duck after a small delay

**Duck Requirements:**
- If ducking while moving, gradually slow down
- Should not continue at full speed while ducking

### 5. Boomerang Throw While Jumping
Make it so a boomerang can be thrown while jumping. Right now we cant throw it while jumping, add the ability to throw while jumping as well.

**Requirements:**
- Allow throw input during jump state
- Boomerang trajectory should work correctly from air
- Animation should blend or work with jump animation

### 6. Throw Momentum
Make it so when a boomerang is thrown you don't come to a dead stop, you slow down gradually with friction.

**Requirements:**
- Player maintains some momentum when throwing
- Apply friction to gradually slow down
- Should feel fluid, not abrupt

**Implementation:**
- Instead of setting velocity to 0 on throw, reduce it by a percentage
- Apply friction over several frames

### 7. Wall Slide Kick Fix
Fix the issue with wall sliding when a player presses kick, they end up kicking the wall. Make it so you cant strike while wall sliding and if they do a kick, it should disengage the wall slide and do a normal jump kick facing away from the wall.

**Requirements:**
- Cannot perform strike while wall sliding
- If kick is pressed during wall slide:
  - Disengage from wall
  - Perform jump kick facing AWAY from wall
  - Normal jump kick behavior after leaving wall

### 8. Jump Kick Angle
Right now the jump kick is horizontal, angle the players kick slightly downwards as well as its hitbox. You should be able to use a vector to rotate the player slightly diagonally.

**Requirements:**
- Jump kick sprite rotated slightly downward (diagonal angle)
- Hitbox also angled to match the visual
- Use vector rotation for proper angle calculation

**Implementation:**
- Apply rotation transform to sprite during jump kick
- Adjust hitbox position/shape to match rotated visual
- Suggested angle: 15-30 degrees downward

## Implementation Tasks

### Bug Fixes
- [ ] **Death respawn bug:** Fix player not respawning after combat death - must also decrement lives and trigger Game Over when lives = 0 (currently only works for DeathZone)

### Improvements
- [ ] **Hitboxes:** Normalize punch/kick hitbox sizes and damage values
- [ ] **Knockback:** Fix knockback consistency and allow knockback over ledges
- [ ] **Invuln:** Extend invuln frames or add stun-lock prevention
- [ ] **Duck slowdown:** Add friction/slowdown when ducking while moving
- [ ] **Slide:** Implement slide when pressing duck while running
- [ ] **Slide animation:** Use kick03 last 4 frames, rotated for ground slide
- [ ] **Slide recovery:** Add delay before standing up after slide
- [ ] **Air throw:** Allow boomerang throw during jump state
- [ ] **Throw momentum:** Replace instant stop with gradual friction
- [ ] **Wall slide kick:** Block strikes during wall slide, kick = wall jump kick away
- [ ] **Jump kick angle:** Rotate sprite and hitbox downward during jump kick

## Files to Modify
- `systems/combat.go` - Hitbox, damage, knockback logic
- `systems/player.go` - Movement states, ducking, sliding, wall slide
- `systems/boomerang.go` - Throw conditions, momentum
- `components/player.go` - State flags if needed
- `config/config.go` - Combat values (damage, knockback, invuln duration, angles)
- Animation/sprite handling for slide and jump kick rotation
