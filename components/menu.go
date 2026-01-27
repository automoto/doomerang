package components

import "github.com/yohamta/donburi"

// MainMenuOption represents the available main menu selections
type MainMenuOption int

const (
	MainMenuStart MainMenuOption = iota
	MainMenuContinue
	MainMenuSettings
	MainMenuExit
)

// MenuData stores the current state of the main menu
type MenuData struct {
	SelectedIndex      int                // Current selection index in VisibleOptions
	VisibleOptions     []MainMenuOption   // Options to display (depends on save state)
	HasSaveGame        bool               // Whether a save game exists
	ShowingConfirmDialog bool             // Whether the overwrite confirmation is showing
	ConfirmSelection   int                // 0 = No, 1 = Yes
}

// Menu is the component type for main menu state
var Menu = donburi.NewComponentType[MenuData]()
