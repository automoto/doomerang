package components

import "github.com/yohamta/donburi"

// RunStatsData tracks per-run statistics during gameplay
type RunStatsData struct {
	Seed           int64
	TotalRooms     int
	RoomsCleared   int
	KillCount      int
	ElapsedTicks   int64
	PrevEnemyCount int       // internal: for delta-based kill detection
	RoomBoundaries []float64 // right-edge X of each placed chunk, in order
	LastRoomIndex  int       // highest room boundary crossed so far
}

var RunStats = donburi.NewComponentType[RunStatsData]()
