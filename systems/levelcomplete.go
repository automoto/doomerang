package systems

import (
	"os"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

// UpdateLevelComplete handles input when level complete overlay is shown
func UpdateLevelComplete(e *ecs.ECS) {
	levelComplete := GetOrCreateLevelComplete(e)
	if !levelComplete.IsComplete {
		return
	}

	input := getOrCreateInput(e)

	// Handle selection - Enter or Start exits the game
	if GetAction(input, cfg.ActionMenuSelect).JustPressed {
		PlaySFX(e, cfg.SoundMenuSelect)
		os.Exit(0)
	}
}

// DrawLevelComplete renders the level complete overlay
func DrawLevelComplete(e *ecs.ECS, screen *ebiten.Image) {
	levelComplete := GetOrCreateLevelComplete(e)
	if !levelComplete.IsComplete {
		return
	}

	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())

	// Draw semi-transparent overlay
	vector.FillRect(
		screen,
		0, 0,
		float32(width), float32(height),
		cfg.LevelComplete.OverlayColor,
		false,
	)

	// Draw title
	titleFont := fonts.ExcelTitle.GetV2()
	title := cfg.LevelComplete.Title
	titleX := centerTextX(title, titleFont, width)
	drawText(screen, title, titleFont, titleX, int(cfg.LevelComplete.TitleY), cfg.LevelComplete.TitleColor)

	// Draw message
	msgFont := fonts.ExcelBold.GetV2()
	msg := cfg.LevelComplete.Message
	msgX := centerTextX(msg, msgFont, width)
	drawText(screen, msg, msgFont, msgX, int(cfg.LevelComplete.MessageY), cfg.LevelComplete.TextColor)

	// Draw continue hint
	hintFont := fonts.ExcelSmall.GetV2()
	input := getOrCreateInput(e)
	hint := getLevelCompleteHint(input.LastInputMethod)
	hintX := centerTextX(hint, hintFont, width)
	drawText(screen, hint, hintFont, hintX, int(cfg.LevelComplete.HintY), cfg.LevelComplete.HintColor)
}

// getLevelCompleteHint returns the appropriate hint for level complete screen
func getLevelCompleteHint(method components.InputMethod) string {
	switch method {
	case components.InputPlayStation:
		return "Press Cross to exit"
	case components.InputXbox:
		return "Press A to exit"
	}
	return cfg.LevelComplete.ContinueHint
}

// GetOrCreateLevelComplete returns the singleton LevelComplete component, creating if needed
func GetOrCreateLevelComplete(e *ecs.ECS) *components.LevelCompleteData {
	if _, ok := components.LevelComplete.First(e.World); !ok {
		ent := e.World.Entry(e.World.Create(components.LevelComplete))
		components.LevelComplete.SetValue(ent, components.LevelCompleteData{
			IsComplete: false,
		})
	}

	ent, _ := components.LevelComplete.First(e.World)
	return components.LevelComplete.Get(ent)
}

// IsLevelComplete checks if the level is complete
func IsLevelComplete(e *ecs.ECS) bool {
	levelComplete := GetOrCreateLevelComplete(e)
	return levelComplete.IsComplete
}

// WithLevelCompleteCheck wraps a system to skip execution when level is complete
func WithLevelCompleteCheck(system ecs.System) ecs.System {
	return func(e *ecs.ECS) {
		if IsLevelComplete(e) {
			return
		}
		system(e)
	}
}

// WithGameplayChecks wraps a system to skip execution when paused or level is complete
func WithGameplayChecks(system ecs.System) ecs.System {
	return WithPauseCheck(WithLevelCompleteCheck(system))
}
