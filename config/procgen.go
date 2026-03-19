package config

// ProcgenConfig contains procedural generation configuration values
type ProcgenConfig struct {
	// Tile dimensions (must match tileset)
	TileWidth  int
	TileHeight int

	// Standard chunk sizes (in tiles)
	ChunkSmallWidth  int // 20 tiles
	ChunkMediumWidth int // 40 tiles
	ChunkLargeWidth  int // 60 tiles
	ChunkMinHeight   int // 10 tiles
	ChunkMaxHeight   int // 30 tiles

	// Enemy budget
	EnemyBudgetBase       float64 // Base budget for enemy placement
	EnemyBudgetMultiplier float64 // Budget multiplier per difficulty level
	EnemyMinSpacing       float64 // Minimum pixels between enemies

	// Enemy point costs
	EnemyCosts map[string]int

	// Difficulty
	MinDifficulty int
	MaxDifficulty int

	// Run length
	DefaultRunLength int // Number of concept graph nodes
	MinRunLength     int
	MaxRunLength     int

	// Connection points
	StandardConnectionHeight int // Standard Y-offset for connection points (in tiles)
	ConnectionOpeningWidth   int // Standard opening width (in tiles)

	// Level layout
	ChunkHeadroomFactor float64 // Fraction of screen height added above chunks for camera headroom

	// Biomes available for graph generation
	Biomes []string
}

// Procgen is the global procgen configuration instance
var Procgen ProcgenConfig

func init() {
	Procgen = ProcgenConfig{
		TileWidth:  16,
		TileHeight: 16,

		ChunkSmallWidth:  20,
		ChunkMediumWidth: 40,
		ChunkLargeWidth:  60,
		ChunkMinHeight:   10,
		ChunkMaxHeight:   30,

		EnemyBudgetBase:       4.0,
		EnemyBudgetMultiplier: 3.0,
		EnemyMinSpacing:       64.0,

		EnemyCosts: map[string]int{
			"LightGuard":   2,
			"Guard":        3,
			"KnifeThrower": 4,
			"HeavyGuard":   5,
		},

		MinDifficulty: 1,
		MaxDifficulty: 5,

		DefaultRunLength: 10,
		MinRunLength:     8,
		MaxRunLength:     25,

		StandardConnectionHeight: 20,  // tiles from top
		ConnectionOpeningWidth:   3,   // tiles wide
		ChunkHeadroomFactor:      0.15, // 15% of screen height

		Biomes: []string{"cyberpunk", "industrial", "neon"},
	}
}
