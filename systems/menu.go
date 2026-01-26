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

		upAction := input.Actions[cfg.ActionMenuUp]
		downAction := input.Actions[cfg.ActionMenuDown]
		selectAction := input.Actions[cfg.ActionMenuSelect]

		// Navigate menu with wrap-around using modulo arithmetic
		numOptions := int(components.MainMenuExit) + 1
		if upAction.JustPressed {
			PlaySFX(e, cfg.SoundMenuNavigate)
			menu.SelectedOption = components.MainMenuOption(
				(int(menu.SelectedOption) - 1 + numOptions) % numOptions,
			)
		}
		if downAction.JustPressed {
			PlaySFX(e, cfg.SoundMenuNavigate)
			menu.SelectedOption = components.MainMenuOption(
				(int(menu.SelectedOption) + 1) % numOptions,
			)
		}

		// Handle selection
		if selectAction.JustPressed {
			PlaySFX(e, cfg.SoundMenuSelect)
			switch menu.SelectedOption {
			case components.MainMenuStart:
				FadeOutMusic(e)
				sceneChanger.ChangeScene(createPlatformerScene())
			case components.MainMenuContinue:
				log.Println("Continue clicked")
			case components.MainMenuLevelSelect:
				log.Println("Level Select clicked")
			case components.MainMenuSettings:
				OpenSettings(e, false)
			case components.MainMenuExit:
				os.Exit(0)
			}
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
	menuOptions := cfg.Menu.MenuOptions

	for i, option := range menuOptions {
		y := cfg.Menu.MenuStartY + float64(i)*(cfg.Menu.MenuItemHeight+cfg.Menu.MenuItemGap)

		// Determine color based on selection
		textColor := cfg.Menu.TextColorNormal
		if components.MainMenuOption(i) == menu.SelectedOption {
			textColor = cfg.Menu.TextColorSelected
		}

		// Center text horizontally (approximate width calculation for 20pt font)
		textWidth := len(option) * 12
		x := int((width - float64(textWidth)) / 2)

		text.Draw(screen, option, menuFont, x, int(y)+int(cfg.Menu.MenuItemHeight), textColor)
	}

	// Draw navigation hint at bottom
	hintFont := fonts.ExcelSmall.Get()
	hint := "Arrows: Navigate   Enter: Select"
	hintWidth := len(hint) * 7
	hintX := int((width - float64(hintWidth)) / 2)
	text.Draw(screen, hint, hintFont, hintX, int(height)-12, cfg.Menu.TextColorNormal)
}

// GetOrCreateMenu returns the singleton Menu component, creating if needed
func GetOrCreateMenu(e *ecs.ECS) *components.MenuData {
	if _, ok := components.Menu.First(e.World); !ok {
		ent := e.World.Entry(e.World.Create(components.Menu))
		components.Menu.SetValue(ent, components.MenuData{
			SelectedOption: components.MainMenuStart,
		})
	}

	ent, _ := components.Menu.First(e.World)
	return components.Menu.Get(ent)
}
