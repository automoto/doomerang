package procgen

import (
	"math/rand"

	"github.com/automoto/doomerang/config"
)

// NodeType represents the gameplay purpose of a concept graph node
type NodeType string

const (
	NodeStart           NodeType = "start"
	NodeCombat          NodeType = "combat"
	NodeTraversal       NodeType = "traversal"
	NodeBreakRoom       NodeType = "break"
	NodeArena           NodeType = "arena"
	NodeExit            NodeType = "exit"
	NodeVerticalAscent  NodeType = "vertical_ascent"
	NodeVerticalDescent NodeType = "vertical_descent"
	NodeVerticalCombat  NodeType = "vertical_combat"
	NodeTransitionHV    NodeType = "transition_hv"
	NodeTransitionVH    NodeType = "transition_vh"
)

// GraphNode represents a single node in the concept graph
type GraphNode struct {
	Type       NodeType
	Difficulty int
	Biome      string
	Tag        ChunkTag // Required chunk tag for selection
}

// ConceptGraph is an ordered sequence of nodes describing the run structure
type ConceptGraph struct {
	Nodes []GraphNode
}

// GenerateGraph creates a concept graph with pacing rules applied.
// length is the number of middle nodes (excluding start and exit).
func GenerateGraph(rng *rand.Rand, length int, biomes []string) *ConceptGraph {
	if length < 1 {
		length = 1
	}
	totalNodes := length + 2 // start + middle + exit

	nodes := make([]GraphNode, 0, totalNodes)

	// Start node
	nodes = append(nodes, GraphNode{
		Type:       NodeStart,
		Difficulty: 1,
		Biome:      pickBiome(rng, biomes),
		Tag:        TagStart,
	})

	// Generate middle nodes with pacing rules
	combatStreak := 0
	combatsSinceBreak := 0

	for i := 0; i < length; i++ {
		position := i + 1 // 1-indexed (0 is start)
		diff := DifficultyAtNode(position, totalNodes)
		biome := pickBiome(rng, biomes)

		nodeType := pickNodeType(rng, position, length, combatStreak, combatsSinceBreak, diff)

		tag := nodeTypeToTag(nodeType)
		nodes = append(nodes, GraphNode{
			Type:       nodeType,
			Difficulty: diff,
			Biome:      biome,
			Tag:        tag,
		})

		switch nodeType {
		case NodeCombat, NodeArena:
			combatStreak++
			combatsSinceBreak++
		case NodeBreakRoom:
			combatStreak = 0
			combatsSinceBreak = 0
		default:
			combatStreak = 0
		}
	}

	// Exit node
	nodes = append(nodes, GraphNode{
		Type:       NodeExit,
		Difficulty: DifficultyAtNode(totalNodes-1, totalNodes),
		Biome:      pickBiome(rng, biomes),
		Tag:        TagExit,
	})

	graph := &ConceptGraph{Nodes: nodes}
	insertVerticalSections(rng, graph)
	return graph
}

func pickNodeType(rng *rand.Rand, position, middleCount, combatStreak, combatsSinceBreak, difficulty int) NodeType {
	// Forced break room after 3 combat encounters
	if combatsSinceBreak >= 3 {
		return NodeBreakRoom
	}

	// Arena at ~75% mark (high difficulty moment)
	arenaPos := int(float64(middleCount) * 0.75)
	if position == arenaPos && difficulty >= 3 {
		return NodeArena
	}

	// Max 2 consecutive combat nodes
	if combatStreak >= 2 {
		// Must pick non-combat
		choices := []NodeType{NodeTraversal, NodeBreakRoom}
		return choices[rng.Intn(len(choices))]
	}

	// Weighted random selection based on difficulty
	// Higher difficulty = more combat, lower = more traversal/break
	combatWeight := 40 + difficulty*5
	traversalWeight := 30
	breakWeight := 15 - difficulty*2
	if breakWeight < 5 {
		breakWeight = 5
	}

	total := combatWeight + traversalWeight + breakWeight
	roll := rng.Intn(total)

	if roll < combatWeight {
		return NodeCombat
	}
	if roll < combatWeight+traversalWeight {
		return NodeTraversal
	}
	return NodeBreakRoom
}

func nodeTypeToTag(nt NodeType) ChunkTag {
	switch nt {
	case NodeStart:
		return TagStart
	case NodeCombat, NodeArena:
		return TagCombat
	case NodeTraversal:
		return TagTraversal
	case NodeBreakRoom:
		return TagBreak
	case NodeExit:
		return TagExit
	case NodeVerticalAscent:
		return TagVerticalAscent
	case NodeVerticalDescent:
		return TagVerticalDescent
	case NodeVerticalCombat:
		return TagVerticalCombat
	case NodeTransitionHV:
		return TagTransitionHV
	case NodeTransitionVH:
		return TagTransitionVH
	default:
		return TagCombat
	}
}

// insertVerticalSections adds a vertical section into the graph's middle nodes.
// It replaces a subsequence with [TransitionHV, VerticalAscent/Combat/Descent, TransitionVH].
func insertVerticalSections(rng *rand.Rand, graph *ConceptGraph) {
	cfg := config.Procgen
	if cfg.MaxVerticalSectionsPerRun <= 0 {
		return
	}

	middleLen := len(graph.Nodes) - 2 // exclude start and exit
	if middleLen < 3 {
		return // not enough room
	}

	// Determine insertion window (indices into middle nodes, 1-indexed in graph.Nodes)
	minIdx := int(float64(middleLen)*cfg.VerticalInsertionMinPosition) + 1 // +1 for start node
	maxIdx := int(float64(middleLen)*cfg.VerticalInsertionMaxPosition) + 1
	if maxIdx > len(graph.Nodes)-3 { // leave room for 3-node insertion before exit
		maxIdx = len(graph.Nodes) - 3
	}
	if minIdx > maxIdx {
		return
	}

	insertIdx := minIdx + rng.Intn(maxIdx-minIdx+1)

	// Pick vertical node type based on difficulty at that position
	diff := graph.Nodes[insertIdx].Difficulty
	vertType := pickVerticalNodeType(rng, diff)

	biome := graph.Nodes[insertIdx].Biome

	vertNodes := []GraphNode{
		{Type: NodeTransitionHV, Difficulty: diff, Biome: biome, Tag: TagTransitionHV},
		{Type: vertType, Difficulty: diff, Biome: biome, Tag: nodeTypeToTag(vertType)},
		{Type: NodeTransitionVH, Difficulty: diff, Biome: biome, Tag: TagTransitionVH},
	}

	// Replace the node at insertIdx with the vertical section (net +2 nodes)
	newNodes := make([]GraphNode, 0, len(graph.Nodes)+2)
	newNodes = append(newNodes, graph.Nodes[:insertIdx]...)
	newNodes = append(newNodes, vertNodes...)
	newNodes = append(newNodes, graph.Nodes[insertIdx+1:]...)
	graph.Nodes = newNodes
}

func pickVerticalNodeType(rng *rand.Rand, difficulty int) NodeType {
	if difficulty >= 3 {
		// At high difficulty, allow vertical combat
		choices := []NodeType{NodeVerticalAscent, NodeVerticalDescent, NodeVerticalCombat}
		return choices[rng.Intn(len(choices))]
	}
	choices := []NodeType{NodeVerticalAscent, NodeVerticalDescent}
	return choices[rng.Intn(len(choices))]
}

func pickBiome(rng *rand.Rand, biomes []string) string {
	if len(biomes) == 0 {
		return "cyberpunk"
	}
	return biomes[rng.Intn(len(biomes))]
}
