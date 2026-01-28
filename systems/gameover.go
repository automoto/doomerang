package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/yohamta/donburi/ecs"
)

// NewUpdateGameOver creates an UpdateGameOver system with scene transition capability
func NewUpdateGameOver(sceneChanger SceneChanger, createPlatformerScene func() interface{}, createMenuScene func() interface{}) ecs.System {
	return func(e *ecs.ECS) {
		gameOver := GetOrCreateGameOver(e)
		input := getOrCreateInput(e)

		// Navigate menu with wrap-around using modulo arithmetic
		numOptions := int(components.GameOverMenu) + 1
		if GetAction(input, cfg.ActionMenuUp).JustPressed {
			gameOver.SelectedOption = components.GameOverOption(
				(int(gameOver.SelectedOption) - 1 + numOptions) % numOptions,
			)
		}
		if GetAction(input, cfg.ActionMenuDown).JustPressed {
			gameOver.SelectedOption = components.GameOverOption(
				(int(gameOver.SelectedOption) + 1) % numOptions,
			)
		}

		// Handle selection
		if GetAction(input, cfg.ActionMenuSelect).JustPressed {
			switch gameOver.SelectedOption {
			case components.GameOverRetry:
				sceneChanger.ChangeScene(createPlatformerScene())
			case components.GameOverMenu:
				sceneChanger.ChangeScene(createMenuScene())
			}
		}
	}
}

// DrawGameOver renders the game over screen
func DrawGameOver(e *ecs.ECS, screen *ebiten.Image) {
	gameOver := GetOrCreateGameOver(e)

	width := float64(screen.Bounds().Dx())
	height := float64(screen.Bounds().Dy())

	// Draw background
	vector.DrawFilledRect(
		screen,
		0, 0,
		float32(width), float32(height),
		cfg.GameOver.BackgroundColor,
		false,
	)

	// Draw "GAME OVER" title
	titleFont := fonts.ExcelTitle.Get()
	title := "YOU DIED"
	titleWidth := len(title) * 20 // Approximate width for title font
	titleX := int((width - float64(titleWidth)) / 2)
	text.Draw(screen, title, titleFont, titleX, int(cfg.GameOver.TitleY), cfg.GameOver.TitleColor)

	// Draw menu options
	menuFont := fonts.ExcelBold.Get()
	menuOptions := cfg.GameOver.MenuOptions

	for i, option := range menuOptions {
		y := cfg.GameOver.MenuStartY + float64(i)*(cfg.GameOver.MenuItemHeight+cfg.GameOver.MenuItemGap)

		// Determine color based on selection
		textColor := cfg.GameOver.TextColorNormal
		if components.GameOverOption(i) == gameOver.SelectedOption {
			textColor = cfg.GameOver.TextColorSelected
		}

		// Center text horizontally
		textWidth := len(option) * 12
		x := int((width - float64(textWidth)) / 2)

		text.Draw(screen, option, menuFont, x, int(y)+int(cfg.GameOver.MenuItemHeight), textColor)
	}
}

// GetOrCreateGameOver returns the singleton GameOver component, creating if needed
func GetOrCreateGameOver(e *ecs.ECS) *components.GameOverData {
	if _, ok := components.GameOver.First(e.World); !ok {
		ent := e.World.Entry(e.World.Create(components.GameOver))
		components.GameOver.SetValue(ent, components.GameOverData{
			SelectedOption: components.GameOverRetry,
		})
	}

	ent, _ := components.GameOver.First(e.World)
	return components.GameOver.Get(ent)
}
