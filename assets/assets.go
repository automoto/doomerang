package assets

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	"path/filepath"

	"github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
	"github.com/yohamta/donburi/features/math"
)

var (
	//go:embed all:levels
	assetFS embed.FS

	//go:embed all:images
	animationFS embed.FS
)

type PlayerSpawn struct {
	X          float64
	Y          float64
	SpawnPoint string
}

type Level struct {
	Background   *ebiten.Image
	SolidTiles   []SolidTile           // Collision tiles from wg-tiles layer
	PatrolPaths  map[string]PatrolPath // New field for patrol paths
	EnemySpawns  []EnemySpawn
	PlayerSpawns []PlayerSpawn
	DeadZones    []DeadZone
	Checkpoints  []CheckpointSpawn
	Fires        []FireSpawn
	Messages     []MessageSpawn
	FinishLines  []FinishLineSpawn
	Name         string
	Width        int
	Height       int
}

// SolidTile represents a solid collision tile
type SolidTile struct {
	X, Y, Width, Height float64
	SlopeType           string // "", "45_up_right", "45_up_left"
}

type EnemySpawn struct {
	X          float64
	Y          float64
	EnemyType  string
	PatrolPath string
}

type DeadZone struct {
	X, Y, Width, Height float64
}

type CheckpointSpawn struct {
	X, Y, Width, Height float64
	CheckpointID        float64
}

type FireSpawn struct {
	X, Y      float64
	FireType  string // "fire_pulsing" or "fire_continuous"
	Direction string // "up", "down", "left", "right" (default: "right")
}

type MessageSpawn struct {
	X, Y      float64
	MessageID float64
}

type FinishLineSpawn struct {
	X, Y, Width, Height float64
}

type LevelLoader struct{}

func NewLevelLoader() *LevelLoader {
	return &LevelLoader{}
}

type Path struct {
	Points []math.Vec2
	Loops  bool
}

type PatrolPath struct {
	Name   string
	Points []math.Vec2 // Converted polyline points to world coordinates
}

type AnimationLoader struct {
	cache      map[string]*ebiten.Image
	frameCache map[string]*ebiten.Image
}

func NewAnimationLoader() *AnimationLoader {
	return &AnimationLoader{
		cache:      make(map[string]*ebiten.Image),
		frameCache: make(map[string]*ebiten.Image),
	}
}

func (l *AnimationLoader) MustLoadImage(path string) *ebiten.Image {
	if img, ok := l.cache[path]; ok {
		return img
	}

	imgBytes, err := animationFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read image file %s: %v", path, err))
	}

	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(imgBytes))
	if err != nil {
		panic(fmt.Sprintf("Failed to create image from bytes for %s: %v", path, err))
	}

	l.cache[path] = img

	return img
}

// GetFrame returns a cached sub-image for a specific animation frame.
// This prevents creating thousands of duplicate *ebiten.Image structs for the same frame.
func (l *AnimationLoader) GetFrame(dir string, state config.StateID, frameIndex int, srcRect image.Rectangle) *ebiten.Image {
	key := fmt.Sprintf("%s/%s/%d", dir, state.String(), frameIndex)
	if img, ok := l.frameCache[key]; ok {
		return img
	}

	// Load the full sprite sheet
	sheetPath := fmt.Sprintf("images/spritesheets/%s/%s.png", dir, state.String())
	sheet := l.MustLoadImage(sheetPath)

	// Create the sub-image for this frame
	frame := sheet.SubImage(srcRect).(*ebiten.Image)
	l.frameCache[key] = frame

	return frame
}

func GetObjectImage(name string) *ebiten.Image {
	path := fmt.Sprintf("images/objects/%s", name)
	return animationLoader.MustLoadImage(path)
}

func GetIconImage(name string) *ebiten.Image {
	path := fmt.Sprintf("images/icons/%s", name)
	return animationLoader.MustLoadImage(path)
}

func GetFrame(dir string, state config.StateID, frameIndex int, srcRect image.Rectangle) *ebiten.Image {
	return animationLoader.GetFrame(dir, state, frameIndex, srcRect)
}

func (l *LevelLoader) MustLoadLevels() []Level {
	entries, err := assetFS.ReadDir("levels")
	if err != nil {
		panic(fmt.Sprintf("Failed to read levels directory: %v", err))
	}

	var levels []Level
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".tmx" {
			levelPath := filepath.Join("levels", entry.Name())
			level := l.MustLoadLevel(levelPath)
			levels = append(levels, level)
		}
	}

	if len(levels) == 0 {
		panic("No level files found in assets/levels directory")
	}

	return levels
}

func (l *LevelLoader) MustLoadLevel(levelPath string) Level {
	levelMap, err := tiled.LoadFile(levelPath, tiled.WithFileSystem(assetFS))
	if err != nil {
		panic(err)
	}

	level := Level{
		SolidTiles:   []SolidTile{},
		PatrolPaths:  make(map[string]PatrolPath),
		EnemySpawns:  []EnemySpawn{},
		PlayerSpawns: []PlayerSpawn{},
		DeadZones:    []DeadZone{},
		Checkpoints:  []CheckpointSpawn{},
		Fires:        []FireSpawn{},
		Messages:     []MessageSpawn{},
		FinishLines:  []FinishLineSpawn{},
		Name:         levelPath,
		Width:        levelMap.Width * levelMap.TileWidth,
		Height:       levelMap.Height * levelMap.TileHeight,
	}

	// Parse object groups for spawns, paths, and dead zones
	for _, og := range levelMap.ObjectGroups {
		switch og.Name {
		case "EnemySpawn":
			for _, o := range og.Objects {
				enemyType := o.Properties.GetString("enemyType")
				patrolPath := o.Properties.GetString("pathName")
				level.EnemySpawns = append(level.EnemySpawns, EnemySpawn{
					X:          o.X,
					Y:          o.Y,
					EnemyType:  enemyType,
					PatrolPath: patrolPath,
				})
			}
		case "PlayerSpawn":
			for _, o := range og.Objects {
				spawnPoint := o.Properties.GetString("spawnPoint")
				level.PlayerSpawns = append(level.PlayerSpawns, PlayerSpawn{
					X:          o.X,
					Y:          o.Y,
					SpawnPoint: spawnPoint,
				})
			}
		case "PatrolPaths":
			// Parse patrol paths from polyline objects
			for _, o := range og.Objects {
				if len(o.PolyLines) > 0 {
					// Use the first polyline if multiple polylines exist
					polyline := o.PolyLines[0]
					if polyline.Points != nil && len(*polyline.Points) >= 2 {
						// Convert polyline points to world coordinates
						points := make([]math.Vec2, len(*polyline.Points))
						for i, point := range *polyline.Points {
							points[i] = math.Vec2{
								X: o.X + point.X,
								Y: o.Y + point.Y,
							}
						}
						level.PatrolPaths[o.Name] = PatrolPath{
							Name:   o.Name,
							Points: points,
						}
					}
				}
			}
		case "DeadZones":
			for _, o := range og.Objects {
				level.DeadZones = append(level.DeadZones, DeadZone{
					X:      o.X,
					Y:      o.Y,
					Width:  o.Width,
					Height: o.Height,
				})
			}
		case "Checkpoint":
			for _, o := range og.Objects {
				checkpointID := o.Properties.GetFloat("checkpointID")
				level.Checkpoints = append(level.Checkpoints, CheckpointSpawn{
					X:            o.X,
					Y:            o.Y,
					Width:        o.Width,
					Height:       o.Height,
					CheckpointID: checkpointID,
				})
			}
		case "Obstacles":
			for _, o := range og.Objects {
				// Parse fire obstacles by object type
				// Note: Using Type field as TMX files use type= attribute
				fireType := o.Class
				if fireType == "" {
					fireType = o.Type //nolint:staticcheck // TMX uses type= attribute
				}
				if fireType == "fire_pulsing" || fireType == "fire_continuous" {
					// Get direction from Tiled properties, default to "right"
					direction := o.Properties.GetString("Direction")
					if direction == "" {
						direction = "right"
					}
					level.Fires = append(level.Fires, FireSpawn{
						X:         o.X,
						Y:         o.Y,
						FireType:  fireType,
						Direction: direction,
					})
				}
			}
		case "Messages":
			for _, o := range og.Objects {
				messageID := o.Properties.GetFloat("message_id")
				level.Messages = append(level.Messages, MessageSpawn{
					X:         o.X,
					Y:         o.Y,
					MessageID: messageID,
				})
			}
		case "FinishLine":
			for _, o := range og.Objects {
				level.FinishLines = append(level.FinishLines, FinishLineSpawn{
					X:      o.X,
					Y:      o.Y,
					Width:  o.Width,
					Height: o.Height,
				})
			}
		}
	}

	// Parse solid tiles from wg-tiles layer for collision
	tileW := float64(levelMap.TileWidth)
	tileH := float64(levelMap.TileHeight)
	for _, layer := range levelMap.Layers {
		if layer.Name != "wg-tiles" {
			continue
		}
		for y := 0; y < levelMap.Height; y++ {
			for x := 0; x < levelMap.Width; x++ {
				tileIndex := y*levelMap.Width + x
				tile := layer.Tiles[tileIndex]
				if tile.IsNil() {
					continue
				}

				// Get slope type from tileset tile properties
				var slopeType string
				if tilesetTile, err := tile.Tileset.GetTilesetTile(tile.ID); err == nil {
					slopeType = tilesetTile.Properties.GetString("slope")
				}

				level.SolidTiles = append(level.SolidTiles, SolidTile{
					X:         float64(x) * tileW,
					Y:         float64(y) * tileH,
					Width:     tileW,
					Height:    tileH,
					SlopeType: slopeType,
				})
			}
		}
		break
	}

	// Create a new image for the background
	level.Background = ebiten.NewImage(levelMap.Width*levelMap.TileWidth, levelMap.Height*levelMap.TileHeight)

	// Render image layers first (backgrounds)
	for _, imgLayer := range levelMap.ImageLayers {
		shouldRender := imgLayer.Properties.GetBool("render")
		if !shouldRender || imgLayer.Image == nil {
			continue
		}

		// Load image from embedded filesystem
		imgPath := filepath.Join("levels", imgLayer.Image.Source)
		imgBytes, err := assetFS.ReadFile(imgPath)
		if err != nil {
			fmt.Printf("Warning: Failed to load image layer %s: %v\n", imgLayer.Name, err)
			continue
		}

		img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(imgBytes))
		if err != nil {
			fmt.Printf("Warning: Failed to decode image layer %s: %v\n", imgLayer.Name, err)
			continue
		}

		// Draw the image at its offset position with opacity
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(imgLayer.OffsetX), float64(imgLayer.OffsetY))
		// Apply layer opacity from Tiled (go-tiled defaults to 1.0 if not specified)
		opacity := imgLayer.Opacity
		// Skip fully transparent layers
		if opacity <= 0 {
			img.Dispose()
			continue
		}
		op.ColorScale.ScaleAlpha(float32(opacity))
		level.Background.DrawImage(img, op)
		// Dispose temporary image to free GPU memory
		img.Dispose()
	}

	// Create a renderer that uses the embedded filesystem
	renderer, err := render.NewRendererWithFileSystem(levelMap, assetFS)
	if err != nil {
		panic(fmt.Sprintf("Failed to create renderer: %v", err))
	}

	// Render all visible tile layers
	for i, layer := range levelMap.Layers {
		// Use "render" custom property to determine visibility
		shouldRender := layer.Properties.GetBool("render")

		if shouldRender {
			if err := renderer.RenderLayer(i); err != nil {
				// Object layers can fail to render as they are not tile layers
				fmt.Printf("Warning: Failed to render layer %d: %v\n", i, err)
				continue
			}
			// Convert the rendered layer to an Ebiten image and draw it with opacity
			layerImage := ebiten.NewImageFromImage(renderer.Result)
			op := &ebiten.DrawImageOptions{}
			// Apply layer opacity from Tiled (go-tiled defaults to 1.0 if not specified)
			opacity := layer.Opacity
			// Skip fully transparent layers
			if opacity <= 0 {
				layerImage.Dispose()
				continue
			}
			op.ColorScale.ScaleAlpha(float32(opacity))
			level.Background.DrawImage(layerImage, op)
			// Dispose temporary image to free GPU memory
			layerImage.Dispose()
		}
	}

	return level
}

func LoadAssets() error {
	loader := NewLevelLoader()
	Levels := loader.MustLoadLevels()
	fmt.Println(Levels)
	// The animation assets are now embedded and loaded on demand,
	// so we no longer need to explicitly load them here.
	return nil
}

var (
	animationLoader = NewAnimationLoader()
)

func GetSheet(dir string, state config.StateID) *ebiten.Image {
	path := fmt.Sprintf("images/spritesheets/%s/%s.png", dir, state.String())
	return animationLoader.MustLoadImage(path)
}

// VFX frame dimensions for preloading
var vfxFrameSizes = map[config.StateID]struct{ W, H int }{
	config.StateJumpDust:       {96, 84},
	config.StateLandDust:       {96, 84},
	config.StateSlideDust:      {96, 84},
	config.StateExplosionShort: {57, 56},
	config.StatePlasma:         {32, 43},
	config.StateGunshot:        {46, 26},
	config.HitExplosion:        {47, 57},
	config.ChargeUp:            {102, 135},
}

// VFX directories for preloading
var vfxDirs = map[config.StateID]string{
	config.StateJumpDust:       "player",
	config.StateLandDust:       "player",
	config.StateSlideDust:      "player",
	config.StateExplosionShort: "sfx",
	config.StatePlasma:         "sfx",
	config.StateGunshot:        "sfx",
	config.HitExplosion:        "sfx",
	config.ChargeUp:            "sfx",
}

// PreloadAllAnimations preloads all sprite sheets and frames to avoid lag on first render.
// This is especially important for WASM where texture uploads are slower.
func PreloadAllAnimations() {
	// Preload player animations (96x84 frames)
	preloadCharacterAnimations("player", 96, 84)

	// Preload VFX/SFX animations with their specific frame sizes
	for state, size := range vfxFrameSizes {
		dir := vfxDirs[state]
		defs := config.CharacterAnimations[dir]
		def, ok := defs[state]
		if !ok {
			continue
		}

		// Load sprite sheet
		_ = GetSheet(dir, state)

		// Pre-cache all frames
		step := def.Step
		if step <= 0 {
			step = 1
		}
		for i := def.First; i <= def.Last; i += step {
			sx := i * size.W
			srcRect := image.Rect(sx, 0, sx+size.W, size.H)
			_ = GetFrame(dir, state, i, srcRect)
		}
	}

	// Preload obstacle animations (fire)
	for fireType, fireCfg := range config.Fire.Types {
		state := fireCfg.State
		def, ok := config.CharacterAnimations["obstacle"][state]
		if !ok {
			continue
		}

		// Load sprite sheet
		_ = GetSheet("obstacle", state)

		// Pre-cache frames
		step := def.Step
		if step <= 0 {
			step = 1
		}
		_ = fireType // Used for logging if needed
		for i := def.First; i <= def.Last; i += step {
			sx := i * fireCfg.FrameWidth
			srcRect := image.Rect(sx, 0, sx+fireCfg.FrameWidth, fireCfg.FrameHeight)
			_ = GetFrame("obstacle", state, i, srcRect)
		}
	}
}

// preloadCharacterAnimations preloads all animations for a character type
func preloadCharacterAnimations(key string, frameWidth, frameHeight int) {
	defs, ok := config.CharacterAnimations[key]
	if !ok {
		return
	}

	for state, def := range defs {
		// Load sprite sheet
		_ = GetSheet(key, state)

		// Pre-cache all frames
		step := def.Step
		if step <= 0 {
			step = 1
		}
		for i := def.First; i <= def.Last; i += step {
			sx := i * frameWidth
			srcRect := image.Rect(sx, 0, sx+frameWidth, frameHeight)
			_ = GetFrame(key, state, i, srcRect)
		}
	}
}
