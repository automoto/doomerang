package components

import (
	"github.com/yohamta/donburi"
)

// SettingsMenuOption represents menu items in the settings menu
type SettingsMenuOption int

const (
	SettingsOptMusicVolume SettingsMenuOption = iota
	SettingsOptSFXVolume
	SettingsOptMute
	SettingsOptFullscreen
	SettingsOptResolution
	SettingsOptInputMode
	SettingsOptControls
	SettingsOptBack
)

// SettingsMenuData stores the current state of the settings menu overlay
type SettingsMenuData struct {
	IsOpen          bool
	SelectedOption  SettingsMenuOption
	OpenedFromPause bool // Track origin for "Back" navigation
	ShowingControls bool // True when displaying controls screen

	// Current settings values
	MusicVolume     float64 // 0.0, 0.25, 0.50, 0.75, 1.0
	SFXVolume       float64
	Muted           bool
	Fullscreen      bool
	ResolutionIndex int
	InputMode       int // 0 = Keyboard, 1 = Controller

	// For mute restore
	PreMuteMusicVol float64
	PreMuteSFXVol   float64
}

// SettingsMenu is the component type for settings menu state
var SettingsMenu = donburi.NewComponentType[SettingsMenuData]()
