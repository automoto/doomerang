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

const numSettingsOptions = int(components.SettingsOptBack) + 1

// UpdateSettingsMenu handles settings navigation and value changes.
func UpdateSettingsMenu(e *ecs.ECS) {
	settings := GetOrCreateSettingsMenu(e)

	if !settings.IsOpen {
		return
	}

	input := getOrCreateInput(e)

	// Handle controls screen separately
	if settings.ShowingControls {
		if GetAction(input, cfg.ActionMenuBack).JustPressed ||
			GetAction(input, cfg.ActionMenuSelect).JustPressed ||
			GetAction(input, cfg.ActionBoomerang).JustPressed ||
			GetAction(input, cfg.ActionPause).JustPressed {
			settings.ShowingControls = false
			PlaySFX(e, cfg.SoundMenuSelect)
		}
		return
	}

	// Navigate up
	if GetAction(input, cfg.ActionMenuUp).JustPressed {
		navigateUp(settings)
		PlaySFX(e, cfg.SoundMenuNavigate)
	}

	// Navigate down
	if GetAction(input, cfg.ActionMenuDown).JustPressed {
		navigateDown(settings)
		PlaySFX(e, cfg.SoundMenuNavigate)
	}

	// Adjust value left
	if GetAction(input, cfg.ActionMenuLeft).JustPressed {
		adjustValue(e, settings, -1)
	}

	// Adjust value right
	if GetAction(input, cfg.ActionMenuRight).JustPressed {
		adjustValue(e, settings, +1)
	}

	// Select/Enter - for toggles and Back button
	if GetAction(input, cfg.ActionMenuSelect).JustPressed {
		handleSelect(e, settings)
	}

	// B/Circle, Start, or Escape to go back
	if GetAction(input, cfg.ActionMenuBack).JustPressed ||
		GetAction(input, cfg.ActionBoomerang).JustPressed ||
		GetAction(input, cfg.ActionPause).JustPressed {
		closeSettings(e, settings)
	}
}

// navigateUp moves selection up, skipping hidden options
func navigateUp(s *components.SettingsMenuData) {
	for {
		s.SelectedOption = components.SettingsMenuOption(
			(int(s.SelectedOption) - 1 + numSettingsOptions) % numSettingsOptions,
		)
		if !isOptionHidden(s, s.SelectedOption) {
			break
		}
	}
}

// navigateDown moves selection down, skipping hidden options
func navigateDown(s *components.SettingsMenuData) {
	for {
		s.SelectedOption = components.SettingsMenuOption(
			(int(s.SelectedOption) + 1) % numSettingsOptions,
		)
		if !isOptionHidden(s, s.SelectedOption) {
			break
		}
	}
}

// isOptionHidden returns true if the option should be hidden
func isOptionHidden(s *components.SettingsMenuData, opt components.SettingsMenuOption) bool {
	// Hide resolution when fullscreen is enabled
	if opt == components.SettingsOptResolution && s.Fullscreen {
		return true
	}
	return false
}

// adjustValue changes the value for the selected option
func adjustValue(e *ecs.ECS, s *components.SettingsMenuData, direction int) {
	switch s.SelectedOption {
	case components.SettingsOptMusicVolume:
		s.MusicVolume = adjustVolumeStep(s.MusicVolume, direction)
		if !s.Muted {
			SetMusicVolume(e, s.MusicVolume)
		}
		PlaySFX(e, cfg.SoundMenuNavigate)

	case components.SettingsOptSFXVolume:
		s.SFXVolume = adjustVolumeStep(s.SFXVolume, direction)
		if !s.Muted {
			SetSFXVolume(e, s.SFXVolume)
		}
		// Play preview sound
		PlaySFX(e, cfg.SoundMenuSelect)

	case components.SettingsOptMute:
		toggleMute(e, s)
		PlaySFX(e, cfg.SoundMenuSelect)

	case components.SettingsOptFullscreen:
		toggleFullscreen(s)
		PlaySFX(e, cfg.SoundMenuSelect)

	case components.SettingsOptResolution:
		cycleResolution(s, direction)
		PlaySFX(e, cfg.SoundMenuNavigate)

	case components.SettingsOptInputMode:
		numModes := len(cfg.SettingsMenu.InputModes)
		s.InputMode = (s.InputMode + direction + numModes) % numModes
		PlaySFX(e, cfg.SoundMenuNavigate)
	}
}

// adjustVolumeStep adjusts volume by stepping through predefined values
func adjustVolumeStep(current float64, direction int) float64 {
	steps := cfg.SettingsMenu.VolumeSteps
	currentIdx := findClosestStepIndex(current, steps)
	newIdx := currentIdx + direction
	if newIdx < 0 {
		newIdx = 0
	}
	if newIdx >= len(steps) {
		newIdx = len(steps) - 1
	}
	return steps[newIdx]
}

// findClosestStepIndex finds the closest step index for a volume value
func findClosestStepIndex(value float64, steps []float64) int {
	closest := 0
	minDiff := 2.0 // Start with a large difference
	for i, step := range steps {
		diff := value - step
		if diff < 0 {
			diff = -diff
		}
		if diff < minDiff {
			minDiff = diff
			closest = i
		}
	}
	return closest
}

// toggleMute toggles the mute state
func toggleMute(e *ecs.ECS, s *components.SettingsMenuData) {
	s.Muted = !s.Muted
	if s.Muted {
		// Store current volumes and set to 0
		s.PreMuteMusicVol = s.MusicVolume
		s.PreMuteSFXVol = s.SFXVolume
		SetMusicVolume(e, 0)
		SetSFXVolume(e, 0)
	} else {
		// Restore volumes
		SetMusicVolume(e, s.MusicVolume)
		SetSFXVolume(e, s.SFXVolume)
	}
}

// toggleFullscreen toggles fullscreen mode
func toggleFullscreen(s *components.SettingsMenuData) {
	s.Fullscreen = !s.Fullscreen
	ebiten.SetFullscreen(s.Fullscreen)
}

// cycleResolution cycles through available resolutions
func cycleResolution(s *components.SettingsMenuData, direction int) {
	numResolutions := len(cfg.SettingsMenu.Resolutions)
	s.ResolutionIndex = (s.ResolutionIndex + direction + numResolutions) % numResolutions

	// Apply the resolution
	res := cfg.SettingsMenu.Resolutions[s.ResolutionIndex]
	ebiten.SetWindowSize(res.Width, res.Height)
}

// handleSelect handles the select/enter action
func handleSelect(e *ecs.ECS, s *components.SettingsMenuData) {
	switch s.SelectedOption {
	case components.SettingsOptMute:
		toggleMute(e, s)
		PlaySFX(e, cfg.SoundMenuSelect)

	case components.SettingsOptFullscreen:
		toggleFullscreen(s)
		PlaySFX(e, cfg.SoundMenuSelect)

	case components.SettingsOptControls:
		s.ShowingControls = true
		PlaySFX(e, cfg.SoundMenuSelect)

	case components.SettingsOptBack:
		closeSettings(e, s)
	}
}

// closeSettings closes the settings menu and saves settings
func closeSettings(e *ecs.ECS, s *components.SettingsMenuData) {
	s.IsOpen = false
	PlaySFX(e, cfg.SoundMenuSelect)
	// Save settings will be called here once persistence is implemented
	SaveCurrentSettings(s)
}

// DrawSettingsMenu renders the settings overlay.
func DrawSettingsMenu(e *ecs.ECS, screen *ebiten.Image) {
	settings := GetOrCreateSettingsMenu(e)

	if !settings.IsOpen {
		return
	}

	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())

	// Draw solid background
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(width), float32(height),
		cfg.Menu.BackgroundColor,
		false,
	)

	// Show controls screen if active
	if settings.ShowingControls {
		drawControlsScreen(e, screen, width, height)
		return
	}

	// Get font
	fontFace := fonts.ExcelBold.Get()
	titleFont := fonts.ExcelTitle.Get()

	// Draw title centered near top
	title := "SETTINGS"
	titleWidth := len(title) * 20
	titleX := int((width - float64(titleWidth)) / 2)
	text.Draw(screen, title, titleFont, titleX, 35, cfg.Menu.TitleColor)

	// Count visible options for layout calculation
	visibleCount := 0
	for opt := components.SettingsOptMusicVolume; opt <= components.SettingsOptBack; opt++ {
		if !isOptionHidden(settings, opt) {
			visibleCount++
		}
	}

	// Calculate menu positioning - center vertically in available space
	menuItemHeight := 24.0
	menuItemGap := 10.0
	totalMenuHeight := float64(visibleCount) * (menuItemHeight + menuItemGap)
	startY := (height-totalMenuHeight)/2 + 10 // Offset slightly down from center

	// Draw each option
	optionIndex := 0
	for opt := components.SettingsOptMusicVolume; opt <= components.SettingsOptBack; opt++ {
		if isOptionHidden(settings, opt) {
			continue
		}

		y := startY + float64(optionIndex)*(menuItemHeight+menuItemGap)

		// Determine color based on selection
		textColor := cfg.Pause.TextColorNormal
		if opt == settings.SelectedOption {
			textColor = cfg.Pause.TextColorSelected
		}

		// Get label and value for this option
		label, value := getOptionDisplay(settings, opt)

		// Draw label on left side (centered layout)
		labelX := int(width/2) - 120
		text.Draw(screen, label, fontFace, labelX, int(y)+int(menuItemHeight), textColor)

		// Draw value on right side (if not Back button)
		if opt != components.SettingsOptBack && opt != components.SettingsOptControls {
			valueX := int(width/2) + 40
			text.Draw(screen, value, fontFace, valueX, int(y)+int(menuItemHeight), textColor)
		} else if opt == components.SettingsOptControls {
			// Draw arrow for Controls option
			valueX := int(width/2) + 100
			text.Draw(screen, value, fontFace, valueX, int(y)+int(menuItemHeight), textColor)
		}

		optionIndex++
	}

	// Draw navigation hint at bottom based on input method
	input := getOrCreateInput(e)
	hint := getSettingsHint(input.LastInputMethod)
	hintFont := fonts.ExcelSmall.Get()
	hintWidth := len(hint) * 7
	hintX := int((width - float64(hintWidth)) / 2)
	text.Draw(screen, hint, hintFont, hintX, int(height)-12, cfg.Pause.TextColorNormal)
}

// drawControlsScreen renders the controls/button mapping screen
func drawControlsScreen(e *ecs.ECS, screen *ebiten.Image, width, height float64) {
	input := getOrCreateInput(e)
	fontFace := fonts.ExcelBold.Get()
	titleFont := fonts.ExcelTitle.Get()
	smallFont := fonts.ExcelSmall.Get()

	// Draw title
	title := "CONTROLS"
	titleWidth := len(title) * 20
	titleX := int((width - float64(titleWidth)) / 2)
	text.Draw(screen, title, titleFont, titleX, 35, cfg.Menu.TitleColor)

	// Get control mappings based on input method
	mappings := getControlMappings(input.LastInputMethod)

	// Calculate layout
	startY := 70.0
	lineHeight := 22.0
	labelX := int(width/2) - 100
	valueX := int(width/2) + 20

	for i, mapping := range mappings {
		y := startY + float64(i)*lineHeight
		text.Draw(screen, mapping.Action, fontFace, labelX, int(y), cfg.Pause.TextColorNormal)
		text.Draw(screen, mapping.Button, fontFace, valueX, int(y), cfg.Pause.TextColorSelected)
	}

	// Draw hint at bottom
	hint := getBackHint(input.LastInputMethod)
	hintWidth := len(hint) * 7
	hintX := int((width - float64(hintWidth)) / 2)
	text.Draw(screen, hint, smallFont, hintX, int(height)-12, cfg.Pause.TextColorNormal)
}

// controlMapping represents a single control mapping entry
type controlMapping struct {
	Action string
	Button string
}

// getControlMappings returns control mappings for the given input method
func getControlMappings(method components.InputMethod) []controlMapping {
	switch method {
	case components.InputPlayStation:
		return []controlMapping{
			{"Move", "Left Stick / D-Pad"},
			{"Jump", "Cross"},
			{"Attack", "Square"},
			{"Boomerang", "Circle"},
			{"Aim Throw", "Hold direction"},
			{"Slide", "Run + Down"},
			{"Wall Slide", "Hold toward wall"},
			{"Pause", "Options"},
		}
	case components.InputXbox:
		return []controlMapping{
			{"Move", "Left Stick / D-Pad"},
			{"Jump", "A"},
			{"Attack", "X"},
			{"Boomerang", "B"},
			{"Aim Throw", "Hold direction"},
			{"Slide", "Run + Down"},
			{"Wall Slide", "Hold toward wall"},
			{"Pause", "Start"},
		}
	default: // Keyboard
		return []controlMapping{
			{"Move", "Arrow Keys / WASD"},
			{"Jump", "X / W"},
			{"Attack", "Z"},
			{"Boomerang", "Space"},
			{"Aim Throw", "Hold direction"},
			{"Slide", "Run + Down"},
			{"Wall Slide", "Hold toward wall"},
			{"Pause", "Esc / P"},
		}
	}
}

// getSettingsHint returns the appropriate hint for settings menu
func getSettingsHint(method components.InputMethod) string {
	switch method {
	case components.InputPlayStation:
		return "Left Stick/D-Pad: Navigate   Left/Right: Change   Cross: Select   Circle: Back"
	case components.InputXbox:
		return "Left Stick/D-Pad: Navigate   Left/Right: Change   A: Select   B: Back"
	}
	return "Arrows: Navigate   Left/Right: Change   Enter: Select   Esc: Back"
}

// getBackHint returns the hint for going back
func getBackHint(method components.InputMethod) string {
	switch method {
	case components.InputPlayStation:
		return "Press any button to go back"
	case components.InputXbox:
		return "Press any button to go back"
	}
	return "Press any key to go back"
}

// getOptionDisplay returns the label and value display for an option
func getOptionDisplay(s *components.SettingsMenuData, opt components.SettingsMenuOption) (string, string) {
	switch opt {
	case components.SettingsOptMusicVolume:
		return "Music Volume", formatVolumeBar(s.MusicVolume)
	case components.SettingsOptSFXVolume:
		return "SFX Volume", formatVolumeBar(s.SFXVolume)
	case components.SettingsOptMute:
		return "Mute", formatToggle(s.Muted)
	case components.SettingsOptFullscreen:
		return "Fullscreen", formatToggle(s.Fullscreen)
	case components.SettingsOptResolution:
		if s.ResolutionIndex < len(cfg.SettingsMenu.Resolutions) {
			return "Resolution", cfg.SettingsMenu.Resolutions[s.ResolutionIndex].Label
		}
		return "Resolution", "Unknown"
	case components.SettingsOptInputMode:
		if s.InputMode < len(cfg.SettingsMenu.InputModes) {
			return "Input", cfg.SettingsMenu.InputModes[s.InputMode]
		}
		return "Input", "Unknown"
	case components.SettingsOptControls:
		return "Controls", ">"
	case components.SettingsOptBack:
		return "< Back", ""
	default:
		return "", ""
	}
}

// formatVolumeBar creates a visual volume bar
func formatVolumeBar(volume float64) string {
	percentage := int(volume * 100)
	filled := int(volume * 10)
	bar := ""
	for i := 0; i < 10; i++ {
		if i < filled {
			bar += "|"
		} else {
			bar += "."
		}
	}
	return fmt.Sprintf("[%s] %d%%", bar, percentage)
}

// formatToggle formats a boolean as On/Off
func formatToggle(value bool) string {
	if value {
		return "[X] On"
	}
	return "[ ] Off"
}

// isControllerConnected checks if any gamepad with standard layout is connected
func isControllerConnected() bool {
	gamepadIDs = ebiten.AppendGamepadIDs(gamepadIDs[:0])
	for _, gpID := range gamepadIDs {
		if ebiten.IsStandardGamepadLayoutAvailable(gpID) {
			return true
		}
	}
	return false
}

// GetOrCreateSettingsMenu returns the singleton SettingsMenu component, creating if needed.
func GetOrCreateSettingsMenu(e *ecs.ECS) *components.SettingsMenuData {
	if _, ok := components.SettingsMenu.First(e.World); !ok {
		ent := e.World.Entry(e.World.Create(components.SettingsMenu))

		// Initialize with current audio values
		musicVol := GetMusicVolume()
		sfxVol := GetSFXVolume()

		// Auto-detect input mode based on connected controllers
		inputMode := int(cfg.InputModeKeyboard)
		if isControllerConnected() {
			inputMode = int(cfg.InputModeController)
		}

		components.SettingsMenu.SetValue(ent, components.SettingsMenuData{
			IsOpen:          false,
			SelectedOption:  components.SettingsOptMusicVolume,
			OpenedFromPause: false,
			MusicVolume:     musicVol,
			SFXVolume:       sfxVol,
			Muted:           false,
			Fullscreen:      ebiten.IsFullscreen(),
			ResolutionIndex: cfg.SettingsMenu.DefaultResolutionIndex,
			InputMode:       inputMode,
			PreMuteMusicVol: musicVol,
			PreMuteSFXVol:   sfxVol,
		})
	}

	ent, _ := components.SettingsMenu.First(e.World)
	return components.SettingsMenu.Get(ent)
}

// OpenSettings opens the settings menu from a specific origin
func OpenSettings(e *ecs.ECS, fromPause bool) {
	settings := GetOrCreateSettingsMenu(e)
	settings.IsOpen = true
	settings.OpenedFromPause = fromPause
	settings.SelectedOption = components.SettingsOptMusicVolume

	// Sync current values
	settings.MusicVolume = GetMusicVolume()
	settings.SFXVolume = GetSFXVolume()
	settings.Fullscreen = ebiten.IsFullscreen()
}

// IsSettingsOpen returns true if the settings menu is currently open
func IsSettingsOpen(e *ecs.ECS) bool {
	settings := GetOrCreateSettingsMenu(e)
	return settings.IsOpen
}
