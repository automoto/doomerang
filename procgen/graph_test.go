package procgen_test

import (
	"math/rand"
	"testing"

	"github.com/automoto/doomerang/procgen"
)

func TestGraphNodeCount(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	graph := procgen.GenerateGraph(rng, 10, []string{"cyberpunk"})

	// 10 middle + start + exit = 12
	if len(graph.Nodes) != 12 {
		t.Errorf("expected 12 nodes, got %d", len(graph.Nodes))
	}
}

func TestGraphStartAndExit(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	graph := procgen.GenerateGraph(rng, 8, []string{"cyberpunk"})

	if graph.Nodes[0].Type != procgen.NodeStart {
		t.Errorf("first node should be Start, got %s", graph.Nodes[0].Type)
	}
	last := graph.Nodes[len(graph.Nodes)-1]
	if last.Type != procgen.NodeExit {
		t.Errorf("last node should be Exit, got %s", last.Type)
	}
}

func TestGraphPacingMaxCombatStreak(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		rng := rand.New(rand.NewSource(seed))
		graph := procgen.GenerateGraph(rng, 10, []string{"cyberpunk"})
		procgen.ValidateGraph(graph)

		combatStreak := 0
		for _, node := range graph.Nodes {
			if node.Type == procgen.NodeCombat || node.Type == procgen.NodeArena {
				combatStreak++
				if combatStreak > 2 {
					t.Errorf("seed %d: combat streak exceeded 2", seed)
					break
				}
			} else {
				combatStreak = 0
			}
		}
	}
}

func TestGraphPacingBreakAfterCombats(t *testing.T) {
	for seed := int64(0); seed < 50; seed++ {
		rng := rand.New(rand.NewSource(seed))
		graph := procgen.GenerateGraph(rng, 12, []string{"cyberpunk"})
		procgen.ValidateGraph(graph)

		combatsSinceBreak := 0
		for _, node := range graph.Nodes[1 : len(graph.Nodes)-1] {
			switch node.Type {
			case procgen.NodeCombat, procgen.NodeArena:
				combatsSinceBreak++
				if combatsSinceBreak > 3 {
					t.Errorf("seed %d: more than 3 combats without break", seed)
				}
			case procgen.NodeBreakRoom:
				combatsSinceBreak = 0
			}
		}
	}
}

func TestGraphDifficultyProgression(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	graph := procgen.GenerateGraph(rng, 10, []string{"cyberpunk"})

	first := graph.Nodes[1] // first middle node
	last := graph.Nodes[len(graph.Nodes)-2] // last middle node

	if last.Difficulty < first.Difficulty {
		t.Errorf("expected difficulty to increase: first=%d, last=%d",
			first.Difficulty, last.Difficulty)
	}
}

func TestGraphNodeTypes(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	graph := procgen.GenerateGraph(rng, 10, []string{"cyberpunk"})

	typeCount := make(map[procgen.NodeType]int)
	for _, node := range graph.Nodes {
		typeCount[node.Type]++
	}

	if typeCount[procgen.NodeStart] != 1 {
		t.Errorf("expected 1 start node, got %d", typeCount[procgen.NodeStart])
	}
	if typeCount[procgen.NodeExit] != 1 {
		t.Errorf("expected 1 exit node, got %d", typeCount[procgen.NodeExit])
	}
	if typeCount[procgen.NodeCombat]+typeCount[procgen.NodeArena] == 0 {
		t.Error("expected at least 1 combat/arena node")
	}
}

func TestAssembleFromGraph(t *testing.T) {
	chunks := loadTestChunks(t)
	rng := rand.New(rand.NewSource(42))
	graph := procgen.GenerateGraph(rng, 5, []string{"cyberpunk"})
	procgen.ValidateGraph(graph)

	assembler := procgen.NewAssembler(42)
	result, err := assembler.AssembleFromGraph(chunks, graph)
	if err != nil {
		t.Fatalf("AssembleFromGraph failed: %v", err)
	}

	if len(result.PlacedChunks) != len(graph.Nodes) {
		t.Errorf("expected %d placed chunks, got %d", len(graph.Nodes), len(result.PlacedChunks))
	}

	// First should be start, last should be exit
	if !result.PlacedChunks[0].Chunk.HasTag(procgen.TagStart) {
		t.Error("first chunk should be start")
	}
	last := result.PlacedChunks[len(result.PlacedChunks)-1]
	if !last.Chunk.HasTag(procgen.TagExit) {
		t.Error("last chunk should be exit")
	}
}

func TestAssembleFromGraphChunkReuse(t *testing.T) {
	chunks := loadTestChunks(t)

	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		graph := procgen.GenerateGraph(rng, 8, []string{"cyberpunk"})
		procgen.ValidateGraph(graph)

		assembler := procgen.NewAssembler(seed)
		result, err := assembler.AssembleFromGraph(chunks, graph)
		if err != nil {
			continue // Some seeds may fail with limited chunks
		}

		usage := make(map[string]int)
		for _, pc := range result.PlacedChunks {
			usage[pc.Chunk.ID]++
			if usage[pc.Chunk.ID] > 2 {
				t.Errorf("seed %d: chunk %s used %d times (max 2)",
					seed, pc.Chunk.ID, usage[pc.Chunk.ID])
			}
		}
	}
}
