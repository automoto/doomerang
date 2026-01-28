package components

import "github.com/yohamta/donburi"

// LevelCompleteData stores the state of the level complete overlay
type LevelCompleteData struct {
	IsComplete     bool
	SelectedOption int // For potential future menu options
}

var LevelComplete = donburi.NewComponentType[LevelCompleteData]()
