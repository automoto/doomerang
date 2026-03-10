package systems_test

import (
	"testing"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func newTestECS() *ecs.ECS {
	return ecs.NewECS(donburi.NewWorld())
}

func addRunStats(e *ecs.ECS, boundaries []float64) *components.RunStatsData {
	entry := e.World.Entry(e.Create(cfg.Default, components.RunStats))
	components.RunStats.SetValue(entry, components.RunStatsData{
		RoomBoundaries: boundaries,
	})
	return components.RunStats.Get(entry)
}

func addPlayer(e *ecs.ECS, x, y float64) *donburi.Entry {
	entry := e.World.Entry(e.Create(cfg.Default, tags.Player, components.Object))
	obj := resolv.NewObject(x, y, 16, 40)
	components.Object.SetValue(entry, components.ObjectData{Object: obj})
	return entry
}

func TestElapsedTicksIncrement(t *testing.T) {
	e := newTestECS()
	stats := addRunStats(e, nil)

	for i := 0; i < 10; i++ {
		systems.UpdateRunStats(e)
	}

	if stats.ElapsedTicks != 10 {
		t.Errorf("expected ElapsedTicks=10, got %d", stats.ElapsedTicks)
	}
}

func TestKillCountDelta(t *testing.T) {
	e := newTestECS()
	stats := addRunStats(e, nil)

	// Add 5 enemies
	enemies := make([]*donburi.Entry, 5)
	for i := range enemies {
		enemies[i] = e.World.Entry(e.Create(cfg.Default, tags.Enemy))
	}

	// First tick establishes baseline
	systems.UpdateRunStats(e)
	if stats.KillCount != 0 {
		t.Errorf("expected KillCount=0 after first tick, got %d", stats.KillCount)
	}

	// Remove 2 enemies
	e.World.Remove(enemies[0].Entity())
	e.World.Remove(enemies[1].Entity())

	// Second tick detects delta
	systems.UpdateRunStats(e)
	if stats.KillCount != 2 {
		t.Errorf("expected KillCount=2, got %d", stats.KillCount)
	}
}

func TestKillCountNoFalsePositives(t *testing.T) {
	e := newTestECS()
	stats := addRunStats(e, nil)

	for i := 0; i < 3; i++ {
		e.World.Entry(e.Create(cfg.Default, tags.Enemy))
	}

	systems.UpdateRunStats(e)
	systems.UpdateRunStats(e)

	if stats.KillCount != 0 {
		t.Errorf("expected KillCount=0 without removals, got %d", stats.KillCount)
	}
}

func TestRoomsClearedAdvances(t *testing.T) {
	e := newTestECS()
	stats := addRunStats(e, []float64{320, 640})
	playerEntry := addPlayer(e, 0, 0)

	// Player at x=350 (past boundary 320)
	components.Object.Get(playerEntry).X = 350
	systems.UpdateRunStats(e)
	if stats.RoomsCleared != 1 {
		t.Errorf("expected RoomsCleared=1 at x=350, got %d", stats.RoomsCleared)
	}

	// Player at x=700 (past boundary 640)
	components.Object.Get(playerEntry).X = 700
	systems.UpdateRunStats(e)
	if stats.RoomsCleared != 2 {
		t.Errorf("expected RoomsCleared=2 at x=700, got %d", stats.RoomsCleared)
	}
}

func TestRoomsNeverGoBackward(t *testing.T) {
	e := newTestECS()
	stats := addRunStats(e, []float64{320, 640})
	playerEntry := addPlayer(e, 0, 0)

	// Advance to room 2
	components.Object.Get(playerEntry).X = 700
	systems.UpdateRunStats(e)
	if stats.RoomsCleared != 2 {
		t.Fatalf("expected RoomsCleared=2, got %d", stats.RoomsCleared)
	}

	// Move player backward
	components.Object.Get(playerEntry).X = 100
	systems.UpdateRunStats(e)
	if stats.RoomsCleared != 2 {
		t.Errorf("RoomsCleared should not decrease: expected 2, got %d", stats.RoomsCleared)
	}
}
