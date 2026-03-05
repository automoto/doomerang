package procgen

import (
	"fmt"

	"github.com/automoto/doomerang/assets"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/lafriks/go-tiled/render"
)

// Compiler converts an AssemblyResult into a playable assets.Level
type Compiler struct{}

// NewCompiler creates a new level compiler
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile converts assembled chunks into an assets.Level with world-space coordinates
func (c *Compiler) Compile(result *AssemblyResult) (*assets.Level, error) {
	if len(result.PlacedChunks) == 0 {
		return nil, fmt.Errorf("no chunks to compile")
	}

	level := &assets.Level{
		SolidTiles:  []assets.SolidTile{},
		PatrolPaths: make(map[string]assets.PatrolPath),
		EnemySpawns: []assets.EnemySpawn{},
		PlayerSpawns: []assets.PlayerSpawn{},
		DeadZones:   []assets.DeadZone{},
		Checkpoints: []assets.CheckpointSpawn{},
		Fires:       []assets.FireSpawn{},
		Messages:    []assets.MessageSpawn{},
		FinishLines: []assets.FinishLineSpawn{},
		Name:        "procgen",
		Width:       result.TotalWidth,
		Height:      result.TotalHeight,
	}

	// Create the background image
	level.Background = ebiten.NewImage(result.TotalWidth, result.TotalHeight)

	for _, pc := range result.PlacedChunks {
		ox := pc.OffsetX
		oy := pc.OffsetY

		// Merge solid tiles with world-space offsets
		for _, tile := range pc.Chunk.SolidTiles {
			level.SolidTiles = append(level.SolidTiles, assets.SolidTile{
				X:         tile.X + ox,
				Y:         tile.Y + oy,
				Width:     tile.Width,
				Height:    tile.Height,
				SlopeType: tile.SlopeType,
			})
		}

		// Render chunk tiles into the background
		if err := c.renderChunkBackground(level.Background, pc); err != nil {
			return nil, fmt.Errorf("failed to render chunk %s background: %w", pc.Chunk.ID, err)
		}

		// Process object groups for each chunk
		c.compileObjectGroups(level, pc)
	}

	// Ensure we have a player spawn
	if len(level.PlayerSpawns) == 0 {
		// Default: place at start chunk's left connection area
		first := result.PlacedChunks[0]
		level.PlayerSpawns = append(level.PlayerSpawns, assets.PlayerSpawn{
			X: first.OffsetX + 32,
			Y: first.OffsetY + float64(first.Chunk.Height) - 72, // Above floor
		})
	}

	// Ensure we have a finish line
	if len(level.FinishLines) == 0 {
		last := result.PlacedChunks[len(result.PlacedChunks)-1]
		level.FinishLines = append(level.FinishLines, assets.FinishLineSpawn{
			X:      last.OffsetX + float64(last.Chunk.Width) - 64,
			Y:      last.OffsetY + float64(last.Chunk.Height) - 80,
			Width:  32,
			Height: 48,
		})
	}

	return level, nil
}

func (c *Compiler) renderChunkBackground(bg *ebiten.Image, pc PlacedChunk) error {
	tiledMap := pc.Chunk.TiledMap
	if tiledMap == nil {
		return nil
	}

	assetFS := assets.GetAssetFS()
	renderer, err := render.NewRendererWithFileSystem(tiledMap, assetFS)
	if err != nil {
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	// Render each visible tile layer
	for i, layer := range tiledMap.Layers {
		shouldRender := layer.Properties.GetBool("render")
		if !shouldRender {
			continue
		}

		if err := renderer.RenderLayer(i); err != nil {
			continue
		}

		layerImage := ebiten.NewImageFromImage(renderer.Result)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(pc.OffsetX, pc.OffsetY)

		opacity := layer.Opacity
		if opacity <= 0 {
			layerImage.Dispose()
			continue
		}
		op.ColorScale.ScaleAlpha(float32(opacity))
		bg.DrawImage(layerImage, op)
		layerImage.Dispose()
	}

	return nil
}

func (c *Compiler) compileObjectGroups(level *assets.Level, pc PlacedChunk) {
	ox := pc.OffsetX
	oy := pc.OffsetY

	for _, og := range pc.Chunk.TiledMap.ObjectGroups {
		switch og.Name {
		case "PlayerSpawn":
			for _, o := range og.Objects {
				level.PlayerSpawns = append(level.PlayerSpawns, assets.PlayerSpawn{
					X:          o.X + ox,
					Y:          o.Y + oy,
					SpawnPoint: o.Properties.GetString("spawnPoint"),
				})
			}
		case "EnemySpawn":
			for _, o := range og.Objects {
				level.EnemySpawns = append(level.EnemySpawns, assets.EnemySpawn{
					X:          o.X + ox,
					Y:          o.Y + oy,
					EnemyType:  o.Properties.GetString("enemyType"),
					PatrolPath: o.Properties.GetString("pathName"),
				})
			}
		case "DeadZones":
			for _, o := range og.Objects {
				level.DeadZones = append(level.DeadZones, assets.DeadZone{
					X:      o.X + ox,
					Y:      o.Y + oy,
					Width:  o.Width,
					Height: o.Height,
				})
			}
		case "HazardSlots":
			for _, o := range og.Objects {
				hazardType := o.Properties.GetString("hazard_type")
				if hazardType == "deadzone" {
					level.DeadZones = append(level.DeadZones, assets.DeadZone{
						X:      o.X + ox,
						Y:      o.Y + oy,
						Width:  o.Width,
						Height: o.Height,
					})
				}
			}
		case "Checkpoint":
			for _, o := range og.Objects {
				level.Checkpoints = append(level.Checkpoints, assets.CheckpointSpawn{
					X:            o.X + ox,
					Y:            o.Y + oy,
					Width:        o.Width,
					Height:       o.Height,
					CheckpointID: o.Properties.GetFloat("checkpointID"),
				})
			}
		case "Obstacles":
			for _, o := range og.Objects {
				fireType := o.Class
				if fireType == "" {
					fireType = o.Type //nolint:staticcheck
				}
				if fireType == "fire_pulsing" || fireType == "fire_continuous" {
					direction := o.Properties.GetString("Direction")
					if direction == "" {
						direction = "right"
					}
					level.Fires = append(level.Fires, assets.FireSpawn{
						X:         o.X + ox,
						Y:         o.Y + oy,
						FireType:  fireType,
						Direction: direction,
					})
				}
			}
		case "FinishLine":
			for _, o := range og.Objects {
				level.FinishLines = append(level.FinishLines, assets.FinishLineSpawn{
					X:      o.X + ox,
					Y:      o.Y + oy,
					Width:  o.Width,
					Height: o.Height,
				})
			}
		}
	}
}
