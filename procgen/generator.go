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

// GenerationResult contains the output of chunk generation
type GenerationResult struct {
	PlacedChunks []PlacedChunk
	TotalWidth   int
	TotalHeight  int
}

// maxChunkReuse is the maximum number of times the same chunk ID may appear in one run.
const maxChunkReuse = 2

// ChunkGenerator chains chunks using directional connection matching
type ChunkGenerator struct {
	rng *rand.Rand
}

// NewChunkGenerator creates a chunk generator with the given random seed
func NewChunkGenerator(seed int64) *ChunkGenerator {
	return &ChunkGenerator{rng: rand.New(rand.NewSource(seed))}
}

// Generate selects and places chunks left-to-right.
// It picks a start chunk, middle chunks, and an exit chunk from the pool.
func (g *ChunkGenerator) Generate(chunks []*Chunk, numMiddle int) (*GenerationResult, error) {
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
	sequence = append(sequence, starts[g.rng.Intn(len(starts))])

	for i := 0; i < numMiddle; i++ {
		sequence = append(sequence, middles[g.rng.Intn(len(middles))])
	}

	sequence = append(sequence, exits[g.rng.Intn(len(exits))])

	return g.placeChunks(sequence)
}

// GenerateFromGraph selects chunks matching concept graph node requirements
// and places them left-to-right. Enforces max 2 reuses of the same chunk ID.
func (g *ChunkGenerator) GenerateFromGraph(chunks []*Chunk, graph *ConceptGraph) (*GenerationResult, error) {
	usageCount := make(map[string]int)
	sequence := make([]*Chunk, 0, len(graph.Nodes))

	for _, node := range graph.Nodes {
		candidates := g.matchChunks(chunks, node, usageCount)
		if len(candidates) == 0 {
			// Fallback: relax biome constraint
			candidates = g.matchChunksRelaxed(chunks, node, usageCount)
		}
		if len(candidates) == 0 {
			return nil, fmt.Errorf("no chunk found for node type=%s tag=%s biome=%s", node.Type, node.Tag, node.Biome)
		}

		selected := candidates[g.rng.Intn(len(candidates))]
		sequence = append(sequence, selected)
		usageCount[selected.ID]++
	}

	return g.placeChunks(sequence)
}

// requiredEdges returns the connection edges a chunk must have for its tag type.
func requiredEdges(tag ChunkTag) (required []ConnectionEdge) {
	switch tag {
	case TagStart:
		return []ConnectionEdge{EdgeRight}
	case TagExit:
		return []ConnectionEdge{EdgeLeft}
	case TagTransitionHV:
		return []ConnectionEdge{EdgeLeft, EdgeBottom}
	case TagTransitionVH:
		return []ConnectionEdge{EdgeTop, EdgeRight}
	case TagVerticalAscent, TagVerticalDescent, TagVerticalCombat:
		return []ConnectionEdge{EdgeTop, EdgeBottom}
	default:
		// Standard horizontal chunks
		return []ConnectionEdge{EdgeLeft, EdgeRight}
	}
}

// matchChunks finds chunks that match a graph node's requirements
func (g *ChunkGenerator) matchChunks(chunks []*Chunk, node GraphNode, usage map[string]int) []*Chunk {
	edges := requiredEdges(node.Tag)
	var result []*Chunk
	for _, c := range chunks {
		if !c.HasTag(node.Tag) {
			continue
		}
		if usage[c.ID] >= maxChunkReuse {
			continue
		}
		if node.Biome != "" && c.Biome != node.Biome {
			continue
		}
		if !hasAllEdges(c, edges) {
			continue
		}
		result = append(result, c)
	}
	return result
}

// matchChunksRelaxed finds chunks matching tag only (ignoring biome)
func (g *ChunkGenerator) matchChunksRelaxed(chunks []*Chunk, node GraphNode, usage map[string]int) []*Chunk {
	edges := requiredEdges(node.Tag)
	var result []*Chunk
	for _, c := range chunks {
		if !c.HasTag(node.Tag) {
			continue
		}
		if usage[c.ID] >= maxChunkReuse {
			continue
		}
		if !hasAllEdges(c, edges) {
			continue
		}
		result = append(result, c)
	}
	return result
}

func hasAllEdges(c *Chunk, edges []ConnectionEdge) bool {
	for _, e := range edges {
		if len(c.GetConnections(e)) == 0 {
			return false
		}
	}
	return true
}

// FindMatchingEdges determines which edges to use for connecting two adjacent chunks.
// It tries Right→Left first (horizontal), then Bottom→Top (downward vertical),
// then Top→Bottom (upward vertical).
func FindMatchingEdges(prev, curr *Chunk) (prevEdge, currEdge ConnectionEdge, err error) {
	// Horizontal: Right → Left
	if len(prev.GetConnections(EdgeRight)) > 0 && len(curr.GetConnections(EdgeLeft)) > 0 {
		return EdgeRight, EdgeLeft, nil
	}
	// Vertical down: Bottom → Top
	if len(prev.GetConnections(EdgeBottom)) > 0 && len(curr.GetConnections(EdgeTop)) > 0 {
		return EdgeBottom, EdgeTop, nil
	}
	// Vertical up: Top → Bottom
	if len(prev.GetConnections(EdgeTop)) > 0 && len(curr.GetConnections(EdgeBottom)) > 0 {
		return EdgeTop, EdgeBottom, nil
	}
	return "", "", fmt.Errorf("no compatible edges between chunk %q and %q", prev.ID, curr.ID)
}

// placeChunks positions chunks using directional connection matching
func (g *ChunkGenerator) placeChunks(sequence []*Chunk) (*GenerationResult, error) {
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

		prevEdge, currEdge, err := FindMatchingEdges(prev.Chunk, curr)
		if err != nil {
			return nil, err
		}

		prevConn := prev.Chunk.GetConnections(prevEdge)[0]
		currConn := curr.GetConnections(currEdge)[0]

		var offsetX, offsetY float64

		switch {
		case prevEdge == EdgeRight && currEdge == EdgeLeft:
			// Horizontal: place to the right, align Y via connection YOffset
			offsetX = prev.OffsetX + float64(prev.Chunk.Width)
			offsetY = prev.OffsetY + prevConn.YOffset - currConn.YOffset

		case prevEdge == EdgeBottom && currEdge == EdgeTop:
			// Vertical down: place below, align X via connection XOffset
			offsetY = prev.OffsetY + float64(prev.Chunk.Height)
			offsetX = prev.OffsetX + prevConn.XOffset - currConn.XOffset

		case prevEdge == EdgeTop && currEdge == EdgeBottom:
			// Vertical up: place above, align X via connection XOffset
			offsetY = prev.OffsetY - float64(curr.Height)
			offsetX = prev.OffsetX + prevConn.XOffset - currConn.XOffset
		}

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
	headroom := screenHeight * config.Procgen.ChunkHeadroomFactor
	for i := range placed {
		placed[i].OffsetY += headroom
	}
	maxBottom += headroom

	return &GenerationResult{
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
