package procgen_test

import (
	"math/rand"
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

	spawns := placer.PlaceEnemies(pc, 3)
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

	spawns := placer.PlaceEnemies(pc, 3)
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

	spawns := placer.PlaceEnemies(pc, 5)
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

	spawns := placer.PlaceEnemies(pc, 3)
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

	spawns := placer.PlaceEnemies(pc, 3)
	for _, s := range spawns {
		if s.X < offsetX || s.X > offsetX+float64(chunk.Width) {
			t.Errorf("enemy X=%.0f outside chunk world bounds [%.0f, %.0f]",
				s.X, offsetX, offsetX+float64(chunk.Width))
		}
	}
}
