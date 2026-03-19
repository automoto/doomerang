package procgen_test

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/procgen"
)

func TestEnemyPlacementCombatChunk(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/combat_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	pc := procgen.PlacedChunk{Chunk: chunk, OffsetX: 0, OffsetY: 0}
	rng := rand.New(rand.NewSource(42))
	placer := procgen.NewEnemyPlacer(rng)

	spawns, _ := placer.PlaceEnemies(pc, 3)
	if len(spawns) == 0 {
		t.Error("expected enemies to be placed in combat chunk")
	}
}

func TestEnemyPlacementBreakChunk(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/break_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	pc := procgen.PlacedChunk{Chunk: chunk, OffsetX: 0, OffsetY: 0}
	rng := rand.New(rand.NewSource(42))
	placer := procgen.NewEnemyPlacer(rng)

	spawns, _ := placer.PlaceEnemies(pc, 3)
	if len(spawns) != 0 {
		t.Errorf("expected no enemies in break chunk, got %d", len(spawns))
	}
}

func TestEnemySpacing(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/combat_02.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	pc := procgen.PlacedChunk{Chunk: chunk, OffsetX: 0, OffsetY: 0}
	rng := rand.New(rand.NewSource(42))
	placer := procgen.NewEnemyPlacer(rng)

	spawns, _ := placer.PlaceEnemies(pc, 5)
	minSpacing := config.Procgen.EnemyMinSpacing

	for i := 0; i < len(spawns); i++ {
		for j := i + 1; j < len(spawns); j++ {
			dx := spawns[i].X - spawns[j].X
			if dx < 0 {
				dx = -dx
			}
			// Only check spacing for enemies on the same Y level
			dy := spawns[i].Y - spawns[j].Y
			if dy < 0 {
				dy = -dy
			}
			if dy < 16 && dx < minSpacing {
				t.Errorf("enemies too close: (%.0f,%.0f) and (%.0f,%.0f), distance=%.0f, min=%.0f",
					spawns[i].X, spawns[i].Y, spawns[j].X, spawns[j].Y, dx, minSpacing)
			}
		}
	}
}

func TestEnemyTypesValidate(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/combat_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	pc := procgen.PlacedChunk{Chunk: chunk, OffsetX: 0, OffsetY: 0}
	rng := rand.New(rand.NewSource(42))
	placer := procgen.NewEnemyPlacer(rng)

	spawns, _ := placer.PlaceEnemies(pc, 3)
	validTypes := map[string]bool{
		"Guard": true, "LightGuard": true,
		"HeavyGuard": true, "KnifeThrower": true,
	}

	for _, s := range spawns {
		if !validTypes[s.EnemyType] {
			t.Errorf("invalid enemy type: %s", s.EnemyType)
		}
	}
}

func TestEnemyWorldSpaceCoords(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/combat_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	offsetX := 500.0
	offsetY := 100.0
	pc := procgen.PlacedChunk{Chunk: chunk, OffsetX: offsetX, OffsetY: offsetY}
	rng := rand.New(rand.NewSource(42))
	placer := procgen.NewEnemyPlacer(rng)

	spawns, _ := placer.PlaceEnemies(pc, 3)
	for _, s := range spawns {
		if s.X < offsetX || s.X > offsetX+float64(chunk.Width) {
			t.Errorf("enemy X=%.0f outside chunk world bounds [%.0f, %.0f]",
				s.X, offsetX, offsetX+float64(chunk.Width))
		}
	}
}

func TestEnemyPatrolPaths(t *testing.T) {
	loader := procgen.NewChunkLoader()
	chunk, err := loader.LoadChunk("chunks/combat_01.tmx")
	if err != nil {
		t.Fatalf("LoadChunk failed: %v", err)
	}

	offsetX := 200.0
	offsetY := 50.0
	pc := procgen.PlacedChunk{Chunk: chunk, OffsetX: offsetX, OffsetY: offsetY}
	rng := rand.New(rand.NewSource(42))
	placer := procgen.NewEnemyPlacer(rng)

	spawns, paths := placer.PlaceEnemies(pc, 3)
	if len(spawns) == 0 {
		t.Fatal("expected enemies to be placed")
	}

	// Every spawn with a patrol path must have a matching entry in paths
	for _, s := range spawns {
		if s.PatrolPath == "" {
			continue
		}
		path, ok := paths[s.PatrolPath]
		if !ok {
			t.Errorf("spawn references patrol path %q but it was not returned", s.PatrolPath)
			continue
		}

		// Path name should contain chunk ID
		if !strings.Contains(s.PatrolPath, chunk.ID) {
			t.Errorf("patrol path name %q does not contain chunk ID %q", s.PatrolPath, chunk.ID)
		}

		// Path must have exactly 2 points
		if len(path.Points) != 2 {
			t.Errorf("patrol path %q has %d points, want 2", s.PatrolPath, len(path.Points))
			continue
		}

		// Left point must be less than right point
		if path.Points[0].X >= path.Points[1].X {
			t.Errorf("patrol path %q: left X (%.0f) >= right X (%.0f)",
				s.PatrolPath, path.Points[0].X, path.Points[1].X)
		}

		// Points must include offset
		if path.Points[0].X < offsetX {
			t.Errorf("patrol path %q: left X (%.0f) < offsetX (%.0f)",
				s.PatrolPath, path.Points[0].X, offsetX)
		}
	}

	// At least one enemy should have a patrol path (combat chunks have wide platforms)
	hasPath := false
	for _, s := range spawns {
		if s.PatrolPath != "" {
			hasPath = true
			break
		}
	}
	if !hasPath {
		t.Error("expected at least one enemy to have a patrol path on a combat chunk")
	}
}
