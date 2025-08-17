package assets

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/lafriks/go-tiled"
	"github.com/lafriks/go-tiled/render"
	"github.com/yohamta/donburi/features/math"
)

var (
	//go:embed all:levels
	assetFS embed.FS
	//go:embed fonts/excel.ttf
	excelFontData []byte

	//go:embed all:images/spritesheets
	animationFS embed.FS
)

type Level struct {
	Background *ebiten.Image
	Paths      map[uint32]Path
	Name       string
	Width      int
	Height     int
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

type AnimationLoader struct {
	// No fields needed for now, as animationFS is a package-level variable
}

func NewAnimationLoader() *AnimationLoader {
	return &AnimationLoader{}
}

func (l *AnimationLoader) MustLoadImage(path string) *ebiten.Image {
	imgBytes, err := animationFS.ReadFile(path)
	if err != nil {
		panic(fmt.Sprintf("Failed to read image file %s: %v", path, err))
	}

	img, _, err := ebitenutil.NewImageFromReader(bytes.NewReader(imgBytes))
	if err != nil {
		panic(fmt.Sprintf("Failed to create image from bytes for %s: %v", path, err))
	}
	return img
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
		Paths:  make(map[uint32]Path),
		Name:   levelPath,
		Width:  levelMap.Width * levelMap.TileWidth,
		Height: levelMap.Height * levelMap.TileHeight,
	}

	// Load ground objects from the ground-walls object group
	for _, og := range levelMap.ObjectGroups {
		if og.Name == "ground-walls" {
			for _, o := range og.Objects {
				// Create a path with two points for each ground object
				level.Paths[o.ID] = Path{
					Loops: false,
					Points: []math.Vec2{
						{X: o.X, Y: o.Y},
						{X: o.X + o.Width, Y: o.Y + o.Height},
					},
				}
			}
		}
	}

	// Create a new image for the background
	level.Background = ebiten.NewImage(levelMap.Width*levelMap.TileWidth, levelMap.Height*levelMap.TileHeight)

	// Create a renderer that uses the embedded filesystem
	renderer, err := render.NewRendererWithFileSystem(levelMap, assetFS)
	if err != nil {
		panic(fmt.Sprintf("Failed to create renderer: %v", err))
	}

	// Render all visible layers
	for i, layer := range levelMap.Layers {
		if layer.Visible {
			if err := renderer.RenderLayer(i); err != nil {
				panic(fmt.Sprintf("Failed to render layer %d: %v", i, err))
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

func GetSheet(dir string, state string) *ebiten.Image {
	path := fmt.Sprintf("images/spritesheets/%s/%s.png", dir, state)
	return animationLoader.MustLoadImage(path)
}
