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

// SceneChanger allows systems to trigger scene transitions
type SceneChanger interface {
	ChangeScene(scene interface{})
}

// PlatformerSceneCreator creates a new platformer scene
type PlatformerSceneCreator interface {
	NewPlatformerScene() interface{}
}

// NewUpdateMenu creates an UpdateMenu system with scene transition capability
func NewUpdateMenu(sceneChanger SceneChanger, createPlatformerScene func() interface{}) ecs.System {
	return func(e *ecs.ECS) {
		// Skip menu input if settings is open
		if IsSettingsOpen(e) {
			return
		}

		menu := GetOrCreateMenu(e)
		input := getOrCreateInput(e)

		// Handle confirmation dialog
		if menu.ShowingConfirmDialog {
			handleConfirmDialog(e, menu, input, sceneChanger, createPlatformerScene)
			return
		}

		// Navigate menu with wrap-around
		numOptions := len(menu.VisibleOptions)
		if numOptions == 0 {
			return
		}

		if GetAction(input, cfg.ActionMenuUp).JustPressed {
			PlaySFX(e, cfg.SoundMenuNavigate)
			menu.SelectedIndex = (menu.SelectedIndex - 1 + numOptions) % numOptions
		}
		if GetAction(input, cfg.ActionMenuDown).JustPressed {
			PlaySFX(e, cfg.SoundMenuNavigate)
			menu.SelectedIndex = (menu.SelectedIndex + 1) % numOptions
		}

		// Handle selection
		if GetAction(input, cfg.ActionMenuSelect).JustPressed {
			PlaySFX(e, cfg.SoundMenuSelect)
			selectedOption := menu.VisibleOptions[menu.SelectedIndex]

			switch selectedOption {
			case components.MainMenuStart:
				if menu.HasSaveGame {
					// Show confirmation dialog
					menu.ShowingConfirmDialog = true
					menu.ConfirmSelection = 0 // Default to "No"
				} else {
					FadeOutMusic(e)
					ClearGameProgress()
					sceneChanger.ChangeScene(createPlatformerScene())
				}
			case components.MainMenuContinue:
				FadeOutMusic(e)
				sceneChanger.ChangeScene(createPlatformerScene())
			case components.MainMenuSettings:
				OpenSettings(e, false)
			case components.MainMenuExit:
				os.Exit(0)
			}
		}

		// Allow back/escape to exit
		if GetAction(input, cfg.ActionMenuBack).JustPressed {
			os.Exit(0)
		}
	}
}

// handleConfirmDialog processes input for the overwrite save confirmation
func handleConfirmDialog(e *ecs.ECS, menu *components.MenuData, input *components.InputData,
	sceneChanger SceneChanger, createPlatformerScene func() interface{}) {

	// Navigate between Yes/No
	if GetAction(input, cfg.ActionMenuLeft).JustPressed || GetAction(input, cfg.ActionMenuRight).JustPressed {
		PlaySFX(e, cfg.SoundMenuNavigate)
		menu.ConfirmSelection = 1 - menu.ConfirmSelection // Toggle 0<->1
	}

	// Cancel dialog
	if GetAction(input, cfg.ActionMenuBack).JustPressed {
		PlaySFX(e, cfg.SoundMenuNavigate)
		menu.ShowingConfirmDialog = false
		return
	}

	// Confirm selection
	if GetAction(input, cfg.ActionMenuSelect).JustPressed {
		PlaySFX(e, cfg.SoundMenuSelect)
		if menu.ConfirmSelection == 1 { // Yes
			FadeOutMusic(e)
			ClearGameProgress()
			sceneChanger.ChangeScene(createPlatformerScene())
		} else { // No
			menu.ShowingConfirmDialog = false
		}
	}
}

// DrawMenu renders the main menu screen
func DrawMenu(e *ecs.ECS, screen *ebiten.Image) {
	menu := GetOrCreateMenu(e)

	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())

	// Draw background
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(width), float32(height),
		cfg.Menu.BackgroundColor,
		false,
	)

	// Draw title
	titleFont := fonts.ExcelTitle.Get()
	title := "DOOMERANG"
	titleWidth := len(title) * 20 // Approximate width for 32pt font
	titleX := int((width - float64(titleWidth)) / 2)
	text.Draw(screen, title, titleFont, titleX, int(cfg.Menu.TitleY), cfg.Menu.TitleColor)

	// Draw menu options
	menuFont := fonts.ExcelBold.Get()

	for i, option := range menu.VisibleOptions {
		y := cfg.Menu.MenuStartY + float64(i)*(cfg.Menu.MenuItemHeight+cfg.Menu.MenuItemGap)

		// Determine color based on selection
		textColor := cfg.Menu.TextColorNormal
		if i == menu.SelectedIndex && !menu.ShowingConfirmDialog {
			textColor = cfg.Menu.TextColorSelected
		}

		// Get option label
		label := getOptionLabel(option)
		textWidth := len(label) * 12
		x := int((width - float64(textWidth)) / 2)

		text.Draw(screen, label, menuFont, x, int(y)+int(cfg.Menu.MenuItemHeight), textColor)
	}

	// Draw confirmation dialog if showing
	if menu.ShowingConfirmDialog {
		drawConfirmDialog(screen, menu, width, height)
	}

	// Draw navigation hint at bottom based on input method
	input := getOrCreateInput(e)
	hint := getMenuHint(input.LastInputMethod)
	hintFont := fonts.ExcelSmall.Get()
	hintWidth := len(hint) * 7
	hintX := int((width - float64(hintWidth)) / 2)
	text.Draw(screen, hint, hintFont, hintX, int(height)-12, cfg.Menu.TextColorNormal)
}

// getMenuHint returns the appropriate hint for menu navigation
func getMenuHint(method components.InputMethod) string {
	switch method {
	case components.InputPlayStation:
		return "Left Stick/D-Pad: Navigate   Cross: Select"
	case components.InputXbox:
		return "Left Stick/D-Pad: Navigate   A: Select"
	}
	return "Arrows: Navigate   Enter: Select"
}

// drawConfirmDialog renders the overwrite save confirmation dialog
func drawConfirmDialog(screen *ebiten.Image, menu *components.MenuData, width, height float64) {
	// Dialog dimensions with adequate padding
	dialogWidth := 320.0
	dialogHeight := 90.0
	dialogX := (width - dialogWidth) / 2
	dialogY := (height - dialogHeight) / 2

	// Draw dialog background
	vector.DrawFilledRect(
		screen,
		float32(dialogX), float32(dialogY),
		float32(dialogWidth), float32(dialogHeight),
		cfg.Menu.BackgroundColor,
		false,
	)

	// Draw dialog border
	vector.StrokeRect(
		screen,
		float32(dialogX), float32(dialogY),
		float32(dialogWidth), float32(dialogHeight),
		2,
		cfg.Menu.TextColorSelected,
		false,
	)

	menuFont := fonts.ExcelBold.Get()

	// Draw message centered in dialog
	message := cfg.Menu.ConfirmDialogMessage
	msgWidth := len(message) * 12
	msgX := int((width - float64(msgWidth)) / 2)
	msgY := int(dialogY) + 35
	text.Draw(screen, message, menuFont, msgX, msgY, cfg.Menu.TextColorNormal)

	// Draw Yes/No options
	noLabel := cfg.Menu.ConfirmDialogNo
	yesLabel := cfg.Menu.ConfirmDialogYes

	// Calculate positions for centered buttons
	buttonGap := 60.0
	centerX := width / 2
	noX := int(centerX - buttonGap - float64(len(noLabel)*12)/2)
	yesX := int(centerX + buttonGap - float64(len(yesLabel)*12)/2)
	buttonY := int(dialogY) + 70

	// Draw No option
	noColor := cfg.Menu.TextColorNormal
	if menu.ConfirmSelection == 0 {
		noColor = cfg.Menu.TextColorSelected
	}
	text.Draw(screen, noLabel, menuFont, noX, buttonY, noColor)

	// Draw Yes option
	yesColor := cfg.Menu.TextColorNormal
	if menu.ConfirmSelection == 1 {
		yesColor = cfg.Menu.TextColorSelected
	}
	text.Draw(screen, yesLabel, menuFont, yesX, buttonY, yesColor)
}

// getOptionLabel returns the display text for a menu option
func getOptionLabel(option components.MainMenuOption) string {
	switch option {
	case components.MainMenuStart:
		return "Start"
	case components.MainMenuContinue:
		return "Continue"
	case components.MainMenuSettings:
		return "Settings"
	case components.MainMenuExit:
		return "Exit"
	default:
		return ""
	}
}

// GetOrCreateMenu returns the singleton Menu component, creating if needed
func GetOrCreateMenu(e *ecs.ECS) *components.MenuData {
	if _, ok := components.Menu.First(e.World); !ok {
		// Check if save game exists
		hasSave := HasSaveGame()

		// Build visible options based on save state
		var visibleOptions []components.MainMenuOption
		if hasSave {
			// Save exists: Continue first, then Start
			visibleOptions = []components.MainMenuOption{
				components.MainMenuContinue,
				components.MainMenuStart,
				components.MainMenuSettings,
				components.MainMenuExit,
			}
		} else {
			// No save: only Start (no Continue)
			visibleOptions = []components.MainMenuOption{
				components.MainMenuStart,
				components.MainMenuSettings,
				components.MainMenuExit,
			}
		}

		ent := e.World.Entry(e.World.Create(components.Menu))
		components.Menu.SetValue(ent, components.MenuData{
			SelectedIndex:        0,
			VisibleOptions:       visibleOptions,
			HasSaveGame:          hasSave,
			ShowingConfirmDialog: false,
			ConfirmSelection:     0,
		})
	}

	ent, _ := components.Menu.First(e.World)
	return components.Menu.Get(ent)
}
