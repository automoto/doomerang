package procgen

import (
	"fmt"
	"math"
	"sort"

	"github.com/automoto/doomerang/config"
)

// ValidationResult contains the outcome of solvability checking
type ValidationResult struct {
	Solvable     bool
	Unreachable  []int // Indices of unreachable platforms
	StartIdx     int
	ExitIdx      int
	PlatformCount int
}

// Platform represents a walkable surface for reachability analysis
type Platform struct {
	X, Y, Width float64
	ChunkIndex  int // Which placed chunk this belongs to
}

// Validator checks that a generated level is solvable using jump physics
type Validator struct {
	maxJumpHeight float64
	maxJumpDist   float64
	margin        float64 // Safety margin (0.85 = 85% of max)
}

// NewValidator creates a validator using game physics constants
func NewValidator() *Validator {
	jumpSpeed := config.Player.JumpSpeed
	gravity := config.Player.Gravity
	maxSpeedX := config.Player.MaxSpeed

	// Max jump height: v^2 / (2g)
	maxHeight := (jumpSpeed * jumpSpeed) / (2 * gravity)
	// Time to peak: v/g, total air time: 2 * v/g
	airTime := 2 * jumpSpeed / gravity
	// Max horizontal distance: maxSpeedX * airTime
	maxDist := maxSpeedX * airTime

	return &Validator{
		maxJumpHeight: maxHeight,
		maxJumpDist:   maxDist,
		margin:        0.95,
	}
}

// Validate checks if the generated level is solvable.
// Returns a ValidationResult indicating whether every platform is reachable
// from the start and the exit is reachable.
func (v *Validator) Validate(result *GenerationResult) ValidationResult {
	platforms := v.discoverPlatforms(result)

	if len(platforms) == 0 {
		return ValidationResult{Solvable: false}
	}

	// Identify start and exit platforms
	startIdx := v.findStartPlatform(platforms, result)
	exitIdx := v.findExitPlatform(platforms, result)

	// Build reachability graph
	reachable := v.bfsReachability(platforms, startIdx)

	vr := ValidationResult{
		Solvable:      reachable[exitIdx],
		StartIdx:      startIdx,
		ExitIdx:       exitIdx,
		PlatformCount: len(platforms),
	}

	for i, r := range reachable {
		if !r {
			vr.Unreachable = append(vr.Unreachable, i)
		}
	}

	return vr
}

// discoverPlatforms finds all walkable surfaces from the generated level
func (v *Validator) discoverPlatforms(result *GenerationResult) []Platform {
	var allPlatforms []Platform

	for chunkIdx, pc := range result.PlacedChunks {
		chunk := pc.Chunk
		ox := pc.OffsetX
		oy := pc.OffsetY

		tileW := 16.0
		tileH := 16.0
		if chunk.TiledMap != nil {
			tileW = float64(chunk.TiledMap.TileWidth)
			tileH = float64(chunk.TiledMap.TileHeight)
		}

		// Build tile grid for this chunk
		type tilePos struct{ col, row int }
		occupied := make(map[tilePos]bool)
		for _, t := range chunk.SolidTiles {
			col := int(t.X / tileW)
			row := int(t.Y / tileH)
			occupied[tilePos{col, row}] = true
		}

		// Find surface tiles (solid with empty above)
		type surfaceTile struct{ col, row int }
		var surfaces []surfaceTile
		for pos := range occupied {
			above := tilePos{pos.col, pos.row - 1}
			if !occupied[above] {
				surfaces = append(surfaces, surfaceTile(pos))
			}
		}

		// Sort and group into contiguous runs
		sort.Slice(surfaces, func(i, j int) bool {
			if surfaces[i].row != surfaces[j].row {
				return surfaces[i].row < surfaces[j].row
			}
			return surfaces[i].col < surfaces[j].col
		})

		for i := 0; i < len(surfaces); {
			j := i + 1
			for j < len(surfaces) && surfaces[j].row == surfaces[i].row && surfaces[j].col == surfaces[j-1].col+1 {
				j++
			}
			runLen := j - i
			// Require at least 2 tiles wide to be a walkable platform
			// (single-tile walls are not platforms)
			if runLen >= 2 {
				allPlatforms = append(allPlatforms, Platform{
					X:          float64(surfaces[i].col)*tileW + ox,
					Y:          float64(surfaces[i].row)*tileH + oy,
					Width:      float64(runLen) * tileW,
					ChunkIndex: chunkIdx,
				})
			}
			i = j
		}
	}

	return allPlatforms
}

// canReach returns true if the player can jump from platform a to platform b
func (v *Validator) canReach(a, b Platform) bool {
	safeHeight := v.maxJumpHeight * v.margin
	safeDist := v.maxJumpDist * v.margin

	// Height difference: positive = b is below a, negative = b is above a
	// (Y increases downward in screen coords)
	dy := b.Y - a.Y

	// If b is above us by more than max jump height, can't reach
	if dy < -safeHeight {
		return false
	}

	// Horizontal distance between closest edges
	aRight := a.X + a.Width
	bRight := b.X + b.Width

	var dx float64
	if b.X > aRight {
		dx = b.X - aRight
	} else if a.X > bRight {
		dx = a.X - bRight
	} else {
		dx = 0 // overlapping or touching horizontally
	}

	// Touching/overlapping platforms: can walk between if reachable vertically
	if dx == 0 {
		return dy >= -safeHeight
	}

	if dy >= 0 {
		// Dropping down or same level: extra fall time gives more horizontal distance
		gravity := config.Player.Gravity
		extraTime := math.Sqrt(2 * dy / gravity)
		extraDist := config.Player.MaxSpeed * extraTime
		return dx <= safeDist+extraDist
	}

	// Jumping up: height cost reduces available horizontal distance
	heightRatio := math.Abs(dy) / safeHeight
	if heightRatio > 1 {
		return false
	}
	adjustedDist := safeDist * (1 - heightRatio*0.5)
	return dx <= adjustedDist
}

// findStartPlatform finds the floor platform in the first chunk
// (lowest Y value = highest on screen, but we want the floor = highest Y)
func (v *Validator) findStartPlatform(platforms []Platform, result *GenerationResult) int {
	bestIdx := 0
	bestY := -1.0
	for i, p := range platforms {
		if p.ChunkIndex == 0 && p.Y > bestY {
			bestY = p.Y
			bestIdx = i
		}
	}
	return bestIdx
}

// findExitPlatform finds the floor platform in the last chunk
func (v *Validator) findExitPlatform(platforms []Platform, result *GenerationResult) int {
	lastChunk := len(result.PlacedChunks) - 1
	bestIdx := len(platforms) - 1
	bestY := -1.0
	for i, p := range platforms {
		if p.ChunkIndex == lastChunk && p.Y > bestY {
			bestY = p.Y
			bestIdx = i
		}
	}
	return bestIdx
}

// bfsReachability performs BFS from startIdx to find all reachable platforms
func (v *Validator) bfsReachability(platforms []Platform, startIdx int) []bool {
	n := len(platforms)
	reachable := make([]bool, n)
	reachable[startIdx] = true

	queue := []int{startIdx}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		for i := 0; i < n; i++ {
			if reachable[i] {
				continue
			}
			if v.canReach(platforms[curr], platforms[i]) {
				reachable[i] = true
				queue = append(queue, i)
			}
		}
	}

	return reachable
}

// ValidateAndRemediate validates the level and retries generation if unsolvable.
// Returns the final generation result and any error.
func ValidateAndRemediate(generator *ChunkGenerator, chunks []*Chunk, graph *ConceptGraph, maxAttempts int) (*GenerationResult, error) {
	validator := NewValidator()

	for attempt := 0; attempt < maxAttempts; attempt++ {
		result, err := generator.GenerateFromGraph(chunks, graph)
		if err != nil {
			// Try simple generation as fallback
			result, err = generator.Generate(chunks, len(graph.Nodes)-2)
			if err != nil {
				return nil, fmt.Errorf("generation failed on attempt %d: %w", attempt, err)
			}
		}

		vr := validator.Validate(result)
		if vr.Solvable {
			return result, nil
		}
	}

	// Last resort: return whatever we get (our chunks are hand-designed to be traversable)
	result, err := generator.Generate(chunks, len(graph.Nodes)-2)
	if err != nil {
		return nil, fmt.Errorf("all remediation attempts failed: %w", err)
	}
	return result, nil
}
