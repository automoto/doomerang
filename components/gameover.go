package components

import "github.com/yohamta/donburi"

// GameOverOption represents the available game over menu selections
type GameOverOption int

const (
	GameOverRetry GameOverOption = iota
	GameOverMenu
)

// GameOverData stores the current state of the game over menu
type GameOverData struct {
	SelectedOption GameOverOption
}

// GameOver is the component type for game over menu state
var GameOver = donburi.NewComponentType[GameOverData]()
