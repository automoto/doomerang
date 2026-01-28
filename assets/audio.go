package assets

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

//go:embed all:audio
var audioFS embed.FS

// AudioLoader handles loading and caching of audio assets
type AudioLoader struct {
	sfxCache map[string][]byte // Cache decoded audio bytes for SFX
	context  *audio.Context
}

// NewAudioLoader creates a new audio loader with the given context
func NewAudioLoader(ctx *audio.Context) *AudioLoader {
	return &AudioLoader{
		sfxCache: make(map[string][]byte),
		context:  ctx,
	}
}

// PreloadSFX decodes a sound effect and caches it without creating a player.
// Call this at startup to avoid decode lag on first play.
func (l *AudioLoader) PreloadSFX(path string) error {
	// Already cached
	if _, ok := l.sfxCache[path]; ok {
		return nil
	}

	// Load from embedded FS
	data, err := audioFS.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read audio file %s: %w", path, err)
	}

	// Decode based on file extension
	var decoded []byte
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".ogg":
		stream, err := vorbis.DecodeWithSampleRate(l.context.SampleRate(), bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decode ogg %s: %w", path, err)
		}
		decoded, err = io.ReadAll(stream)
		if err != nil {
			return fmt.Errorf("failed to read decoded audio %s: %w", path, err)
		}

	case ".wav":
		stream, err := wav.DecodeWithSampleRate(l.context.SampleRate(), bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("failed to decode wav %s: %w", path, err)
		}
		decoded, err = io.ReadAll(stream)
		if err != nil {
			return fmt.Errorf("failed to read decoded audio %s: %w", path, err)
		}

	default:
		return fmt.Errorf("unsupported audio format: %s", ext)
	}

	l.sfxCache[path] = decoded
	return nil
}

// LoadSFX loads a sound effect and returns a new player each time.
// SFX are cached as decoded bytes for instant playback.
func (l *AudioLoader) LoadSFX(path string) (*audio.Player, error) {
	// Check cache first
	if cachedBytes, ok := l.sfxCache[path]; ok {
		player, err := l.context.NewPlayer(bytes.NewReader(cachedBytes))
		return player, err
	}

	// Load from embedded FS
	data, err := audioFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio file %s: %w", path, err)
	}

	// Decode based on file extension
	var decoded []byte
	ext := strings.ToLower(filepath.Ext(path))

	switch ext {
	case ".ogg":
		stream, err := vorbis.DecodeWithSampleRate(l.context.SampleRate(), bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode ogg %s: %w", path, err)
		}
		decoded, err = io.ReadAll(stream)
		if err != nil {
			return nil, fmt.Errorf("failed to read decoded audio %s: %w", path, err)
		}

	case ".wav":
		stream, err := wav.DecodeWithSampleRate(l.context.SampleRate(), bytes.NewReader(data))
		if err != nil {
			return nil, fmt.Errorf("failed to decode wav %s: %w", path, err)
		}
		decoded, err = io.ReadAll(stream)
		if err != nil {
			return nil, fmt.Errorf("failed to read decoded audio %s: %w", path, err)
		}

	default:
		return nil, fmt.Errorf("unsupported audio format: %s", ext)
	}

	l.sfxCache[path] = decoded

	player, err := l.context.NewPlayer(bytes.NewReader(decoded))
	return player, err
}

// LoadMusic returns a streaming player for music with looping.
// Music is not cached - it streams from the embedded file.
func (l *AudioLoader) LoadMusic(path string) (*audio.Player, error) {
	data, err := audioFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read music file %s: %w", path, err)
	}

	// Decode OGG (music files are always OGG)
	stream, err := vorbis.DecodeWithSampleRate(l.context.SampleRate(), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode music ogg %s: %w", path, err)
	}

	// Create infinite loop for music
	loop := audio.NewInfiniteLoop(stream, stream.Length())

	player, err := l.context.NewPlayer(loop)
	return player, err
}

// LoadLevelMusic loads music from a level's music directory.
// Returns the player and the music path, or an error if no music found.
func (l *AudioLoader) LoadLevelMusic(levelName string) (*audio.Player, string, error) {
	// Level music is in the levels embed, not audio embed
	// We need to look in assets/levels/{levelName}/music/
	musicDir := fmt.Sprintf("levels/%s/music", levelName)

	entries, err := assetFS.ReadDir(musicDir)
	if err != nil {
		return nil, "", fmt.Errorf("no music directory for level %s: %w", levelName, err)
	}

	// Find first .ogg file
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext == ".ogg" {
			musicPath := filepath.Join(musicDir, entry.Name())

			// Load from assetFS (levels embed)
			data, err := assetFS.ReadFile(musicPath)
			if err != nil {
				return nil, "", fmt.Errorf("failed to read level music %s: %w", musicPath, err)
			}

			stream, err := vorbis.DecodeWithSampleRate(l.context.SampleRate(), bytes.NewReader(data))
			if err != nil {
				return nil, "", fmt.Errorf("failed to decode level music %s: %w", musicPath, err)
			}

			loop := audio.NewInfiniteLoop(stream, stream.Length())
			player, err := l.context.NewPlayer(loop)
			if err != nil {
				return nil, "", err
			}

			return player, musicPath, nil
		}
	}

	return nil, "", fmt.Errorf("no music found in %s", musicDir)
}
