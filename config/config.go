package config

import "image/color"

// PlayerConfig contains all player-related configuration values
type PlayerConfig struct {
	// Movement
	JumpSpeed    float64
	Acceleration float64
	AttackAccel  float64
	MaxSpeed     float64

	// Combat
	Health       int
	InvulnFrames int

	// Physics
	Gravity        float64
	Friction       float64
	AttackFriction float64

	// Dimensions
	FrameWidth      int
	FrameHeight     int
	CollisionWidth  int
	CollisionHeight int
}

// EnemyTypeConfig contains configuration for specific enemy types
type EnemyTypeConfig struct {
	Name             string
	Health           int
	PatrolSpeed      float64
	ChaseSpeed       float64
	AttackRange      float64
	ChaseRange       float64
	StoppingDistance float64
	AttackCooldown   int
	InvulnFrames     int
	AttackDuration   int // frames
	HitstunDuration  int // frames

	// Combat
	Damage         int
	KnockbackForce float64

	// Physics
	Gravity  float64
	Friction float64
	MaxSpeed float64

	// Dimensions
	FrameWidth      int
	FrameHeight     int
	CollisionWidth  int
	CollisionHeight int

	// Visual
	TintColor      color.RGBA // RGBA color tint for this enemy type
	SpriteSheetKey string     // e.g., "player", "guard", "slime"
}

// EnemyConfig contains enemy system configuration
type EnemyConfig struct {
	// Default enemy type configurations
	Types map[string]EnemyTypeConfig

	// AI behavior constants
	HysteresisMultiplier  float64 // For chase range hysteresis
	DefaultPatrolDistance float64 // Default patrol range when no custom path
}

// CombatConfig contains combat-related configuration values
type CombatConfig struct {
	// Player damage values
	PlayerPunchDamage    int
	PlayerKickDamage     int
	PlayerPunchKnockback float64
	PlayerKickKnockback  float64

	// Hitbox sizes
	PunchHitboxWidth  float64
	PunchHitboxHeight float64
	KickHitboxWidth   float64
	KickHitboxHeight  float64

	// Timing
	HitboxLifetime  int     // frames
	ChargeBonusRate float64 // Bonus per frame charged
	MaxChargeTime   int     // frames

	// Invulnerability
	PlayerInvulnFrames int
	EnemyInvulnFrames  int

	// Health bar display
	HealthBarDuration int // frames
}

// PhysicsConfig contains physics-related configuration values
type PhysicsConfig struct {
	// Global physics
	Gravity      float64
	MaxFallSpeed float64
	MaxRiseSpeed float64

	// Wall sliding
	WallSlideSpeed float64

	// Collision
	PlatformDropThreshold float64 // Pixels above platform to allow drop-through
	CharacterPushback     float64 // Pushback force for character collisions
	VerticalSpeedClamp    float64 // Maximum vertical speed magnitude
}

// AnimationConfig contains animation-related configuration values
type AnimationConfig struct {
	// Default animation speeds (ticks per frame)
	DefaultSpeed  int
	FastSpeed     int
	SlowSpeed     int
	VerySlowSpeed int

	// State durations (frames)
	AttackTransition  int
	HitstunDuration   int
	KnockbackDuration int
	DeathDuration     int

	// Animation frame counts (will be moved to external definitions later)
	FrameCounts map[string]int
}

// UIConfig contains UI-related configuration values
type UIConfig struct {
	// HUD dimensions
	HealthBarWidth  float64
	HealthBarHeight float64
	HealthBarMargin float64

	// Colors (RGBA)
	HealthBarBgColor [4]uint8
	HealthBarFgColor [4]uint8
	HUDTextBgColor   [4]uint8
	HUDTextColor     [4]uint8

	// Debug colors
	DebugHitboxColors map[string][4]uint8
	DebugEntityColors map[string][4]uint8

	// Font sizes
	HUDFontSize   float64
	DebugFontSize float64
}

type BoomerangConfig struct {
	ThrowSpeed     float64
	ReturnSpeed    float64
	BaseRange      float64
	MaxChargeRange float64
	PierceDistance float64
	Gravity        float64
	MaxChargeTime  int
}

// Config holds general game configuration
type Config struct {
	Width  int
	Height int
}

// DebugConfig contains debug visualization settings
type DebugConfig struct {
	GridEmpty       color.RGBA
	GridOccupied    color.RGBA
	CollisionRect   color.RGBA
	EntityDefault   color.RGBA
	EntitySolid     color.RGBA
	EntityPlayer    color.RGBA
	EntityEnemy     color.RGBA
	EntityBoomerang color.RGBA
}

// Global configuration instances
var C *Config
var Player PlayerConfig
var Enemy EnemyConfig
var Combat CombatConfig
var Physics PhysicsConfig
var Animation AnimationConfig
var UI UIConfig
var Boomerang BoomerangConfig
var Debug DebugConfig

// Shared RGBA color constants
var (
	White       = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	Yellow      = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	Orange      = color.RGBA{R: 255, G: 165, B: 0, A: 255}
	Red         = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	Green       = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	BrightGreen = color.RGBA{R: 0, G: 255, B: 60, A: 255}
	Blue        = color.RGBA{R: 0, G: 100, B: 255, A: 255}
	Purple      = color.RGBA{R: 128, G: 0, B: 255, A: 255}
	LightRed    = color.RGBA{R: 255, G: 60, B: 60, A: 255}
	Magenta     = color.RGBA{R: 255, G: 0, B: 255, A: 255}

	// Debug colors (with alpha for transparency)
	DarkGray         = color.RGBA{R: 20, G: 20, B: 20, A: 255}
	GrayTransparent  = color.RGBA{R: 60, G: 60, B: 60, A: 128}
	Gray             = color.RGBA{R: 100, G: 100, B: 100, A: 100}
	Cyan             = color.RGBA{R: 0, G: 255, B: 255, A: 100}
	PureBlue         = color.RGBA{R: 0, G: 0, B: 255, A: 100}
	RedTransparent   = color.RGBA{R: 255, G: 0, B: 0, A: 100}
	GreenTransparent = color.RGBA{R: 0, G: 255, B: 0, A: 200}
)

// Direction constants for player facing
const (
	DirectionLeft  = -1.0
	DirectionRight = 1.0
)

func init() {
	C = &Config{
		Width:  640,
		Height: 360,
	}

	// Physics Config
	Physics = PhysicsConfig{
		// Global physics
		Gravity:      0.75,
		MaxFallSpeed: 10.0,
		MaxRiseSpeed: -10.0,

		// Wall sliding
		WallSlideSpeed: 1.0,

		// Collision
		PlatformDropThreshold: 4.0,  // Pixels above platform to allow drop-through
		CharacterPushback:     2.0,  // Pushback force for character collisions
		VerticalSpeedClamp:    10.0, // Maximum vertical speed magnitude
	}

	// Player Config
	Player = PlayerConfig{
		// Movement
		JumpSpeed:    15.0,
		Acceleration: 0.75,
		AttackAccel:  0.1,
		MaxSpeed:     6.0,

		// Combat
		Health:       100,
		InvulnFrames: 30,

		// Physics
		Gravity:        0.75,
		Friction:       0.5,
		AttackFriction: 0.2,

		// Dimensions
		FrameWidth:      96,
		FrameHeight:     84,
		CollisionWidth:  16,
		CollisionHeight: 40,
	}

	// Boomerang Config
	Boomerang = BoomerangConfig{
		ThrowSpeed:     6.0,
		ReturnSpeed:    8.0,
		BaseRange:      150.0,
		MaxChargeRange: 250.0,
		PierceDistance: 40.0,
		Gravity:        0.2,
		MaxChargeTime:  60,
	}

	// Enemy Config
	guardType := EnemyTypeConfig{
		Name:             "Guard",
		Health:           60,
		PatrolSpeed:      2.0,
		ChaseSpeed:       2.5,
		AttackRange:      36.0,
		ChaseRange:       80.0,
		StoppingDistance: 28.0,
		AttackCooldown:   60,
		InvulnFrames:     15,
		AttackDuration:   30,
		HitstunDuration:  15,
		Damage:           10,
		KnockbackForce:   5.0,
		Gravity:          0.75,
		Friction:         0.2,
		MaxSpeed:         6.0,
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   16,
		CollisionHeight:  40,
		TintColor:        White,
		SpriteSheetKey:   "player",
	}

	lightGuardType := EnemyTypeConfig{
		Name:             "LightGuard",
		Health:           40,
		PatrolSpeed:      3.0,
		ChaseSpeed:       3.5,
		AttackRange:      32.0,
		ChaseRange:       100.0,
		StoppingDistance: 24.0,
		AttackCooldown:   40,
		InvulnFrames:     10,
		AttackDuration:   20,
		HitstunDuration:  10,
		Damage:           8,
		KnockbackForce:   3.0,
		Gravity:          0.8,
		Friction:         0.25,
		MaxSpeed:         7.0,
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   14,
		CollisionHeight:  36,
		TintColor:        Yellow,
		SpriteSheetKey:   "player",
	}

	heavyGuardType := EnemyTypeConfig{
		Name:             "HeavyGuard",
		Health:           100,
		PatrolSpeed:      1.5,
		ChaseSpeed:       2.0,
		AttackRange:      40.0,
		ChaseRange:       60.0,
		StoppingDistance: 32.0,
		AttackCooldown:   90,
		InvulnFrames:     25,
		AttackDuration:   45,
		HitstunDuration:  25,
		Damage:           18,
		KnockbackForce:   8.0,
		Gravity:          0.7,
		Friction:         0.15,
		MaxSpeed:         4.0,
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   20,
		CollisionHeight:  44,
		TintColor:        Orange,
		SpriteSheetKey:   "player",
	}

	Enemy = EnemyConfig{
		Types: map[string]EnemyTypeConfig{
			"Guard":      guardType,
			"LightGuard": lightGuardType,
			"HeavyGuard": heavyGuardType,
		},
		HysteresisMultiplier:  1.5,
		DefaultPatrolDistance: 32.0,
	}

	// Combat Config (Populated with default values matching the previous constants)
	Combat = CombatConfig{
		PlayerPunchDamage:    15,
		PlayerKickDamage:     22,
		PlayerPunchKnockback: 3.0,
		PlayerKickKnockback:  5.0,

		PunchHitboxWidth:  20,
		PunchHitboxHeight: 16,
		KickHitboxWidth:   28,
		KickHitboxHeight:  20,

		HitboxLifetime:  10,
		ChargeBonusRate: 0, // Calculated dynamically in code, but good to have here
		MaxChargeTime:   60,

		PlayerInvulnFrames: 30,
		EnemyInvulnFrames:  30, // Was hardcoded to 15 in some places, 30 in others

		HealthBarDuration: 180,
	}

	// Debug Config
	Debug = DebugConfig{
		GridEmpty:       DarkGray,
		GridOccupied:    Yellow,
		CollisionRect:   GrayTransparent,
		EntityDefault:   Cyan,
		EntitySolid:     Gray,
		EntityPlayer:    PureBlue,
		EntityEnemy:     RedTransparent,
		EntityBoomerang: GreenTransparent,
	}
}
