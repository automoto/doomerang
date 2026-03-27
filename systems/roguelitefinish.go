package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

// NewUpdateRogueliteFinish returns a system that watches for level complete and transitions
// to the run summary scene when the player confirms.
func NewUpdateRogueliteFinish(sceneChanger SceneChanger, createSummaryScene func() interface{}) ecs.System {
	return func(e *ecs.ECS) {
		levelComplete := GetOrCreateLevelComplete(e)
		if !levelComplete.IsComplete {
			return
		}

		input := getOrCreateInput(e)
		if GetAction(input, cfg.ActionMenuSelect).JustPressed {
			PlaySFX(e, cfg.SoundMenuSelect)
			sceneChanger.ChangeScene(createSummaryScene())
		}
	}
}

// DrawRogueliteFinish renders the finish overlay for roguelite mode.
func DrawRogueliteFinish(e *ecs.ECS, screen *ebiten.Image) {
	levelComplete := GetOrCreateLevelComplete(e)
	if !levelComplete.IsComplete {
		return
	}

	width := float64(screen.Bounds().Dx())

	// Semi-transparent overlay
	vector.FillRect(
		screen,
		0, 0,
		float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()),
		cfg.LevelComplete.OverlayColor,
		false,
	)

	// Title "RUN COMPLETE"
	titleFont := fonts.ExcelTitle.GetV2()
	title := cfg.RunSummary.Title
	titleX := centerTextX(title, titleFont, width)
	drawText(screen, title, titleFont, titleX, int(cfg.RunSummary.TitleY), cfg.RunSummary.TitleColor)

	// Input hint
	hintFont := fonts.ExcelSmall.GetV2()
	input := getOrCreateInput(e)
	hint := getRunFinishHint(input.LastInputMethod)
	hintX := centerTextX(hint, hintFont, width)
	drawText(screen, hint, hintFont, hintX, int(cfg.LevelComplete.HintY), cfg.LevelComplete.HintColor)
}

func getRunFinishHint(method components.InputMethod) string {
	switch method {
	case components.InputPlayStation:
		return "Press Cross to view results"
	case components.InputXbox:
		return "Press A to view results"
	}
	return "Press ENTER to view results"
}
