package config

// SoundID represents a logical sound effect
type SoundID int

const (
	SoundNone SoundID = iota
	// Combat sounds
	SoundPunch
	SoundKick
	SoundHit
	SoundDeath
	// Movement sounds
	SoundJump
	SoundLand
	SoundWallAttach
	SoundSlide
	// Boomerang sounds
	SoundBoomerangThrow
	SoundBoomerangCatch
	SoundBoomerangImpact
	SoundBoomerangCharge
	// UI sounds
	SoundMenuNavigate
	SoundMenuSelect
)

// AudioConfig contains audio-related configuration values
type AudioConfig struct {
	SampleRate        int
	DefaultMusicVol   float64
	DefaultSFXVol     float64
	MusicFadeDuration int // frames for music fade out (60 = 1 second at 60fps)
}

// SoundConfig maps sound IDs to file paths
type SoundConfig struct {
	MenuMusic         string
	SFXPaths          map[SoundID]string
	VolumeMultipliers map[SoundID]float64
}

var Audio AudioConfig
var Sound SoundConfig

func init() {
	Audio = AudioConfig{
		SampleRate:        44100,
		DefaultMusicVol:   0.75,
		DefaultSFXVol:     1.0,
		MusicFadeDuration: 60,
	}

	Sound = SoundConfig{
		MenuMusic: "audio/music/menu.ogg",
		SFXPaths: map[SoundID]string{
			SoundPunch:           "audio/sfx/punch.wav",
			SoundKick:            "audio/sfx/kick.wav",
			SoundHit:             "audio/sfx/hit.wav",
			SoundDeath:           "audio/sfx/death.wav",
			SoundJump:            "audio/sfx/jump.wav",
			SoundLand:            "audio/sfx/land.wav",
			SoundWallAttach:      "audio/sfx/wall_attach.wav",
			SoundSlide:           "audio/sfx/slide.wav",
			SoundBoomerangThrow:  "audio/sfx/boomerang_throw.wav",
			SoundBoomerangCatch:  "audio/sfx/boomerang_catch.wav",
			SoundBoomerangImpact: "audio/sfx/boomerang_impact.wav",
			SoundBoomerangCharge: "audio/sfx/boomerang_charge.wav",
			SoundMenuNavigate:    "audio/sfx/menu_navigate.wav",
			SoundMenuSelect:      "audio/sfx/menu_select.wav",
		},
		VolumeMultipliers: map[SoundID]float64{
			SoundHit:             1.5,
			SoundBoomerangImpact: 1.5,
		},
	}
}
