# Procedural Level Generation -- Full System Vision

## Overview

Doomerang currently has hand-crafted TMX levels designed in Tiled. To increase replayability, we're adding a **separate roguelite game mode** that procedurally generates levels using a **hybrid chunk-based approach** (Dead Cells style). Hand-designed room chunks are authored in Tiled and assembled algorithmically with dynamic enemy/hazard placement. The existing campaign mode remains unchanged.

**Approach:** Hybrid chunk-based (hand-authored chunks assembled by algorithm)
**Content scope:** Layout + enemies + hazards (all dynamically placed)

---

## Architecture Overview

A new `procgen/` package handles all generation logic. The pipeline produces an `assets.Level` struct that the existing game systems consume unchanged -- no modifications to physics, collision, rendering, or combat systems.

### New Files

| File | Purpose |
|------|---------|
| `procgen/chunk.go` | Chunk struct, loader, connection point parsing |
| `procgen/graph.go` | Concept graph generation |
| `procgen/assembler.go` | Chunk selection + spatial layout |
| `procgen/compiler.go` | Compile assembled chunks into `assets.Level` |
| `procgen/enemies.go` | Dynamic enemy placement with budget system |
| `procgen/hazards.go` | Dynamic hazard (fire, dead zone) placement |
| `procgen/validator.go` | Solvability verification via jump arc math |
| `procgen/director.go` | AI pacing director |
| `procgen/difficulty.go` | Difficulty scaling functions |
| `procgen/decorator.go` | Visual variety (prop scatter, tinting) |
| `config/procgen.go` | Generation configuration values |
| `scenes/roguelite.go` | Roguelite scene with run state management |
| `components/run.go` | Run state ECS component |
| `assets/chunks/` | Hand-authored chunk TMX files |

### Reused Unchanged

- `assets/assets.go` -- `Level` struct, TMX parsing via `go-tiled`
- `systems/factory/*.go` -- `CreateWall`, `CreateEnemy`, `CreateDeadZone`, `CreateFire`, `CreateCheckpoint`, `CreateFinishLine`, `CreatePlayer`
- `systems/physics.go`, `systems/collision.go`, `systems/camera.go` -- all game systems
- `archetypes/archetypes.go`, `tags/tags.go`

---

## Subsystem Designs

### 1. Chunk Format & Authoring

Chunks are standard Tiled TMX files with additional custom properties. Same layer structure as `level1.tmx`:
- `wg-tiles` layer for collision geometry
- Visual tile layers with `render=true` for backgrounds
- New `Connections` object group with entry/exit points encoding: edge (left/right/top/bottom), slot index, Y-offset, opening width

**Map-level custom properties:** `chunk_id`, `biome`, `difficulty` (1-5), `tags` (combat/traversal/break/vertical/hazard), `min_enemies`, `max_enemies`

**Standard chunk widths:** Small (20 tiles), Medium (40 tiles), Large (60 tiles). Height variable 10-30 tiles.

### 2. Concept Graph Generation

Before selecting chunks, build an abstract node graph describing room types and sequence:
- Node types: Start, Combat, Traversal, BreakRoom, Arena, Exit
- Each node has target difficulty, biome hint, and required tags
- Graph spine (critical path) is linear; branches added post-MVP

### 3. Chunk Assembly

Walk the concept graph, select chunks matching each node's requirements (biome, difficulty, tags, connectivity), lay them out spatially left-to-right with connection point alignment. Maintain a "cursor" position tracking the right edge of the last placed chunk.

**Selection scoring:** Variety bonus (penalize reuse), difficulty proximity, size appropriateness for node type. Fallback to universal connector chunks if no match found.

### 4. Dynamic Enemy Placement

Budget system: each enemy type has a point cost (LightGuard=2, Guard=3, KnifeThrower=4, HeavyGuard=5). Budget = base + (difficulty * multiplier). Platform surfaces discovered from SolidTiles. Enemies distributed proportionally to platform length with spacing enforcement (32px minimum).

### 5. Dynamic Hazard Placement

- **DeadZones:** placed in floor gaps wider than 3 tiles, probability scales with difficulty
- **Fire:** chunks define HazardSlot positions, difficulty determines activation percentage and fire type (pulsing=easier, continuous=harder)

### 6. Level Compilation

Transform all chunk-local coordinates to world-space, merge into a single `assets.Level`. Composite background by rendering each chunk's visual tile layers onto one `ebiten.Image`. Place player spawn at start chunk, finish line at exit chunk, checkpoints at break rooms.

**Critical integration point:** The compiled `assets.Level` feeds directly into the existing `world.go` configure() pipeline (lines 116-240) which creates walls, dead zones, fires, enemies, etc. from the level data.

### 7. Solvability Validation

Using physics constants from `config/config.go` and derived values from `docs/jump-physics-reference.md`:
- Max jump height: 9.4 tiles (150px)
- Max jump distance: 15 tiles (240px)
- Discover platforms from SolidTiles, build reachability graph via parabolic arc simulation, BFS from start to exit
- Use 85% of max values for comfortable margin
- Remediation: insert bridge chunks or re-select on failure

### 8. Difficulty Scaling

- **Intra-run:** S-curve from difficulty 1 to 5 across the concept graph nodes
- **Inter-run:** Each completed run adds +0.5 to base difficulty, increases enemy budget +15%, increases run length (+2 nodes, capped at 25)

### 9. AI Director / Pacing

Rules enforced during concept graph generation:
- Max 2 consecutive Combat nodes
- BreakRoom after every 3 combat encounters
- Three-act structure: Act 1 (0-30%) = introductory, Act 2 (30-70%) = ramping, Act 3 (70-100%) = climax + resolution
- Arena node at ~75% mark

### 10. Decorative Variation

- Biome-based tileset swaps
- Multiple background variants per biome (3-4)
- Prop scatter at decoration zones defined in chunks
- Subtle color tinting per seed

### 11. Camera Adaptation

Existing camera in `systems/camera.go` already adapts to `level.Width`/`level.Height`. No changes needed for MVP. Post-MVP: add vertical look-ahead for vertical chunk transitions.

### 12. Roguelite Meta-Progression

- Run state tracked via new `RunStateData` component (seed, depth, kills, rooms cleared)
- Persistence via existing `gdata` system -- store run depth, total runs, best rooms, lifetime stats
- Menu integration: new `MainMenuRoguelite` option in `components/menu.go`, handled in `systems/menu.go`

---

## Full System Milestones

| # | Milestone | Depends On | Summary |
|---|-----------|------------|---------|
| 1 | Chunk Infrastructure | -- | Chunk format, loader, 5-8 test chunks, config |
| 2 | Linear Assembly + Compilation | M1 | Assemble chunks left-to-right, compile to `assets.Level`, roguelite scene, menu option |
| 3 | Concept Graph + Enemy/Hazard Placement | M2 | Intelligent room sequencing, dynamic enemy/hazard placement, pacing |
| 4 | Solvability Validation | M3 | Jump arc math, reachability graph, remediation |
| 5 | Decorative Variation | M4 | Prop scatter, background variants, tinting, biome tilesets |
| 6 | Roguelite Meta-Progression | M4 | Run persistence, stats, inter-run difficulty scaling, run summary screen |
| 7 | Advanced Features | M5+M6 | Branching paths, vertical chunks, secret rooms, camera enhancements |

---

## Verification

### How to test end-to-end
1. `make run` -- game launches, main menu shows "Roguelite" option
2. Select Roguelite -- procedurally generated level loads and is playable
3. Player can move, jump, attack, take damage, and reach the finish line
4. Camera follows correctly within generated level bounds
5. Enemies patrol, chase, and attack as in campaign mode
6. Checkpoints save progress within the run
7. Run multiple seeds (change seed in config or via menu) -- layouts differ
8. Test 50+ seeds via automated script to verify no crashes or panics

### Automated testing
- Unit tests for chunk loading (`procgen/chunk_test.go`)
- Unit tests for concept graph generation (`procgen/graph_test.go`)
- Unit tests for assembler connection alignment (`procgen/assembler_test.go`)
- Unit tests for enemy budget calculation and placement validation (`procgen/enemies_test.go`)
- Integration test: generate + compile a level, verify `assets.Level` struct is well-formed
