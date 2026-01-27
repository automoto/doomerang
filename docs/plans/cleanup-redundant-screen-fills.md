# Cleanup: Redundant screen.Fill(color.Black) Calls

## Background
After fixing the white flash issue (removing `SetScreenClearedEveryFrame(false)` and upgrading Ebitengine to v2.9.7), the `screen.Fill(color.Black)` calls in scene Draw() methods are now partially redundant.

## Current State
All three scenes have `screen.Fill(color.Black)` at the start of their Draw() methods:
- `scenes/menu.go:38`
- `scenes/gameover.go:33`
- `scenes/world.go:53`

## Analysis

| Scene | Draws full-screen opaque background? | Fill redundant? |
|-------|-------------------------------------|-----------------|
| menu.go | Yes - `DrawMenu` draws rect with `cfg.Menu.BackgroundColor` | Yes |
| gameover.go | Yes - `DrawGameOver` draws rect with `cfg.GameOver.BackgroundColor` | Yes |
| world.go | No - level background may have transparent areas | No |

## Recommendation
- **Remove** `screen.Fill(color.Black)` from `menu.go` and `gameover.go`
- **Keep** `screen.Fill(color.Black)` in `world.go` as a safety net

## Priority
Low - the redundant fills have negligible performance impact.
