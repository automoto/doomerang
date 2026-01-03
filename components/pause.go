package components

import "github.com/yohamta/donburi"

// PauseMenuOption represents menu items in the pause menu
type PauseMenuOption int

const (
	MenuResume PauseMenuOption = iota
	MenuSettings
	MenuExit
)

// PauseData stores the pause state and menu selection
type PauseData struct {
	IsPaused       bool
	SelectedOption PauseMenuOption
}

var Pause = donburi.NewComponentType[PauseData]()
