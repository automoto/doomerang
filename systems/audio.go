package systems

import (
	"strings"
	"sync"

	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/yohamta/donburi/ecs"
)

// Global audio state - created once and shared across all scenes
var (
	globalAudioContext *audio.Context
	globalAudioLoader  *assets.AudioLoader
	globalMusicPlayer  *audio.Player
	globalMusicKey     string
	globalMusicVolume  float64 = cfg.Audio.DefaultMusicVol
	globalSFXVolume    float64 = cfg.Audio.DefaultSFXVol
	globalFadeTimer    int
	globalFadeDuration int
	globalFadeStart    float64
	audioInitOnce      sync.Once
)

// initGlobalAudio initializes the global audio context (called once)
func initGlobalAudio() {
	audioInitOnce.Do(func() {
		globalAudioContext = audio.NewContext(cfg.Audio.SampleRate)
		globalAudioLoader = assets.NewAudioLoader(globalAudioContext)
	})
}

// PreloadAllSFX decodes all sound effects at startup to avoid lag on first play.
// This is especially important for WASM where decoding is slower.
func PreloadAllSFX() {
	initGlobalAudio()

	for _, path := range cfg.Sound.SFXPaths {
		_ = globalAudioLoader.PreloadSFX(path)
	}
}

// UpdateAudio processes pending SFX and manages music transitions
func UpdateAudio(e *ecs.ECS) {
	initGlobalAudio()

	// Handle music fade out
	if globalFadeTimer > 0 {
		globalFadeTimer--
		if globalFadeDuration > 0 {
			progress := float64(globalFadeTimer) / float64(globalFadeDuration)
			if globalMusicPlayer != nil {
				globalMusicPlayer.SetVolume(globalFadeStart * progress)
			}
		}
		if globalFadeTimer == 0 && globalMusicPlayer != nil {
			_ = globalMusicPlayer.Close()
			globalMusicPlayer = nil
			globalMusicKey = ""
		}
	}

	// Process pending SFX from the ECS audio data (if exists)
	entry, ok := components.Audio.First(e.World)
	if ok {
		audioData := components.Audio.Get(entry)
		for _, soundID := range audioData.PendingSFX {
			playSFX(soundID)
		}
		audioData.PendingSFX = audioData.PendingSFX[:0]
	}
}

func playSFX(soundID cfg.SoundID) {
	if globalSFXVolume <= 0 {
		return
	}

	path, ok := cfg.Sound.SFXPaths[soundID]
	if !ok {
		return
	}

	player, err := globalAudioLoader.LoadSFX(path)
	if err != nil {
		return
	}

	volume := globalSFXVolume
	if mult, ok := cfg.Sound.VolumeMultipliers[soundID]; ok {
		volume *= mult
	}

	player.SetVolume(volume)
	player.Play()
}

// PlayMusic starts playing music with the given path (looping)
func PlayMusic(e *ecs.ECS, musicPath string) {
	initGlobalAudio()

	// Already playing this music
	if globalMusicKey == musicPath {
		return
	}

	// Stop current music
	if globalMusicPlayer != nil {
		_ = globalMusicPlayer.Close()
	}

	player, err := globalAudioLoader.LoadMusic(musicPath)
	if err != nil {
		return
	}

	player.SetVolume(globalMusicVolume)
	player.Play()

	globalMusicPlayer = player
	globalMusicKey = musicPath
	globalFadeTimer = 0
}

// PlayLevelMusic loads and plays music for a specific level
func PlayLevelMusic(e *ecs.ECS, levelPath string) {
	initGlobalAudio()

	// Extract level name from path (e.g., "levels/level01.tmx" -> "level01")
	levelName := extractLevelName(levelPath)
	if levelName == "" {
		return
	}

	player, path, err := globalAudioLoader.LoadLevelMusic(levelName)
	if err != nil {
		return
	}

	// Stop current music
	if globalMusicPlayer != nil {
		_ = globalMusicPlayer.Close()
	}

	globalMusicPlayer = player
	globalMusicKey = path
	globalFadeTimer = 0
	player.SetVolume(globalMusicVolume)
	player.Play()
}

// extractLevelName extracts the level directory name from a level path
func extractLevelName(levelPath string) string {
	// Handle paths like "levels/level01.tmx" -> "level01"
	// or "levels/level01/level01.tmx" -> "level01"
	parts := strings.Split(levelPath, "/")
	for _, part := range parts {
		// Match "level" followed by a digit (level01, level02, etc.)
		// This excludes "levels" directory name
		if strings.HasPrefix(part, "level") && len(part) > 5 && part[5] >= '0' && part[5] <= '9' {
			// Remove .tmx extension if present
			return strings.TrimSuffix(part, ".tmx")
		}
	}
	return ""
}

// FadeOutMusic starts a music fade out transition
func FadeOutMusic(e *ecs.ECS) {
	if globalMusicPlayer == nil {
		return
	}
	globalFadeTimer = cfg.Audio.MusicFadeDuration
	globalFadeDuration = cfg.Audio.MusicFadeDuration
	globalFadeStart = globalMusicVolume
}

// StopMusic immediately stops the current music
func StopMusic(e *ecs.ECS) {
	if globalMusicPlayer != nil {
		_ = globalMusicPlayer.Close()
		globalMusicPlayer = nil
		globalMusicKey = ""
	}
	globalFadeTimer = 0
}

// PauseMusic pauses the current music playback
func PauseMusic(e *ecs.ECS) {
	if globalMusicPlayer != nil {
		globalMusicPlayer.Pause()
	}
}

// ResumeMusic resumes paused music playback
func ResumeMusic(e *ecs.ECS) {
	if globalMusicPlayer != nil {
		globalMusicPlayer.Play()
	}
}

// PlaySFX queues a sound effect to be played
func PlaySFX(e *ecs.ECS, sound cfg.SoundID) {
	initGlobalAudio()

	// Get or create audio data for this ECS to queue SFX
	audioData := GetOrCreateAudio(e)
	audioData.PendingSFX = append(audioData.PendingSFX, sound)
}

// SetMusicVolume changes the music volume (0.0 - 1.0)
func SetMusicVolume(e *ecs.ECS, volume float64) {
	globalMusicVolume = volume
	if globalMusicPlayer != nil && globalFadeTimer == 0 {
		globalMusicPlayer.SetVolume(volume)
	}
}

// SetSFXVolume changes the SFX volume (0.0 - 1.0)
func SetSFXVolume(e *ecs.ECS, volume float64) {
	globalSFXVolume = volume
}

// GetMusicVolume returns the current music volume (0.0 - 1.0)
func GetMusicVolume() float64 {
	return globalMusicVolume
}

// GetSFXVolume returns the current SFX volume (0.0 - 1.0)
func GetSFXVolume() float64 {
	return globalSFXVolume
}

// GetOrCreateAudio returns the singleton Audio component for this ECS, creating it if needed
func GetOrCreateAudio(e *ecs.ECS) *components.AudioData {
	initGlobalAudio()

	entry, ok := components.Audio.First(e.World)
	if !ok {
		entry = e.World.Entry(e.World.Create(components.Audio))
		components.Audio.SetValue(entry, components.AudioData{
			Context:     globalAudioContext,
			MusicVolume: globalMusicVolume,
			SFXVolume:   globalSFXVolume,
			PendingSFX:  make([]cfg.SoundID, 0, 8),
		})
	}
	return components.Audio.Get(entry)
}
