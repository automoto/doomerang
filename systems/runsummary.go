package systems

import (
	"fmt"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

// RunSummaryOption represents the menu choices on the run summary screen.
type RunSummaryOption int

const (
	RunSummaryPlayAgain RunSummaryOption = iota
	RunSummaryMainMenu
)

// runSummaryMenuOptions are the display strings for each option.
var runSummaryMenuOptions = []string{"Play Again", "Main Menu"}

// GetOrCreateRunSummaryMenu returns the menu selection state for the run summary screen.
// Intentionally reuses GameOverData — both screens need a single int selection field
// and sharing avoids a redundant component type for identical data.
func GetOrCreateRunSummaryMenu(e *ecs.ECS) *components.GameOverData {
	return GetOrCreateGameOver(e)
}

// NewUpdateRunSummary creates the update system for the run summary screen.
// It follows the same closure pattern as NewUpdateGameOver.
func NewUpdateRunSummary(
	sceneChanger SceneChanger,
	stats FinalRunStats,
	createRogueliteScene func() interface{},
	createMenuScene func() interface{},
) ecs.System {
	// firstFrame guard prevents input bleed: if the player is still holding the
	// confirm key when this scene loads (carried over from the finish overlay),
	// the fresh InputData sees Current=true, Previous=false → JustPressed=true,
	// which would immediately trigger a transition. Skipping the first frame lets
	// UpdateInput record the held state into Previous before we act on it.
	firstFrame := true
	return func(e *ecs.ECS) {
		if firstFrame {
			firstFrame = false
			return
		}

		menu := GetOrCreateRunSummaryMenu(e)
		input := getOrCreateInput(e)

		numOptions := len(runSummaryMenuOptions)
		if GetAction(input, cfg.ActionMenuUp).JustPressed {
			menu.SelectedOption = components.GameOverOption(
				(int(menu.SelectedOption) - 1 + numOptions) % numOptions,
			)
		}
		if GetAction(input, cfg.ActionMenuDown).JustPressed {
			menu.SelectedOption = components.GameOverOption(
				(int(menu.SelectedOption) + 1) % numOptions,
			)
		}

		if GetAction(input, cfg.ActionMenuSelect).JustPressed {
			switch RunSummaryOption(menu.SelectedOption) {
			case RunSummaryPlayAgain:
				sceneChanger.ChangeScene(createRogueliteScene())
			case RunSummaryMainMenu:
				sceneChanger.ChangeScene(createMenuScene())
			}
		}
	}
}

// DrawRunSummary renders the full-screen run summary.
// Formatted stat strings are pre-computed once at construction time since stats never change.
func DrawRunSummary(stats FinalRunStats) func(e *ecs.ECS, screen *ebiten.Image) {
	type statRow struct{ label, value string }
	rows := []statRow{
		{"Rooms Cleared", fmt.Sprintf("%d / %d", stats.RoomsCleared, stats.TotalRooms)},
		{"Enemies Killed", fmt.Sprintf("%d", stats.KillCount)},
		{"Time", fmt.Sprintf("%dm %02ds", stats.ElapsedSecs/60, stats.ElapsedSecs%60)},
		{"Seed", fmt.Sprintf("#%05d", absInt64(stats.Seed)%100000)},
	}

	return func(e *ecs.ECS, screen *ebiten.Image) {
		menu := GetOrCreateRunSummaryMenu(e)
		width := float64(screen.Bounds().Dx())

		// Full-screen dark background
		vector.DrawFilledRect(
			screen,
			0, 0,
			float32(screen.Bounds().Dx()), float32(screen.Bounds().Dy()),
			cfg.RunSummary.BackgroundColor,
			false,
		)

		// Title
		titleFont := fonts.ExcelTitle.Get()
		titleX := centerTextX(cfg.RunSummary.Title, titleFont, width)
		text.Draw(screen, cfg.RunSummary.Title, titleFont, titleX, int(cfg.RunSummary.TitleY), cfg.RunSummary.TitleColor)

		// Stat rows — centered as a unit with a fixed pixel gap between label and value.
		// BoundString trailing-space width is unreliable for pixel fonts, so we measure
		// "label:" and value separately, sum them with an explicit gap, then center.
		const labelValueGap = 10
		statsFont := fonts.ExcelBold.Get()
		for i, row := range rows {
			y := int(cfg.RunSummary.StatsStartY) + i*int(cfg.RunSummary.StatsRowHeight)

			labelPart := row.label + ":"
			labelW := text.BoundString(statsFont, labelPart).Dx()
			valueW := text.BoundString(statsFont, row.value).Dx()
			totalW := float64(labelW + labelValueGap + valueW)
			startX := int((width - totalW) / 2)

			text.Draw(screen, labelPart, statsFont, startX, y, cfg.RunSummary.LabelColor)
			text.Draw(screen, row.value, statsFont, startX+labelW+labelValueGap, y, cfg.RunSummary.ValueColor)
		}

		// Menu options (reuse statsFont — same ExcelBold face)
		for i, option := range runSummaryMenuOptions {
			y := int(cfg.RunSummary.MenuStartY) + i*int(cfg.RunSummary.MenuItemHeight)
			color := cfg.RunSummary.TextColorNormal
			if RunSummaryOption(menu.SelectedOption) == RunSummaryOption(i) {
				color = cfg.RunSummary.TextColorSelected
			}
			x := centerTextX(option, statsFont, width)
			text.Draw(screen, option, statsFont, x, y+int(cfg.RunSummary.MenuItemHeight), color)
		}
	}
}

func absInt64(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
