package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func CreateLevel(ecs *ecs.ECS) *donburi.Entry {
	level := archetypes.Level.Spawn(ecs)

	// Load all levels
	loader := assets.NewLevelLoader()
	levels := loader.MustLoadLevels()

	if len(levels) == 0 {
		panic("No levels found in assets/levels directory")
	}

	// Set up level data
	levelData := &components.LevelData{
		Levels:       levels,
		LevelIndex:   0,
		CurrentLevel: &levels[0],
	}

	components.Level.Set(level, levelData)

	return level
}
