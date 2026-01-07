package config

// Resolution represents a display resolution option
type Resolution struct {
	Width  int
	Height int
	Label  string
}

// InputModeID represents the input mode
type InputModeID int

const (
	InputModeKeyboard InputModeID = iota
	InputModeController
)

// SettingsMenuConfig contains settings screen configuration
type SettingsMenuConfig struct {
	Resolutions            []Resolution
	DefaultResolutionIndex int
	VolumeSteps            []float64
	InputModes             []string
}

// SettingsMenu is the global settings menu configuration
var SettingsMenu SettingsMenuConfig

func init() {
	SettingsMenu = SettingsMenuConfig{
		Resolutions: []Resolution{
			{Width: 1280, Height: 720, Label: "1280 x 720"},
			{Width: 1600, Height: 900, Label: "1600 x 900"},
			{Width: 1920, Height: 1080, Label: "1920 x 1080"},
			{Width: 2560, Height: 1440, Label: "2560 x 1440"},
		},
		DefaultResolutionIndex: 0,
		VolumeSteps:            []float64{0, 0.25, 0.5, 0.75, 1.0},
		InputModes:             []string{"Keyboard", "Controller"},
	}
}
