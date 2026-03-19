package procgen

import (
	"testing"

	"github.com/automoto/doomerang/assets"
	tiled "github.com/lafriks/go-tiled"
)

func TestCompileDeadZonesFromChunks(t *testing.T) {
	compiler := NewCompiler()
	level := &assets.Level{
		DeadZones: []assets.DeadZone{},
	}

	chunk := &Chunk{
		ID:     "test_dz",
		Width:  640,
		Height: 320,
		TiledMap: &tiled.Map{
			ObjectGroups: []*tiled.ObjectGroup{
				{
					Name: "DeadZones",
					Objects: []*tiled.Object{
						{X: 100, Y: 280, Width: 200, Height: 32},
						{X: 400, Y: 300, Width: 100, Height: 16},
					},
				},
			},
		},
	}

	pc := PlacedChunk{
		Chunk:   chunk,
		OffsetX: 320,
		OffsetY: 64,
	}

	compiler.compileObjectGroups(level, pc)

	if len(level.DeadZones) != 2 {
		t.Fatalf("expected 2 dead zones, got %d", len(level.DeadZones))
	}

	// First dead zone: chunk-local (100, 280) + offset (320, 64)
	dz0 := level.DeadZones[0]
	if dz0.X != 420 || dz0.Y != 344 {
		t.Errorf("dead zone 0: expected position (420, 344), got (%v, %v)", dz0.X, dz0.Y)
	}
	if dz0.Width != 200 || dz0.Height != 32 {
		t.Errorf("dead zone 0: expected size (200, 32), got (%v, %v)", dz0.Width, dz0.Height)
	}

	// Second dead zone: chunk-local (400, 300) + offset (320, 64)
	dz1 := level.DeadZones[1]
	if dz1.X != 720 || dz1.Y != 364 {
		t.Errorf("dead zone 1: expected position (720, 364), got (%v, %v)", dz1.X, dz1.Y)
	}
	if dz1.Width != 100 || dz1.Height != 16 {
		t.Errorf("dead zone 1: expected size (100, 16), got (%v, %v)", dz1.Width, dz1.Height)
	}
}

func TestCompileDeadZonesCoexistWithHazardSlots(t *testing.T) {
	compiler := NewCompiler()
	level := &assets.Level{
		DeadZones: []assets.DeadZone{},
	}

	chunk := &Chunk{
		ID:     "test_dz_hazard",
		Width:  640,
		Height: 320,
		TiledMap: &tiled.Map{
			ObjectGroups: []*tiled.ObjectGroup{
				{
					Name: "DeadZones",
					Objects: []*tiled.Object{
						{X: 50, Y: 280, Width: 100, Height: 32},
					},
				},
				{
					Name: "HazardSlots",
					Objects: []*tiled.Object{
						{
							X: 300, Y: 290, Width: 80, Height: 24,
							Properties: tiled.Properties{
								&tiled.Property{Name: "hazard_type", Value: "deadzone"},
							},
						},
					},
				},
			},
		},
	}

	pc := PlacedChunk{
		Chunk:   chunk,
		OffsetX: 0,
		OffsetY: 0,
	}

	compiler.compileObjectGroups(level, pc)

	if len(level.DeadZones) != 2 {
		t.Fatalf("expected 2 dead zones (1 DeadZones + 1 HazardSlot), got %d", len(level.DeadZones))
	}

	// DeadZones layer entry
	if level.DeadZones[0].X != 50 || level.DeadZones[0].Width != 100 {
		t.Errorf("dead zone 0: expected X=50 Width=100, got X=%v Width=%v",
			level.DeadZones[0].X, level.DeadZones[0].Width)
	}

	// HazardSlots deadzone entry
	if level.DeadZones[1].X != 300 || level.DeadZones[1].Width != 80 {
		t.Errorf("dead zone 1: expected X=300 Width=80, got X=%v Width=%v",
			level.DeadZones[1].X, level.DeadZones[1].Width)
	}
}
