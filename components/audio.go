package components

import (
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/yohamta/donburi"
)

// AudioData stores global audio state (singleton component)
type AudioData struct {
	Context          *audio.Context
	MusicPlayer      *audio.Player
	MusicVolume      float64 // 0.0 - 1.0
	SFXVolume        float64 // 0.0 - 1.0
	CurrentMusicKey  string  // Track which music is playing
	FadeOutTimer     int     // Frames remaining for fade out
	FadeOutDuration  int     // Total fade duration in frames
	FadeStartVolume  float64 // Volume at start of fade
	PendingSFX       []cfg.SoundID
}

var Audio = donburi.NewComponentType[AudioData]()
