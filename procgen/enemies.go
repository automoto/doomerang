package procgen

import (
	"math/rand"
	"sort"

	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/config"
)

// EnemyPlacer handles dynamic enemy placement within chunks
type EnemyPlacer struct {
	rng *rand.Rand
}

// NewEnemyPlacer creates an enemy placer with the given RNG
func NewEnemyPlacer(rng *rand.Rand) *EnemyPlacer {
	return &EnemyPlacer{rng: rng}
}

// PlaceEnemies generates enemy spawns for a placed chunk based on difficulty.
// Returns enemy spawns in world-space coordinates.
func (ep *EnemyPlacer) PlaceEnemies(pc PlacedChunk, difficulty int) []assets.EnemySpawn {
	chunk := pc.Chunk
	if chunk.MaxEnemies == 0 {
		return nil
	}

	// Calculate budget for this room
	budget := config.Procgen.EnemyBudgetBase + float64(difficulty)*config.Procgen.EnemyBudgetMultiplier

	// Determine enemy count within chunk's min/max
	count := ep.enemyCountFromBudget(budget, chunk.MinEnemies, chunk.MaxEnemies)
	if count == 0 {
		return nil
	}

	// Find platforms to place enemies on
	platforms := discoverPlatforms(chunk)
	if len(platforms) == 0 {
		return nil
	}

	// Select enemy types based on difficulty
	types := ep.selectEnemyTypes(count, difficulty, platforms)

	// Distribute enemies across platforms
	return ep.distributeEnemies(types, platforms, pc)
}

func (ep *EnemyPlacer) enemyCountFromBudget(budget float64, minE, maxE int) int {
	// Use average cost to estimate count from budget
	avgCost := 3.0 // Guard cost as baseline
	count := int(budget / avgCost)

	if count < minE {
		count = minE
	}
	if count > maxE {
		count = maxE
	}
	return count
}

func (ep *EnemyPlacer) selectEnemyTypes(count, difficulty int, platforms []platform) []string {
	types := make([]string, 0, count)

	hasElevated := false
	for _, p := range platforms {
		if p.elevated {
			hasElevated = true
			break
		}
	}

	for i := 0; i < count; i++ {
		types = append(types, ep.pickEnemyType(difficulty, hasElevated))
	}
	return types
}

func (ep *EnemyPlacer) pickEnemyType(difficulty int, hasElevated bool) string {
	// Build weighted pool based on difficulty
	type choice struct {
		name   string
		weight int
	}

	pool := []choice{
		{"LightGuard", 30 - difficulty*3},
		{"Guard", 30},
	}

	if difficulty >= 2 {
		pool = append(pool, choice{"HeavyGuard", difficulty * 5})
	}
	if difficulty >= 2 && hasElevated {
		pool = append(pool, choice{"KnifeThrower", difficulty * 4})
	}

	// Clamp weights
	total := 0
	for i := range pool {
		if pool[i].weight < 5 {
			pool[i].weight = 5
		}
		total += pool[i].weight
	}

	roll := ep.rng.Intn(total)
	cumulative := 0
	for _, c := range pool {
		cumulative += c.weight
		if roll < cumulative {
			return c.name
		}
	}
	return "Guard"
}

type platform struct {
	x, y, width float64
	elevated    bool // true if not at floor level
}

func discoverPlatforms(chunk *Chunk) []platform {
	if len(chunk.EnemySlots) > 0 {
		return platformsFromSlots(chunk)
	}
	return platformsFromTiles(chunk)
}

// platformsFromSlots converts EnemySlot objects into platforms
func platformsFromSlots(chunk *Chunk) []platform {
	floorY := float64(chunk.Height) - 48 // 3 tiles from bottom
	platforms := make([]platform, 0, len(chunk.EnemySlots))
	for _, slot := range chunk.EnemySlots {
		platforms = append(platforms, platform{
			x:        slot.X,
			y:        slot.Y,
			width:    slot.PlatformWidth,
			elevated: slot.Y < floorY,
		})
	}
	return platforms
}

// platformsFromTiles discovers horizontal platform surfaces from solid tiles
func platformsFromTiles(chunk *Chunk) []platform {
	// Group solid tiles by row, find contiguous horizontal runs
	type tilePos struct{ x, y int }
	occupied := make(map[tilePos]bool)
	tileW := 16.0
	tileH := 16.0
	if chunk.TiledMap != nil {
		tileW = float64(chunk.TiledMap.TileWidth)
		tileH = float64(chunk.TiledMap.TileHeight)
	}

	for _, t := range chunk.SolidTiles {
		col := int(t.X / tileW)
		row := int(t.Y / tileH)
		occupied[tilePos{col, row}] = true
	}

	// Find surface tiles (solid tile with empty tile above)
	type surfaceTile struct{ col, row int }
	var surfaces []surfaceTile
	for pos := range occupied {
		above := tilePos{pos.x, pos.y - 1}
		if !occupied[above] {
			surfaces = append(surfaces, surfaceTile{pos.x, pos.y})
		}
	}

	// Sort by row then column
	sort.Slice(surfaces, func(i, j int) bool {
		if surfaces[i].row != surfaces[j].row {
			return surfaces[i].row < surfaces[j].row
		}
		return surfaces[i].col < surfaces[j].col
	})

	// Group into contiguous runs
	floorRow := chunk.TileHeight - 3 // Bottom 3 rows are floor/fill
	var platforms []platform
	for i := 0; i < len(surfaces); {
		j := i + 1
		for j < len(surfaces) && surfaces[j].row == surfaces[i].row && surfaces[j].col == surfaces[j-1].col+1 {
			j++
		}

		// Create platform from run
		runLen := j - i
		if runLen >= 2 { // At least 2 tiles wide
			px := float64(surfaces[i].col) * tileW
			py := float64(surfaces[i].row) * tileH
			pw := float64(runLen) * tileW
			platforms = append(platforms, platform{
				x:        px,
				y:        py,
				width:    pw,
				elevated: surfaces[i].row < floorRow,
			})
		}
		i = j
	}

	return platforms
}

func (ep *EnemyPlacer) distributeEnemies(types []string, platforms []platform, pc PlacedChunk) []assets.EnemySpawn {
	var spawns []assets.EnemySpawn
	minSpacing := config.Procgen.EnemyMinSpacing
	usedPositions := make([]float64, 0) // track X positions for spacing

	// Sort platforms by width (largest first) for better distribution
	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].width > platforms[j].width
	})

	for _, enemyType := range types {
		// KnifeThrowers prefer elevated platforms
		isKnifeThrower := enemyType == "KnifeThrower"

		placed := false
		for _, p := range platforms {
			if isKnifeThrower && !p.elevated {
				continue
			}

			// Try to find a valid position on this platform
			x := ep.findSpawnX(p, usedPositions, minSpacing)
			if x < 0 {
				continue
			}

			// Enemy collision height is ~40px, spawn above the platform surface
			spawnY := p.y - 40

			spawns = append(spawns, assets.EnemySpawn{
				X:         x + pc.OffsetX,
				Y:         spawnY + pc.OffsetY,
				EnemyType: enemyType,
			})
			usedPositions = append(usedPositions, x)
			placed = true
			break
		}

		// If KnifeThrower couldn't find elevated platform, place on ground
		if !placed && isKnifeThrower {
			for _, p := range platforms {
				x := ep.findSpawnX(p, usedPositions, minSpacing)
				if x < 0 {
					continue
				}
				spawnY := p.y - 40
				spawns = append(spawns, assets.EnemySpawn{
					X:         x + pc.OffsetX,
					Y:         spawnY + pc.OffsetY,
					EnemyType: "Guard", // Downgrade to Guard on ground
				})
				usedPositions = append(usedPositions, x)
				break
			}
		}
	}

	return spawns
}

func (ep *EnemyPlacer) findSpawnX(p platform, used []float64, minSpacing float64) float64 {
	// Try random positions within the platform
	margin := 16.0 // Keep away from platform edges
	availWidth := p.width - margin*2
	if availWidth < 16 {
		return -1
	}

	for attempt := 0; attempt < 10; attempt++ {
		x := p.x + margin + ep.rng.Float64()*availWidth

		valid := true
		for _, ux := range used {
			if abs64(x-ux) < minSpacing {
				valid = false
				break
			}
		}
		if valid {
			return x
		}
	}
	return -1
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
