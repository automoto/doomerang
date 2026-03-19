package procgen_test

import (
	"testing"

	"github.com/automoto/doomerang/procgen"
)

func loadTestChunks(t *testing.T) []*procgen.Chunk {
	t.Helper()
	loader := procgen.NewChunkLoader()
	chunks, err := loader.LoadAllChunks("chunks")
	if err != nil {
		t.Fatalf("LoadAllChunks failed: %v", err)
	}
	return chunks
}

func TestLoadAllChunks(t *testing.T) {
	chunks := loadTestChunks(t)
	if len(chunks) < 15 {
		t.Errorf("expected at least 15 chunks, got %d", len(chunks))
	}
}

func TestLoadSingleChunk(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/start_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	if chunk.ID != "start_01" {
		t.Errorf("expected chunk_id 'start_01', got '%s'", chunk.ID)
	}
	if chunk.Biome != "cyberpunk" {
		t.Errorf("expected biome 'cyberpunk', got '%s'", chunk.Biome)
	}
	if chunk.Difficulty != 1 {
		t.Errorf("expected difficulty 1, got %d", chunk.Difficulty)
	}
	if !chunk.HasTag(procgen.TagStart) {
		t.Error("expected chunk to have 'start' tag")
	}
}

func TestChunkDimensions(t *testing.T) {
	loader := procgen.NewChunkLoader()

	tests := []struct {
		path       string
		tileWidth  int
		tileHeight int
		pixelW     int
		pixelH     int
	}{
		{"chunks/start_01.tmx", 20, 20, 320, 320},
		{"chunks/combat_01.tmx", 40, 20, 640, 320},
		{"chunks/traversal_01.tmx", 40, 20, 640, 320},
		{"chunks/break_01.tmx", 20, 20, 320, 320},
		{"chunks/exit_01.tmx", 20, 20, 320, 320},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			chunk, err := loader.LoadChunk(tt.path)
			if err != nil {
				t.Fatalf("LoadChunk(%s) failed: %v", tt.path, err)
			}
			if chunk.TileWidth != tt.tileWidth || chunk.TileHeight != tt.tileHeight {
				t.Errorf("tile dimensions: got %dx%d, want %dx%d",
					chunk.TileWidth, chunk.TileHeight, tt.tileWidth, tt.tileHeight)
			}
			if chunk.Width != tt.pixelW || chunk.Height != tt.pixelH {
				t.Errorf("pixel dimensions: got %dx%d, want %dx%d",
					chunk.Width, chunk.Height, tt.pixelW, tt.pixelH)
			}
		})
	}
}

func TestChunkSolidTiles(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/start_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	if len(chunk.SolidTiles) == 0 {
		t.Error("expected solid tiles to be parsed from wg-tiles layer")
	}

	// start_01 has left wall (col 0, GID 18) + floor/fill (rows 17-19, GID 16/28)
	// Col 0: 20 wall tiles, Row 17: 19 floor tiles, Rows 18-19: 19 fill tiles each
	// Each tile is 16x16
	for _, tile := range chunk.SolidTiles {
		if tile.Width != 16 || tile.Height != 16 {
			t.Errorf("expected 16x16 tiles, got %vx%v", tile.Width, tile.Height)
		}
	}
}

func TestChunkConnectionPoints(t *testing.T) {
	loader := procgen.NewChunkLoader()

	// start_01 should have 1 right connection
	chunk, err := loader.LoadChunk("chunks/start_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	rightConns := chunk.GetConnections(procgen.EdgeRight)
	if len(rightConns) != 1 {
		t.Errorf("start_01: expected 1 right connection, got %d", len(rightConns))
	}

	leftConns := chunk.GetConnections(procgen.EdgeLeft)
	if len(leftConns) != 0 {
		t.Errorf("start_01: expected 0 left connections, got %d", len(leftConns))
	}

	// combat_01 should have left and right connections
	combat, err := loader.LoadChunk("chunks/combat_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	if len(combat.GetConnections(procgen.EdgeLeft)) != 1 {
		t.Error("combat_01: expected 1 left connection")
	}
	if len(combat.GetConnections(procgen.EdgeRight)) != 1 {
		t.Error("combat_01: expected 1 right connection")
	}

	// exit_01 should have only left connection
	exit, err := loader.LoadChunk("chunks/exit_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	if len(exit.GetConnections(procgen.EdgeLeft)) != 1 {
		t.Error("exit_01: expected 1 left connection")
	}
	if len(exit.GetConnections(procgen.EdgeRight)) != 0 {
		t.Errorf("exit_01: expected 0 right connections, got %d", len(exit.GetConnections(procgen.EdgeRight)))
	}
}

func TestChunkConnectionWidth(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/start_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	conns := chunk.GetConnections(procgen.EdgeRight)
	if len(conns) == 0 {
		t.Fatal("expected at least 1 right connection")
	}
	if conns[0].Width != 48 { // 3 tiles * 16px
		t.Errorf("expected connection width 48, got %v", conns[0].Width)
	}
}

func TestChunkEnemySlots(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/combat_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	if len(chunk.EnemySlots) != 2 {
		t.Errorf("combat_01: expected 2 enemy slots, got %d", len(chunk.EnemySlots))
	}

	if chunk.MinEnemies != 2 || chunk.MaxEnemies != 4 {
		t.Errorf("combat_01: expected min/max enemies 2/4, got %d/%d",
			chunk.MinEnemies, chunk.MaxEnemies)
	}
}

func TestChunkHazardSlots(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/traversal_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	// traversal_01 uses DeadZones object layer instead of HazardSlots
	if len(chunk.HazardSlots) != 0 {
		t.Errorf("traversal_01: expected 0 hazard slots, got %d", len(chunk.HazardSlots))
	}

	// Verify the DeadZones object group exists in the tiled map
	var deadZoneCount int
	for _, og := range chunk.TiledMap.ObjectGroups {
		if og.Name == "DeadZones" {
			deadZoneCount += len(og.Objects)
		}
	}
	if deadZoneCount == 0 {
		t.Error("traversal_01: expected DeadZones object group with objects")
	}
}

func TestChunkTags(t *testing.T) {
	loader := procgen.NewChunkLoader()

	tests := []struct {
		path string
		tag  procgen.ChunkTag
	}{
		{"chunks/start_01.tmx", procgen.TagStart},
		{"chunks/combat_01.tmx", procgen.TagCombat},
		{"chunks/traversal_01.tmx", procgen.TagTraversal},
		{"chunks/break_01.tmx", procgen.TagBreak},
		{"chunks/exit_01.tmx", procgen.TagExit},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			chunk, err := loader.LoadChunk(tt.path)
			if err != nil {
				t.Fatalf("LoadChunk(%s) failed: %v", tt.path, err)
			}
			if !chunk.HasTag(tt.tag) {
				t.Errorf("chunk %s: expected tag '%s'", tt.path, tt.tag)
			}
		})
	}
}

func TestInvalidChunkPath(t *testing.T) {
	loader := procgen.NewChunkLoader()
	_, err := loader.LoadChunk("chunks/nonexistent.tmx")
	if err == nil {
		t.Error("expected error for nonexistent chunk file")
	}
}

func TestChunkMetadata(t *testing.T) {
	loader := procgen.NewChunkLoader()

	// Verify all chunks have unique IDs
	chunks, err := loader.LoadAllChunks("chunks")
	if err != nil {
		t.Fatalf("LoadAllChunks failed: %v", err)
	}

	ids := make(map[string]bool)
	for _, chunk := range chunks {
		if ids[chunk.ID] {
			t.Errorf("duplicate chunk_id: %s", chunk.ID)
		}
		ids[chunk.ID] = true

		// Every chunk should have a biome
		if chunk.Biome == "" {
			t.Errorf("chunk %s: missing biome", chunk.ID)
		}

		// Difficulty should be 1-5
		if chunk.Difficulty < 1 || chunk.Difficulty > 5 {
			t.Errorf("chunk %s: difficulty %d out of range [1,5]", chunk.ID, chunk.Difficulty)
		}

		// Every chunk should have at least one tag
		if len(chunk.Tags) == 0 {
			t.Errorf("chunk %s: no tags", chunk.ID)
		}
	}
}
