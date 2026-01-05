package systems

import (
	"log"
	"os"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

// UpdatePause handles pause toggle and menu navigation.
// This system should run AFTER UpdateInput but BEFORE other game systems.
func UpdatePause(ecs *ecs.ECS) {
	pause := GetOrCreatePause(ecs)
	input := getOrCreateInput(ecs)

	pauseAction := input.Actions[cfg.ActionPause]

	// Toggle pause on ESC or P
	if pauseAction.JustPressed {
		pause.IsPaused = !pause.IsPaused
		if pause.IsPaused {
			pause.SelectedOption = components.MenuResume
			PauseMusic(ecs)
		} else {
			ResumeMusic(ecs)
		}
	}

	// Only process menu input while paused
	if !pause.IsPaused {
		return
	}

	upAction := input.Actions[cfg.ActionMenuUp]
	downAction := input.Actions[cfg.ActionMenuDown]
	selectAction := input.Actions[cfg.ActionMenuSelect]

	// Navigate menu with wrap-around using modulo arithmetic
	numOptions := int(components.MenuExit) + 1
	if upAction.JustPressed {
		pause.SelectedOption = components.PauseMenuOption(
			(int(pause.SelectedOption) - 1 + numOptions) % numOptions,
		)
		PlaySFX(ecs, cfg.SoundMenuNavigate)
	}
	if downAction.JustPressed {
		pause.SelectedOption = components.PauseMenuOption(
			(int(pause.SelectedOption) + 1) % numOptions,
		)
		PlaySFX(ecs, cfg.SoundMenuNavigate)
	}

	// Handle selection
	if selectAction.JustPressed {
		PlaySFX(ecs, cfg.SoundMenuSelect)
		switch pause.SelectedOption {
		case components.MenuResume:
			pause.IsPaused = false
			ResumeMusic(ecs)
		case components.MenuSettings:
			log.Println("Settings clicked")
		case components.MenuExit:
			os.Exit(0)
		}
	}
}

// DrawPause renders the pause overlay and menu.
func DrawPause(ecs *ecs.ECS, screen *ebiten.Image) {
	pause := GetOrCreatePause(ecs)

	if !pause.IsPaused {
		return
	}

	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())

	// Draw semi-transparent overlay
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(width), float32(height),
		cfg.Pause.OverlayColor,
		false,
	)

	// Calculate menu positioning
	menuOptions := cfg.Pause.MenuOptions
	totalMenuHeight := float64(len(menuOptions)) * (cfg.Pause.MenuItemHeight + cfg.Pause.MenuItemGap)
	startY := (height - totalMenuHeight) / 2

	// Get font for text rendering (larger bold font)
	fontFace := fonts.ExcelBold.Get()

	// Draw menu options
	for i, option := range menuOptions {
		y := startY + float64(i)*(cfg.Pause.MenuItemHeight+cfg.Pause.MenuItemGap)

		// Determine color based on selection
		textColor := cfg.Pause.TextColorNormal
		if components.PauseMenuOption(i) == pause.SelectedOption {
			textColor = cfg.Pause.TextColorSelected
		}

		// Center text horizontally (approximate width calculation for 20pt font)
		textWidth := len(option) * 12
		x := int((width - float64(textWidth)) / 2)

		text.Draw(screen, option, fontFace, x, int(y)+int(cfg.Pause.MenuItemHeight), textColor)
	}
}

// WithPauseCheck wraps a system to skip execution when paused.
func WithPauseCheck(system ecs.System) ecs.System {
	return func(e *ecs.ECS) {
		if pause := GetOrCreatePause(e); pause.IsPaused {
			return
		}
		system(e)
	}
}

// GetOrCreatePause returns the singleton Pause component, creating if needed.
func GetOrCreatePause(ecs *ecs.ECS) *components.PauseData {
	if _, ok := components.Pause.First(ecs.World); !ok {
		ent := ecs.World.Entry(ecs.World.Create(components.Pause))
		components.Pause.SetValue(ent, components.PauseData{
			IsPaused:       false,
			SelectedOption: components.MenuResume,
		})
	}

	ent, _ := components.Pause.First(ecs.World)
	return components.Pause.Get(ent)
}
