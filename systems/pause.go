package systems

import (
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

	// Toggle pause on ESC or P
	if GetAction(input, cfg.ActionPause).JustPressed {
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

	// Skip pause menu input if settings is open
	if IsSettingsOpen(ecs) {
		return
	}

	// Navigate menu with wrap-around using modulo arithmetic
	numOptions := int(components.MenuExit) + 1
	if GetAction(input, cfg.ActionMenuUp).JustPressed {
		pause.SelectedOption = components.PauseMenuOption(
			(int(pause.SelectedOption) - 1 + numOptions) % numOptions,
		)
		PlaySFX(ecs, cfg.SoundMenuNavigate)
	}
	if GetAction(input, cfg.ActionMenuDown).JustPressed {
		pause.SelectedOption = components.PauseMenuOption(
			(int(pause.SelectedOption) + 1) % numOptions,
		)
		PlaySFX(ecs, cfg.SoundMenuNavigate)
	}

	// Handle selection
	if GetAction(input, cfg.ActionMenuSelect).JustPressed {
		PlaySFX(ecs, cfg.SoundMenuSelect)
		switch pause.SelectedOption {
		case components.MenuResume:
			pause.IsPaused = false
			ResumeMusic(ecs)
		case components.MenuSettings:
			OpenSettings(ecs, true)
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

	// Draw navigation hint at bottom based on input method
	input := getOrCreateInput(ecs)
	hint := getPauseHint(input.LastInputMethod)
	hintFont := fonts.ExcelSmall.Get()
	hintWidth := len(hint) * 7
	hintX := int((width - float64(hintWidth)) / 2)
	text.Draw(screen, hint, hintFont, hintX, int(height)-12, cfg.Pause.TextColorNormal)
}

// getPauseHint returns the appropriate hint for pause menu
func getPauseHint(method components.InputMethod) string {
	switch method {
	case components.InputPlayStation:
		return "Left Stick/D-Pad: Navigate   Cross: Select   Options: Resume"
	case components.InputXbox:
		return "Left Stick/D-Pad: Navigate   A: Select   Start: Resume"
	}
	return "Arrows: Navigate   Enter: Select   Esc: Resume"
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
