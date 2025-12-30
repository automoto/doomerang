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

// Global configuration instances
var C *Config
var Player PlayerConfig
var Enemy EnemyConfig
var Combat CombatConfig
var Physics PhysicsConfig
var Animation AnimationConfig
var UI UIConfig
var Boomerang BoomerangConfig

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
)

func init() {
	C = &Config{
		Width:  640,
		Height: 360,
	}
}
