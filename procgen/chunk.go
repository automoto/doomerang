package procgen

import (
	"embed"
	"fmt"
	"math"
	"strings"

	"github.com/automoto/doomerang/assets"
	"github.com/lafriks/go-tiled"
)

// ConnectionEdge represents which edge of the chunk a connection point is on
type ConnectionEdge string

const (
	EdgeLeft  ConnectionEdge = "left"
	EdgeRight ConnectionEdge = "right"
)

// ConnectionPoint defines an entry/exit point on a chunk's edge
type ConnectionPoint struct {
	Edge      ConnectionEdge
	SlotIndex int     // Identifies which slot on this edge (0, 1, 2...)
	YOffset   float64 // Bottom Y of connection object in pixels from chunk top
	XOffset   float64 // X position in pixels from chunk left (for top/bottom edges)
	Width     float64 // Opening width in pixels
}

// ChunkTag categorizes the type of gameplay a chunk provides
type ChunkTag string

const (
	TagCombat    ChunkTag = "combat"
	TagTraversal ChunkTag = "traversal"
	TagBreak     ChunkTag = "break"
	TagHazard    ChunkTag = "hazard"
	TagStart     ChunkTag = "start"
	TagExit      ChunkTag = "exit"
)

// EnemySlot defines a valid position for enemy placement within a chunk
type EnemySlot struct {
	X, Y          float64 // Position in chunk-local coordinates
	PlatformWidth float64 // Width of the platform this slot is on
}

// HazardSlot defines a valid position for hazard placement within a chunk
type HazardSlot struct {
	X, Y     float64 // Position in chunk-local coordinates
	SlotType string  // "fire" or "deadzone"
	Width    float64 // Width of hazard area
	Height   float64 // Height of hazard area
}

// Chunk represents a hand-authored room piece loaded from a TMX file
type Chunk struct {
	ID          string
	Biome       string
	Difficulty  int
	Tags        []ChunkTag
	MinEnemies  int
	MaxEnemies  int
	Width       int // In pixels
	Height      int // In pixels
	TileWidth   int // In tiles
	TileHeight  int // In tiles
	Connections []ConnectionPoint
	EnemySlots  []EnemySlot
	HazardSlots []HazardSlot
	SolidTiles  []assets.SolidTile
	TiledMap    *tiled.Map
	SourcePath  string
}

// HasTag returns true if the chunk has the given tag
func (c *Chunk) HasTag(tag ChunkTag) bool {
	for _, t := range c.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// GetConnections returns all connection points matching the given edge
func (c *Chunk) GetConnections(edge ConnectionEdge) []ConnectionPoint {
	var result []ConnectionPoint
	for _, cp := range c.Connections {
		if cp.Edge == edge {
			result = append(result, cp)
		}
	}
	return result
}

// ChunkLoader loads and parses chunk TMX files into Chunk structs
type ChunkLoader struct {
	fs embed.FS
}

// NewChunkLoader creates a ChunkLoader using the assets embedded filesystem
func NewChunkLoader() *ChunkLoader {
	return &ChunkLoader{fs: assets.GetAssetFS()}
}

// NewChunkLoaderWithFS creates a ChunkLoader with a custom filesystem (for testing)
func NewChunkLoaderWithFS(fs embed.FS) *ChunkLoader {
	return &ChunkLoader{fs: fs}
}

// LoadChunk loads a single chunk from a TMX file path (relative to the embedded FS root)
func (cl *ChunkLoader) LoadChunk(path string) (*Chunk, error) {
	tiledMap, err := tiled.LoadFile(path, tiled.WithFileSystem(cl.fs))
	if err != nil {
		return nil, fmt.Errorf("failed to load chunk TMX %s: %w", path, err)
	}

	chunk := &Chunk{
		TileWidth:  tiledMap.Width,
		TileHeight: tiledMap.Height,
		Width:      tiledMap.Width * tiledMap.TileWidth,
		Height:     tiledMap.Height * tiledMap.TileHeight,
		TiledMap:   tiledMap,
		SourcePath: path,
	}

	if err := parseChunkProperties(tiledMap, chunk); err != nil {
		return nil, fmt.Errorf("failed to parse chunk properties for %s: %w", path, err)
	}

	parseSolidTiles(tiledMap, chunk)
	parseObjectGroups(tiledMap, chunk)

	return chunk, nil
}

// LoadAllChunks loads all TMX files from the given directory path
func (cl *ChunkLoader) LoadAllChunks(dir string) ([]*Chunk, error) {
	entries, err := cl.fs.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read chunks directory %s: %w", dir, err)
	}

	var chunks []*Chunk
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmx") {
			continue
		}
		path := dir + "/" + entry.Name()
		chunk, err := cl.LoadChunk(path)
		if err != nil {
			return nil, fmt.Errorf("failed to load chunk %s: %w", path, err)
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func parseChunkProperties(m *tiled.Map, c *Chunk) error {
	c.ID = m.Properties.GetString("chunk_id")
	if c.ID == "" {
		return fmt.Errorf("chunk is missing required 'chunk_id' property")
	}

	c.Biome = m.Properties.GetString("biome")
	if c.Biome == "" {
		c.Biome = "default"
	}

	c.Difficulty = m.Properties.GetInt("difficulty")
	if c.Difficulty < 1 || c.Difficulty > 5 {
		c.Difficulty = 1
	}

	tagsStr := m.Properties.GetString("tags")
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			tag := ChunkTag(strings.TrimSpace(t))
			c.Tags = append(c.Tags, tag)
		}
	}

	c.MinEnemies = m.Properties.GetInt("min_enemies")
	c.MaxEnemies = m.Properties.GetInt("max_enemies")

	return nil
}

func parseSolidTiles(m *tiled.Map, c *Chunk) {
	tileW := float64(m.TileWidth)
	tileH := float64(m.TileHeight)

	for _, layer := range m.Layers {
		if layer.Name != "wg-tiles" {
			continue
		}
		for y := 0; y < m.Height; y++ {
			for x := 0; x < m.Width; x++ {
				tileIndex := y*m.Width + x
				tile := layer.Tiles[tileIndex]
				if tile.IsNil() {
					continue
				}

				var slopeType string
				if tilesetTile, err := tile.Tileset.GetTilesetTile(tile.ID); err == nil {
					slopeType = tilesetTile.Properties.GetString("slope")
				}

				c.SolidTiles = append(c.SolidTiles, assets.SolidTile{
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
}

func parseObjectGroups(m *tiled.Map, c *Chunk) {
	for _, og := range m.ObjectGroups {
		switch og.Name {
		case "Connections":
			parseConnections(og, c)
		case "EnemySlots":
			parseEnemySlots(og, c)
		case "HazardSlots":
			parseHazardSlots(og, c)
		}
	}
}

func parseConnections(og *tiled.ObjectGroup, c *Chunk) {
	for _, o := range og.Objects {
		edge := ConnectionEdge(o.Properties.GetString("edge"))
		if edge == "" {
			edge = inferEdge(o.X, float64(c.Width))
		}

		cp := ConnectionPoint{
			Edge:      edge,
			SlotIndex: o.Properties.GetInt("slot"),
			YOffset:   math.Round(o.Y + o.Height),
			XOffset:   o.X,
			Width:     o.Width,
		}
		if cp.Width == 0 {
			cp.Width = float64(c.TiledMap.TileWidth * 3)
		}
		c.Connections = append(c.Connections, cp)
	}
}

func inferEdge(x, chunkW float64) ConnectionEdge {
	const threshold = 2.0
	if x >= chunkW-threshold {
		return EdgeRight
	}
	return EdgeLeft
}

func parseEnemySlots(og *tiled.ObjectGroup, c *Chunk) {
	for _, o := range og.Objects {
		slot := EnemySlot{
			X:             o.X,
			Y:             o.Y,
			PlatformWidth: o.Width,
		}
		c.EnemySlots = append(c.EnemySlots, slot)
	}
}

func parseHazardSlots(og *tiled.ObjectGroup, c *Chunk) {
	for _, o := range og.Objects {
		slotType := o.Properties.GetString("hazard_type")
		if slotType == "" {
			slotType = "fire_pulsing"
		}
		slot := HazardSlot{
			X:        o.X,
			Y:        o.Y,
			SlotType: slotType,
			Width:    o.Width,
			Height:   o.Height,
		}
		c.HazardSlots = append(c.HazardSlots, slot)
	}
}
