package systems

import (
	"encoding/json"
	"log"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/quasilyte/gdata"
	"github.com/yohamta/donburi/ecs"
)

// SavedSettings represents the settings data stored on disk
type SavedSettings struct {
	MusicVolume     float64 `json:"musicVolume"`
	SFXVolume       float64 `json:"sfxVolume"`
	Muted           bool    `json:"muted"`
	Fullscreen      bool    `json:"fullscreen"`
	ResolutionIndex int     `json:"resolutionIndex"`
	InputMode       int     `json:"inputMode"`
}

var gdataManager *gdata.Manager
var gdataInitialized bool

// InitPersistence initializes the gdata manager for settings storage
func InitPersistence() error {
	m, err := gdata.Open(gdata.Config{
		AppName: "doomerang",
	})
	if err != nil {
		log.Printf("Warning: Could not initialize persistence: %v", err)
		return err
	}
	gdataManager = m
	gdataInitialized = true
	return nil
}

// LoadSettings loads settings from disk
func LoadSettings() (*SavedSettings, error) {
	if !gdataInitialized || gdataManager == nil {
		return nil, nil
	}

	data, err := gdataManager.LoadItem("settings")
	if err != nil {
		log.Printf("Warning: Could not load settings: %v", err)
		return nil, nil
	}
	if data == nil {
		// No saved settings yet, use defaults
		return nil, nil
	}

	var settings SavedSettings
	if err := json.Unmarshal(data, &settings); err != nil {
		log.Printf("Warning: Could not parse saved settings: %v", err)
		return nil, err
	}

	return &settings, nil
}

// SaveSettings saves settings to disk
func SaveSettings(s *SavedSettings) error {
	if !gdataInitialized || gdataManager == nil {
		return nil
	}

	data, err := json.Marshal(s)
	if err != nil {
		log.Printf("Warning: Could not serialize settings: %v", err)
		return err
	}

	if err := gdataManager.SaveItem("settings", data); err != nil {
		log.Printf("Warning: Could not save settings: %v", err)
		return err
	}
	return nil
}

// SaveCurrentSettings saves the current settings from the SettingsMenuData component
func SaveCurrentSettings(s *components.SettingsMenuData) {
	saved := &SavedSettings{
		MusicVolume:     s.MusicVolume,
		SFXVolume:       s.SFXVolume,
		Muted:           s.Muted,
		Fullscreen:      s.Fullscreen,
		ResolutionIndex: s.ResolutionIndex,
		InputMode:       s.InputMode,
	}
	_ = SaveSettings(saved)
}

// ApplySavedSettings applies loaded settings to the game systems
func ApplySavedSettings(e *ecs.ECS, saved *SavedSettings) {
	if saved == nil {
		return
	}

	// Apply audio settings
	SetMusicVolume(e, saved.MusicVolume)
	SetSFXVolume(e, saved.SFXVolume)

	// Apply mute
	if saved.Muted {
		SetMusicVolume(e, 0)
		SetSFXVolume(e, 0)
	}

	// Apply fullscreen
	ebiten.SetFullscreen(saved.Fullscreen)

	// Apply resolution (only if not fullscreen)
	if !saved.Fullscreen && saved.ResolutionIndex < len(cfg.SettingsMenu.Resolutions) {
		res := cfg.SettingsMenu.Resolutions[saved.ResolutionIndex]
		ebiten.SetWindowSize(res.Width, res.Height)
	}

	// Update settings menu component if it exists
	if entry, ok := components.SettingsMenu.First(e.World); ok {
		settings := components.SettingsMenu.Get(entry)
		settings.MusicVolume = saved.MusicVolume
		settings.SFXVolume = saved.SFXVolume
		settings.Muted = saved.Muted
		settings.Fullscreen = saved.Fullscreen
		settings.ResolutionIndex = saved.ResolutionIndex
		settings.InputMode = saved.InputMode
		if saved.Muted {
			settings.PreMuteMusicVol = saved.MusicVolume
			settings.PreMuteSFXVol = saved.SFXVolume
		}
	}
}

// ApplySavedSettingsGlobal applies settings without needing an ECS reference
// Used during initial game startup before scenes are created
func ApplySavedSettingsGlobal(saved *SavedSettings) {
	if saved == nil {
		return
	}

	// Apply audio settings using the global variables directly
	// Note: We can't use SetMusicVolume/SetSFXVolume here as they need ECS
	// Instead we'll set globals and let the first scene pick them up
	globalMusicVolume = saved.MusicVolume
	globalSFXVolume = saved.SFXVolume

	if saved.Muted {
		globalMusicVolume = 0
		globalSFXVolume = 0
	}

	// Apply fullscreen
	ebiten.SetFullscreen(saved.Fullscreen)

	// Apply resolution (only if not fullscreen)
	if !saved.Fullscreen && saved.ResolutionIndex < len(cfg.SettingsMenu.Resolutions) {
		res := cfg.SettingsMenu.Resolutions[saved.ResolutionIndex]
		ebiten.SetWindowSize(res.Width, res.Height)
	}
}

type SavedGameProgress struct {
	LevelIndex       int     `json:"levelIndex"`
	CheckpointID     float64 `json:"checkpointId"`
	CheckpointSpawnX float64 `json:"checkpointSpawnX"`
	CheckpointSpawnY float64 `json:"checkpointSpawnY"`
}

func LoadGameProgress() (*SavedGameProgress, error) {
	if !gdataInitialized || gdataManager == nil {
		return nil, nil
	}

	data, err := gdataManager.LoadItem("progress")
	if err != nil {
		log.Printf("Warning: Could not load game progress: %v", err)
		return nil, nil
	}
	if data == nil || len(data) == 0 {
		return nil, nil
	}

	var progress SavedGameProgress
	if err := json.Unmarshal(data, &progress); err != nil {
		log.Printf("Warning: Could not parse saved progress: %v", err)
		return nil, err
	}

	return &progress, nil
}

func SaveGameProgress(levelIndex int, checkpoint *components.ActiveCheckpointData) error {
	if !gdataInitialized || gdataManager == nil || checkpoint == nil {
		return nil
	}

	progress := &SavedGameProgress{
		LevelIndex:       levelIndex,
		CheckpointID:     checkpoint.CheckpointID,
		CheckpointSpawnX: checkpoint.SpawnX,
		CheckpointSpawnY: checkpoint.SpawnY,
	}

	data, err := json.Marshal(progress)
	if err != nil {
		log.Printf("Warning: Could not serialize game progress: %v", err)
		return err
	}

	if err := gdataManager.SaveItem("progress", data); err != nil {
		log.Printf("Warning: Could not save game progress: %v", err)
		return err
	}

	return nil
}

// HasSaveGame returns true if a saved game progress exists
func HasSaveGame() bool {
	if !gdataInitialized || gdataManager == nil {
		return false
	}

	data, err := gdataManager.LoadItem("progress")
	if err != nil || data == nil || len(data) == 0 {
		return false
	}

	return true
}

// ClearGameProgress removes any saved game progress
func ClearGameProgress() error {
	if !gdataInitialized || gdataManager == nil {
		return nil
	}

	// Save empty/nil data to clear the progress
	if err := gdataManager.SaveItem("progress", nil); err != nil {
		log.Printf("Warning: Could not clear game progress: %v", err)
		return err
	}

	return nil
}
