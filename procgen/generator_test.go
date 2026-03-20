package procgen_test

import (
	"testing"

	"github.com/automoto/doomerang/procgen"
)

func TestGenerateBasic(t *testing.T) {
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	result, err := gen.Generate(chunks, 3)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Should have start + 3 middle + exit = 5 chunks
	if len(result.PlacedChunks) != 5 {
		t.Errorf("expected 5 placed chunks, got %d", len(result.PlacedChunks))
	}

	// First chunk should be a start
	if !result.PlacedChunks[0].Chunk.HasTag(procgen.TagStart) {
		t.Error("first chunk should have start tag")
	}

	// Last chunk should be an exit
	last := result.PlacedChunks[len(result.PlacedChunks)-1]
	if !last.Chunk.HasTag(procgen.TagExit) {
		t.Error("last chunk should have exit tag")
	}
}

func TestGenerateAlignment(t *testing.T) {
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	result, err := gen.Generate(chunks, 2)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify chunks don't overlap horizontally
	for i := 1; i < len(result.PlacedChunks); i++ {
		prev := result.PlacedChunks[i-1]
		curr := result.PlacedChunks[i]

		prevRight := prev.OffsetX + float64(prev.Chunk.Width)
		if curr.OffsetX < prevRight-0.01 {
			t.Errorf("chunk %d overlaps chunk %d: prev right=%v, curr left=%v",
				i, i-1, prevRight, curr.OffsetX)
		}
		if curr.OffsetX > prevRight+0.01 {
			t.Errorf("gap between chunk %d and %d: prev right=%v, curr left=%v",
				i-1, i, prevRight, curr.OffsetX)
		}
	}
}

func TestGenerateConnectionAlignment(t *testing.T) {
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	result, err := gen.Generate(chunks, 2)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Verify connection points align vertically between adjacent chunks
	for i := 1; i < len(result.PlacedChunks); i++ {
		prev := result.PlacedChunks[i-1]
		curr := result.PlacedChunks[i]

		rightConns := prev.Chunk.GetConnections(procgen.EdgeRight)
		leftConns := curr.Chunk.GetConnections(procgen.EdgeLeft)

		if len(rightConns) == 0 || len(leftConns) == 0 {
			continue
		}

		prevConnY := prev.OffsetY + rightConns[0].YOffset
		currConnY := curr.OffsetY + leftConns[0].YOffset

		if abs(prevConnY-currConnY) > 0.01 {
			t.Errorf("connection Y mismatch between chunk %d and %d: %v vs %v",
				i-1, i, prevConnY, currConnY)
		}
	}
}

func TestGenerateDimensions(t *testing.T) {
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	result, err := gen.Generate(chunks, 3)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if result.TotalWidth <= 0 {
		t.Errorf("expected positive total width, got %d", result.TotalWidth)
	}
	if result.TotalHeight <= 0 {
		t.Errorf("expected positive total height, got %d", result.TotalHeight)
	}

	// Total width should be sum of all chunk widths
	var expectedWidth float64
	for _, pc := range result.PlacedChunks {
		expectedWidth += float64(pc.Chunk.Width)
	}
	if abs(float64(result.TotalWidth)-expectedWidth) > 1.0 {
		t.Errorf("total width %d doesn't match sum of chunk widths %v",
			result.TotalWidth, expectedWidth)
	}
}

func TestGenerateDifferentSeeds(t *testing.T) {
	chunks := loadTestChunks(t)

	// With only 3 middle chunks, sequences may repeat.
	// Just verify both generate without error and produce valid results.
	g1 := procgen.NewChunkGenerator(1)
	r1, err := g1.Generate(chunks, 5)
	if err != nil {
		t.Fatalf("Generate seed 1 failed: %v", err)
	}

	g2 := procgen.NewChunkGenerator(2)
	r2, err := g2.Generate(chunks, 5)
	if err != nil {
		t.Fatalf("Generate seed 2 failed: %v", err)
	}

	if len(r1.PlacedChunks) != len(r2.PlacedChunks) {
		t.Error("different seeds should produce same number of chunks for same numMiddle")
	}
}

func TestGenerateNoStartChunks(t *testing.T) {
	// Create chunks with no start tag
	chunks := []*procgen.Chunk{
		{ID: "mid", Tags: []procgen.ChunkTag{procgen.TagCombat}, Width: 320, Height: 320,
			Connections: []procgen.ConnectionPoint{
				{Edge: procgen.EdgeLeft, YOffset: 224, Width: 48},
				{Edge: procgen.EdgeRight, YOffset: 224, Width: 48},
			}},
	}

	gen := procgen.NewChunkGenerator(42)
	_, err := gen.Generate(chunks, 1)
	if err == nil {
		t.Error("expected error when no start chunks available")
	}
}

func TestGenerateNoExitChunks(t *testing.T) {
	chunks := []*procgen.Chunk{
		{ID: "start", Tags: []procgen.ChunkTag{procgen.TagStart}, Width: 320, Height: 320,
			Connections: []procgen.ConnectionPoint{
				{Edge: procgen.EdgeRight, YOffset: 224, Width: 48},
			}},
	}

	gen := procgen.NewChunkGenerator(42)
	_, err := gen.Generate(chunks, 1)
	if err == nil {
		t.Error("expected error when no exit chunks available")
	}
}

func TestGenerateYNormalization(t *testing.T) {
	chunks := loadTestChunks(t)

	gen := procgen.NewChunkGenerator(42)
	result, err := gen.Generate(chunks, 3)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// All Y offsets should be >= 0 after normalization
	for i, pc := range result.PlacedChunks {
		if pc.OffsetY < -0.01 {
			t.Errorf("chunk %d has negative Y offset: %v", i, pc.OffsetY)
		}
	}
}

func TestFindMatchingEdges(t *testing.T) {
	tests := []struct {
		name     string
		prev     *procgen.Chunk
		curr     *procgen.Chunk
		wantPrev procgen.ConnectionEdge
		wantCurr procgen.ConnectionEdge
		wantErr  bool
	}{
		{
			name: "horizontal right-left",
			prev: &procgen.Chunk{ID: "a", Connections: []procgen.ConnectionPoint{
				{Edge: procgen.EdgeRight, YOffset: 224, Width: 48},
			}},
			curr: &procgen.Chunk{ID: "b", Connections: []procgen.ConnectionPoint{
				{Edge: procgen.EdgeLeft, YOffset: 224, Width: 48},
			}},
			wantPrev: procgen.EdgeRight,
			wantCurr: procgen.EdgeLeft,
		},
		{
			name: "no compatible edges",
			prev: &procgen.Chunk{ID: "a", Connections: []procgen.ConnectionPoint{
				{Edge: procgen.EdgeLeft, YOffset: 224, Width: 48},
			}},
			curr: &procgen.Chunk{ID: "b", Connections: []procgen.ConnectionPoint{
				{Edge: procgen.EdgeRight, YOffset: 224, Width: 48},
			}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevEdge, currEdge, err := procgen.FindMatchingEdges(tt.prev, tt.curr)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if prevEdge != tt.wantPrev {
				t.Errorf("prev edge: got %s, want %s", prevEdge, tt.wantPrev)
			}
			if currEdge != tt.wantCurr {
				t.Errorf("curr edge: got %s, want %s", currEdge, tt.wantCurr)
			}
		})
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
