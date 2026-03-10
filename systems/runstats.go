package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const ticksPerSecond = 60

// UpdateRunStats updates per-run statistics each frame.
// Must run after UpdateDeaths so kill deltas reflect removed enemies.
func UpdateRunStats(e *ecs.ECS) {
	stats := GetOrCreateRunStats(e)

	stats.ElapsedTicks++

	// Count live enemies for kill delta
	liveCount := 0
	tags.Enemy.Each(e.World, func(_ *donburi.Entry) {
		liveCount++
	})
	delta := stats.PrevEnemyCount - liveCount
	if delta > 0 {
		stats.KillCount += delta
	}
	stats.PrevEnemyCount = liveCount

	// Advance rooms cleared based on player X position
	playerEntry, ok := tags.Player.First(e.World)
	if !ok || !playerEntry.HasComponent(components.Object) {
		return
	}
	playerX := components.Object.Get(playerEntry).X
	for i := stats.LastRoomIndex; i < len(stats.RoomBoundaries); i++ {
		if playerX < stats.RoomBoundaries[i] {
			break
		}
		stats.RoomsCleared = i + 1
		stats.LastRoomIndex = i + 1
	}
}

// GetOrCreateRunStats returns the singleton RunStats component, creating if needed.
func GetOrCreateRunStats(e *ecs.ECS) *components.RunStatsData {
	if _, ok := components.RunStats.First(e.World); !ok {
		ent := e.World.Entry(e.World.Create(components.RunStats))
		components.RunStats.SetValue(ent, components.RunStatsData{})
	}
	ent, _ := components.RunStats.First(e.World)
	return components.RunStats.Get(ent)
}

// FinalRunStats is the snapshot passed to the run summary scene.
type FinalRunStats struct {
	Seed         int64
	TotalRooms   int
	RoomsCleared int
	KillCount    int
	ElapsedSecs  int64
}

// SnapshotRunStats converts the live RunStatsData into a FinalRunStats.
func SnapshotRunStats(e *ecs.ECS) FinalRunStats {
	stats := GetOrCreateRunStats(e)
	elapsedSecs := stats.ElapsedTicks / ticksPerSecond
	return FinalRunStats{
		Seed:         stats.Seed,
		TotalRooms:   stats.TotalRooms,
		RoomsCleared: stats.RoomsCleared,
		KillCount:    stats.KillCount,
		ElapsedSecs:  elapsedSecs,
	}
}
