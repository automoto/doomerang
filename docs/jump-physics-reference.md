# Jump Physics Reference

This document provides jump physics calculations and gap sizing guidelines for level design.

## Physics Values

| Parameter | Value | Unit |
|-----------|-------|------|
| Jump Speed | 15.0 | pixels/frame |
| Gravity | 0.75 | pixels/frame² |
| Max Horizontal Speed | 6.0 | pixels/frame |
| Tile Size | 16×16 | pixels |

## Jump Characteristics

### Maximum Jump Height

Using kinematics: `height = v₀² / (2g)`

```
Height = 15² / (2 × 0.75) = 225 / 1.5 = 150 pixels
```

**Maximum Jump Height: ~9.4 tiles** (150 pixels)

### Time in Air

- Time to peak: `t = v₀/g = 15/0.75 = 20 frames`
- Total air time: **40 frames** (~0.67 seconds at 60fps)

### Maximum Horizontal Distance

At max horizontal speed (6 px/frame) over 40 frames:

```
Distance = 6 × 40 = 240 pixels = 15 tiles
```

**Maximum Jump Distance: 15 tiles** (with full speed running start)

## Gap Size Reference Chart

| Difficulty | Gap Width (tiles) | Height Drop | Notes |
|------------|-------------------|-------------|-------|
| **Trivial** | 3-4 | 0 | Can clear without running start |
| **Easy** | 5-7 | 0-2 down | Comfortable with running start |
| **Medium** | 8-10 | 0-3 down | Requires full speed, some precision |
| **Hard** | 11-13 | 2-5 down | Tight timing, max speed required |
| **Expert** | 14-15 | 4-6 down | Pixel-perfect, late jump needed |

## Height Advantage Bonus

Jumping from higher platforms increases horizontal distance due to extra fall time:

| Drop (tiles) | Extra Horizontal (tiles) |
|--------------|--------------------------|
| +2 | ~2.5 |
| +4 | ~4.5 |
| +6 | ~6.0 |

## Practical Recommendations

- **Safe platforming sections**: 4-6 tile gaps
- **Challenging but fair**: 7-9 tile gaps
- **"Can I make that?" moments**: 10-12 tile gaps
- **Leave 1-2 tile margin** for player error on intended jumps
- **Wall jump recovery**: If wall mechanics are available, gaps can be ~20% wider

## Quick Reference

| Scenario | Max Tiles |
|----------|-----------|
| Standing jump height | ~9 |
| Running jump distance (flat) | ~15 |
| Comfortable gap (with margin) | 10-12 |
| Safe gap for most players | 6-8 |
