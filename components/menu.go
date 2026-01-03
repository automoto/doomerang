package components

import "github.com/yohamta/donburi"

// MainMenuOption represents the available main menu selections
type MainMenuOption int

const (
	MainMenuStart MainMenuOption = iota
	MainMenuContinue
	MainMenuLevelSelect
	MainMenuSettings
	MainMenuExit
)

// MenuData stores the current state of the main menu
type MenuData struct {
	SelectedOption MainMenuOption
}

// Menu is the component type for main menu state
var Menu = donburi.NewComponentType[MenuData]()
