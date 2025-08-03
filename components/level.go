package components

import (
	"github.com/automoto/doomerang/assets"
	"github.com/yohamta/donburi"
)

type LevelData struct {
	CurrentLevel *assets.Level
	LevelIndex   int
	Levels       []assets.Level
}

var Level = donburi.NewComponentType[LevelData]()
