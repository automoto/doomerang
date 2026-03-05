package procgen

import (
	"fmt"
	"math/rand"

	"github.com/automoto/doomerang/config"
)

// PlacedChunk represents a chunk positioned in world space
type PlacedChunk struct {
	Chunk   *Chunk
	OffsetX float64 // World X offset (pixels)
	OffsetY float64 // World Y offset to align connection points
}

// AssemblyResult contains the output of chunk assembly
type AssemblyResult struct {
	PlacedChunks []PlacedChunk
	TotalWidth   int
	TotalHeight  int
}

// Assembler chains chunks left-to-right using connection points
type Assembler struct {
	rng *rand.Rand
}

// NewAssembler creates an assembler with the given random seed
func NewAssembler(seed int64) *Assembler {
	return &Assembler{rng: rand.New(rand.NewSource(seed))}
}

// Assemble selects and places chunks left-to-right.
// It picks a start chunk, middle chunks, and an exit chunk from the pool.
func (a *Assembler) Assemble(chunks []*Chunk, numMiddle int) (*AssemblyResult, error) {
	starts := filterByTag(chunks, TagStart)
	exits := filterByTag(chunks, TagExit)
	middles := filterMiddle(chunks)

	if len(starts) == 0 {
		return nil, fmt.Errorf("no start chunks available")
	}
	if len(exits) == 0 {
		return nil, fmt.Errorf("no exit chunks available")
	}
	if len(middles) == 0 {
		return nil, fmt.Errorf("no middle chunks available")
	}

	// Build the sequence: start + N middles + exit
	sequence := make([]*Chunk, 0, numMiddle+2)
	sequence = append(sequence, starts[a.rng.Intn(len(starts))])

	for i := 0; i < numMiddle; i++ {
		sequence = append(sequence, middles[a.rng.Intn(len(middles))])
	}

	sequence = append(sequence, exits[a.rng.Intn(len(exits))])

	return a.placeChunks(sequence)
}

// AssembleFromGraph selects chunks matching concept graph node requirements
// and places them left-to-right. Enforces max 2 reuses of the same chunk ID.
func (a *Assembler) AssembleFromGraph(chunks []*Chunk, graph *ConceptGraph) (*AssemblyResult, error) {
	usageCount := make(map[string]int)
	sequence := make([]*Chunk, 0, len(graph.Nodes))

	for _, node := range graph.Nodes {
		candidates := a.matchChunks(chunks, node, usageCount)
		if len(candidates) == 0 {
			// Fallback: relax biome constraint
			candidates = a.matchChunksRelaxed(chunks, node, usageCount)
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("no chunk found for node type=%s tag=%s biome=%s", node.Type, node.Tag, node.Biome)
		}

		selected := candidates[a.rng.Intn(len(candidates))]
		sequence = append(sequence, selected)
		usageCount[selected.ID]++
	}

	return a.placeChunks(sequence)
}

// matchChunks finds chunks that match a graph node's requirements
func (a *Assembler) matchChunks(chunks []*Chunk, node GraphNode, usage map[string]int) []*Chunk {
	var result []*Chunk
	for _, c := range chunks {
		if !c.HasTag(node.Tag) {
			continue
		}
		if usage[c.ID] >= 2 {
			continue
		}
		if node.Biome != "" && c.Biome != node.Biome {
			continue
		}
		// For non-start/exit, require both connections
		if node.Tag != TagStart && node.Tag != TagExit {
			if len(c.GetConnections(EdgeLeft)) == 0 || len(c.GetConnections(EdgeRight)) == 0 {
				continue
			}
		}
		result = append(result, c)
	}
	return result
}

// matchChunksRelaxed finds chunks matching tag only (ignoring biome)
func (a *Assembler) matchChunksRelaxed(chunks []*Chunk, node GraphNode, usage map[string]int) []*Chunk {
	var result []*Chunk
	for _, c := range chunks {
		if !c.HasTag(node.Tag) {
			continue
		}
		if usage[c.ID] >= 2 {
			continue
		}
		if node.Tag != TagStart && node.Tag != TagExit {
			if len(c.GetConnections(EdgeLeft)) == 0 || len(c.GetConnections(EdgeRight)) == 0 {
				continue
			}
		}
		result = append(result, c)
	}
	return result
}

// placeChunks positions chunks left-to-right, aligning connection points vertically
func (a *Assembler) placeChunks(sequence []*Chunk) (*AssemblyResult, error) {
	if len(sequence) == 0 {
		return nil, fmt.Errorf("empty chunk sequence")
	}

	placed := make([]PlacedChunk, len(sequence))

	// Place first chunk at origin
	placed[0] = PlacedChunk{
		Chunk:   sequence[0],
		OffsetX: 0,
		OffsetY: 0,
	}

	for i := 1; i < len(sequence); i++ {
		prev := placed[i-1]
		curr := sequence[i]

		// Find matching connection points
		rightConns := prev.Chunk.GetConnections(EdgeRight)
		leftConns := curr.GetConnections(EdgeLeft)

		if len(rightConns) == 0 {
			return nil, fmt.Errorf("chunk %q has no right connection", prev.Chunk.ID)
		}
		if len(leftConns) == 0 {
			return nil, fmt.Errorf("chunk %q has no left connection", curr.ID)
		}

		// Use slot 0 connections for alignment
		prevConn := rightConns[0]
		currConn := leftConns[0]

		// X: place immediately after previous chunk
		offsetX := prev.OffsetX + float64(prev.Chunk.Width)

		// Y: align connection point Y positions
		// prev connection world Y = prev.OffsetY + prevConn.YOffset
		// curr connection world Y = offsetY + currConn.YOffset
		// We want them equal: offsetY = prev.OffsetY + prevConn.YOffset - currConn.YOffset
		offsetY := prev.OffsetY + prevConn.YOffset - currConn.YOffset

		placed[i] = PlacedChunk{
			Chunk:   curr,
			OffsetX: offsetX,
			OffsetY: offsetY,
		}
	}

	// Calculate total dimensions
	var maxRight float64
	var minY, maxBottom float64
	for i, p := range placed {
		right := p.OffsetX + float64(p.Chunk.Width)
		bottom := p.OffsetY + float64(p.Chunk.Height)
		if i == 0 || right > maxRight {
			maxRight = right
		}
		if i == 0 || p.OffsetY < minY {
			minY = p.OffsetY
		}
		if i == 0 || bottom > maxBottom {
			maxBottom = bottom
		}
	}

	// Normalize Y offsets so minimum is 0
	if minY < 0 {
		for i := range placed {
			placed[i].OffsetY -= minY
		}
		maxBottom -= minY
	}

	// Add minimal headroom above chunks so the camera can follow jumps
	// without clipping the player off-screen. Less headroom = camera
	// stays lower = player sees more ground below them during jumps.
	// The camera's minCameraY clamp (screenHeight/2) prevents the
	// player from going off the top of the screen.
	screenHeight := float64(config.C.Height)
	headroom := screenHeight * 0.15
	for i := range placed {
		placed[i].OffsetY += headroom
	}
	maxBottom += headroom

	return &AssemblyResult{
		PlacedChunks: placed,
		TotalWidth:   int(maxRight),
		TotalHeight:  int(maxBottom),
	}, nil
}

func filterByTag(chunks []*Chunk, tag ChunkTag) []*Chunk {
	var result []*Chunk
	for _, c := range chunks {
		if c.HasTag(tag) {
			result = append(result, c)
		}
	}
	return result
}

func filterMiddle(chunks []*Chunk) []*Chunk {
	var result []*Chunk
	for _, c := range chunks {
		if c.HasTag(TagStart) || c.HasTag(TagExit) {
			continue
		}
		// Must have both left and right connections
		if len(c.GetConnections(EdgeLeft)) > 0 && len(c.GetConnections(EdgeRight)) > 0 {
			result = append(result, c)
		}
	}
	return result
}
