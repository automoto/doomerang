package assets

import (
	"bytes"
	"embed"
	"fmt"
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
	Name         string
	Width        int
	Height       int
}

// SolidTile represents a solid collision tile
type SolidTile struct {
	X, Y, Width, Height float64
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

type LevelLoader struct {
	Tilesets map[string]*tiled.Tileset
}

func NewLevelLoader() *LevelLoader {
	return &LevelLoader{
		Tilesets: make(map[string]*tiled.Tileset),
	}
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
	cache map[string]*ebiten.Image
}

func NewAnimationLoader() *AnimationLoader {
	return &AnimationLoader{
		cache: make(map[string]*ebiten.Image),
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

func GetObjectImage(name string) *ebiten.Image {
	path := fmt.Sprintf("images/objects/%s", name)
	return animationLoader.MustLoadImage(path)
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
				level.SolidTiles = append(level.SolidTiles, SolidTile{
					X:      float64(x) * tileW,
					Y:      float64(y) * tileH,
					Width:  tileW,
					Height: tileH,
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

		// Draw the image at its offset position
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(float64(imgLayer.OffsetX), float64(imgLayer.OffsetY))
		level.Background.DrawImage(img, op)
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
			// Convert the rendered layer to an Ebiten image and draw it
			layerImage := ebiten.NewImageFromImage(renderer.Result)
			op := &ebiten.DrawImageOptions{}
			level.Background.DrawImage(layerImage, op)
		}
	}

	// Cache tilesets for future use
	for _, ts := range levelMap.Tilesets {
		if _, ok := l.Tilesets[ts.Class]; !ok {
			l.Tilesets[ts.Class] = ts
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
