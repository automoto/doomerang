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

	// Lives
	StartingLives       int
	RespawnInvulnFrames int

	// Physics
	Gravity        float64
	Friction       float64
	AttackFriction float64

	// Slide mechanics
	SlideSpeedThreshold float64 // Minimum speed required to initiate slide
	SlideFriction       float64 // Friction applied during slide
	SlideHitboxHeight   float64 // Reduced hitbox height during slide
	SlideMinSpeed       float64 // Speed below which slide ends
	SlideRecoveryFrames int     // Delay before standing up after slide
	SlideRotation       float64 // Radians to rotate sprite during slide

	// Crouch mechanics
	CrouchWalkSpeed float64 // Max speed while crouch-walking

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
	MaxVerticalChase float64 // Max vertical distance to chase/attack player (melee only)
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

	// Ranged combat (for knife thrower type)
	IsRanged           bool    // If true, enemy throws projectiles instead of melee
	ThrowRange         float64 // Distance at which enemy can throw
	ThrowCooldown      int     // Frames between throws
	ThrowWindupTime    int     // Frames before knife is released (for animation sync)
	MinVerticalToThrow float64 // Min vertical distance below which to walk to edge instead of direct throw
	EdgeApproachSpeed  float64 // Speed when walking to platform edge
	EdgeThrowDistance  float64 // Max horizontal distance to throw from edge
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

	// Knockback
	KnockbackUpwardForce float64 // Upward velocity applied on knockback

	// Jump kick
	JumpKickRotation float64 // Radians to rotate sprite during jump kick

	// Health bar display
	HealthBarDuration int // frames

	// Flash effects (frames)
	HitFlashFrames    int // white flash when dealing damage
	DamageFlashFrames int // red flash when taking damage
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
	ThrowSpeed           float64
	ReturnSpeed          float64
	BaseRange            float64
	MaxChargeRange       float64
	PierceDistance       float64
	Gravity              float64
	MaxChargeTime        int
	HitKnockback         float64 // horizontal knockback applied to enemies on hit
	BaseDamage           int     // minimum damage at no charge
	MaxChargeDamageBonus int     // additional damage at full charge
}

// KnifeConfig contains knife projectile configuration
type KnifeConfig struct {
	Speed            float64 // Projectile speed (faster than player)
	Damage           int     // Damage dealt to player
	Width            float64 // Collision width
	Height           float64 // Collision height
	KnockbackForce   float64 // Knockback on hit
	MaxDownwardAngle float64 // Maximum downward angle in radians (0 = horizontal, positive = downward)
}

// FireHitboxPhase defines hitbox scaling for a range of animation frames
type FireHitboxPhase struct {
	StartFrame int
	EndFrame   int
	StartScale float64 // 0.0 to 1.0
	EndScale   float64 // 0.0 to 1.0
}

// FireTypeConfig contains configuration for specific fire obstacle types
type FireTypeConfig struct {
	Damage         int
	KnockbackForce float64
	FrameWidth     int
	FrameHeight    int
	State          StateID
	HitboxPhases   []FireHitboxPhase // nil = static hitbox (for fire_continuous)
	HitboxScale    float64           // Scale factor for hitbox (1.0 = sprite size)
}

// FireConfig contains fire obstacle configuration
type FireConfig struct {
	Types map[string]FireTypeConfig
}

// PauseConfig contains pause menu configuration values
type PauseConfig struct {
	OverlayColor      color.RGBA
	TextColorNormal   color.RGBA
	TextColorSelected color.RGBA
	MenuItemHeight    float64
	MenuItemGap       float64
	MenuOptions       []string
}

// MenuConfig contains main menu configuration values
type MenuConfig struct {
	BackgroundColor      color.RGBA
	TitleColor           color.RGBA
	TextColorNormal      color.RGBA
	TextColorSelected    color.RGBA
	TitleY               float64
	MenuStartY           float64
	MenuItemHeight       float64
	MenuItemGap          float64
	MenuOptions          []string
	ConfirmDialogMessage string
	ConfirmDialogYes     string
	ConfirmDialogNo      string
}

// GameOverConfig contains game over screen configuration values
type GameOverConfig struct {
	BackgroundColor   color.RGBA
	TitleColor        color.RGBA
	TextColorNormal   color.RGBA
	TextColorSelected color.RGBA
	TitleY            float64
	MenuStartY        float64
	MenuItemHeight    float64
	MenuItemGap       float64
	MenuOptions       []string
}

// ScreenShakeConfig contains screen shake effect configuration
type ScreenShakeConfig struct {
	MeleeIntensity        float64 // pixels - punch, kick, jump kick (all same)
	MeleeDuration         int     // frames
	PlayerDamageIntensity float64 // pixels
	PlayerDamageDuration  int     // frames
	BoomerangIntensity    float64 // pixels - charged throw impact
	BoomerangDuration     int     // frames
}

// CameraConfig contains camera behavior configuration
type CameraConfig struct {
	FollowSmoothing         float64 // How fast camera follows player (0.0-1.0)
	LookAheadDistanceX      float64 // Max horizontal look-ahead offset in pixels
	LookAheadSmoothing      float64 // How fast look-ahead offset changes (0.0-1.0)
	LookAheadMovingScale    float64 // Scale when player is moving (1.0)
	LookAheadSpeedThreshold float64 // Minimum speed to update look-ahead
}

// DeathZoneConfig contains death zone effect configuration
type DeathZoneConfig struct {
	RespawnDelayFrames   int     // Frames before respawn (~0.75s at 60fps)
	ScreenShakeIntensity float64 // Screen shake intensity in pixels
	ScreenShakeDuration  int     // Screen shake duration in frames
	ExplosionScale       float64 // Scale of explosion VFX
}

// SquashStretchConfig contains squash/stretch effect configuration
type SquashStretchConfig struct {
	JumpScaleX float64 // horizontal scale on jump (< 1 = narrower)
	JumpScaleY float64 // vertical scale on jump (> 1 = taller)
	LandScaleX float64 // horizontal scale on land (> 1 = wider)
	LandScaleY float64 // vertical scale on land (< 1 = shorter)
	LerpSpeed  float64 // how fast to return to normal scale
}

// LevelCompleteConfig contains level complete overlay configuration
type LevelCompleteConfig struct {
	OverlayColor color.RGBA
	TitleColor   color.RGBA
	TextColor    color.RGBA
	HintColor    color.RGBA
	TitleY       float64
	MessageY     float64
	HintY        float64
	Title        string
	Message      string
	ContinueHint string
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
var Knife KnifeConfig
var Fire FireConfig
var Pause PauseConfig
var Menu MenuConfig
var GameOver GameOverConfig
var ScreenShake ScreenShakeConfig
var SquashStretch SquashStretchConfig
var DeathZone DeathZoneConfig
var Debug DebugConfig
var Message MessageConfig
var LevelComplete LevelCompleteConfig
var Camera CameraConfig

// DebugConfig contains debug/testing command-line options
type DebugConfig struct {
	SkipMenu        bool    // Skip menu and go directly to game
	StartCheckpoint float64 // Checkpoint ID to spawn at (-1 = use default)
}

// MessageConfig contains message popup configuration
type MessageConfig struct {
	ActivationRadius float64    // Pixels to trigger message display
	DisplayDuration  int        // Frames to display message after triggering
	BoxPadding       float64    // Padding inside message box
	BoxColor         color.RGBA // Semi-transparent background color
	TextColor        color.RGBA // Text color
	TopMargin        float64    // Distance from top of screen

	// Message content by ID (float64 to match Tiled property format)
	Messages map[float64]string

	// Input labels for different input methods
	KeyboardLabels    map[string]string
	XboxLabels        map[string]string
	PlayStationLabels map[string]string
}

// Shared RGBA color constants
var (
	White        = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	Yellow       = color.RGBA{R: 255, G: 255, B: 0, A: 255}
	BrightYellow = color.RGBA{R: 255, G: 255, B: 100, A: 255}
	Orange       = color.RGBA{R: 255, G: 140, B: 0, A: 255}
	BrightOrange = color.RGBA{R: 255, G: 180, B: 50, A: 255}
	Red          = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	Green        = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	BrightGreen  = color.RGBA{R: 0, G: 255, B: 60, A: 255}
	LightGreen   = color.RGBA{R: 100, G: 255, B: 100, A: 255}
	Blue         = color.RGBA{R: 0, G: 100, B: 255, A: 255}
	Purple       = color.RGBA{R: 128, G: 0, B: 255, A: 255}
	LightRed     = color.RGBA{R: 255, G: 60, B: 60, A: 255}
	Magenta      = color.RGBA{R: 255, G: 0, B: 255, A: 255}
	BlackOverlay = color.RGBA{R: 0, G: 0, B: 0, A: 180}
	LightBlue    = color.RGBA{R: 100, G: 180, B: 255, A: 255} // Selected menu items
	DarkBlue     = color.RGBA{R: 60, G: 100, B: 160, A: 255}  // Unselected menu items
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
		Health:       60,
		InvulnFrames: 30,

		// Lives
		StartingLives:       3,
		RespawnInvulnFrames: 60,

		// Physics
		Gravity:        0.75,
		Friction:       0.5,
		AttackFriction: 0.2,

		// Slide mechanics
		SlideSpeedThreshold: 4.0,   // Must be moving at least this fast to slide
		SlideFriction:       0.08,  // Gradual slowdown during slide (lower = longer slide)
		SlideHitboxHeight:   20.0,  // Reduced height during slide (normal is 40)
		SlideMinSpeed:       0.3,   // Stop sliding when speed drops below this
		SlideRecoveryFrames: 8,     // Frames of delay before standing up
		SlideRotation:       -0.35, // Rotate sprite for ground slide look

		// Crouch mechanics
		CrouchWalkSpeed: 1.5, // Slow movement while crouched

		// Dimensions
		FrameWidth:      96,
		FrameHeight:     84,
		CollisionWidth:  16,
		CollisionHeight: 40,
	}

	// Boomerang Config
	Boomerang = BoomerangConfig{
		ThrowSpeed:           6.0,
		ReturnSpeed:          8.0,
		BaseRange:            150.0,
		MaxChargeRange:       250.0,
		PierceDistance:       40.0,
		Gravity:              0.2,
		MaxChargeTime:        60,
		HitKnockback:         2.0,
		BaseDamage:           15,
		MaxChargeDamageBonus: 15,
	}

	// Knife Config
	Knife = KnifeConfig{
		Speed:            8.0, // Faster than player (6.0)
		Damage:           30,
		Width:            36.0, // Actual knife dimensions within 96x84 sprite
		Height:           7.0,
		KnockbackForce:   4.0,
		MaxDownwardAngle: 0.5, // ~30 degrees max downward angle to avoid hitting ground
	}

	// Fire Config
	Fire = FireConfig{
		Types: map[string]FireTypeConfig{
			"fire_pulsing": {
				Damage:         15,
				KnockbackForce: 6.0,
				FrameWidth:     65,
				FrameHeight:    43,
				State:          FirePulsing,
				HitboxScale:    0.85,
				HitboxPhases: []FireHitboxPhase{
					{StartFrame: 0, EndFrame: 10, StartScale: 0.3, EndScale: 0.6},  // Igniting
					{StartFrame: 11, EndFrame: 23, StartScale: 0.6, EndScale: 1.0}, // Growing
					{StartFrame: 24, EndFrame: 34, StartScale: 1.0, EndScale: 0.0}, // Dying
					// Frames 35-44: no entry = no hitbox
				},
			},
			"fire_continuous": {
				Damage:         15,
				KnockbackForce: 6.0,
				FrameWidth:     66,
				FrameHeight:    47,
				State:          FireContinuous,
				HitboxScale:    0.85,
			},
		},
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
		MaxVerticalChase: 144.0,
		AttackCooldown:   60,
		InvulnFrames:     15,
		AttackDuration:   30,
		HitstunDuration:  15,
		Damage:           40,
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
		Health:           30,
		PatrolSpeed:      3.0,
		ChaseSpeed:       3.5,
		AttackRange:      32.0,
		ChaseRange:       100.0,
		StoppingDistance: 24.0,
		MaxVerticalChase: 144.0,
		AttackCooldown:   40,
		InvulnFrames:     10,
		AttackDuration:   20,
		HitstunDuration:  10,
		Damage:           20,
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
		MaxVerticalChase: 144.0,
		AttackCooldown:   90,
		InvulnFrames:     25,
		AttackDuration:   45,
		HitstunDuration:  25,
		Damage:           30,
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

	knifeThrowerType := EnemyTypeConfig{
		Name:             "KnifeThrower",
		Health:           30,
		PatrolSpeed:      1.5, // Patrol when player not in range
		ChaseSpeed:       0,   // Does not chase
		AttackRange:      0,   // Not used for ranged
		ChaseRange:       0,   // Not used
		StoppingDistance: 0,   // Not used
		AttackCooldown:   0,   // Not used (use ThrowCooldown instead)
		InvulnFrames:     15,
		AttackDuration:   0, // Not used
		HitstunDuration:  20,
		Damage:           0, // Not used (knife has its own damage)
		KnockbackForce:   0, // Not used
		Gravity:          0.75,
		Friction:         0.2,
		MaxSpeed:         3.0, // Allow movement for patrol
		FrameWidth:       96,
		FrameHeight:      84,
		CollisionWidth:   16,
		CollisionHeight:  40,
		TintColor:        Purple,
		SpriteSheetKey:   "player",
		// Ranged specific
		IsRanged:           true,
		ThrowRange:         300.0, // Detection/attack range
		ThrowCooldown:      120,   // 2 seconds at 60fps
		ThrowWindupTime:    15,    // Frames before knife spawns
		MinVerticalToThrow: 32.0,  // If player is more than 32px below, walk to edge
		EdgeApproachSpeed:  1.5,   // Speed when approaching platform edge
		EdgeThrowDistance:  200.0, // Max horizontal distance to throw from edge
	}

	Enemy = EnemyConfig{
		Types: map[string]EnemyTypeConfig{
			"Guard":        guardType,
			"LightGuard":   lightGuardType,
			"HeavyGuard":   heavyGuardType,
			"KnifeThrower": knifeThrowerType,
		},
		HysteresisMultiplier:  1.5,
		DefaultPatrolDistance: 32.0,
	}

	// Combat Config (Populated with default values matching the previous constants)
	Combat = CombatConfig{
		// Normalized: punch and kick now have identical values
		PlayerPunchDamage:    22,
		PlayerKickDamage:     22,
		PlayerPunchKnockback: 5.0,
		PlayerKickKnockback:  5.0,

		// Normalized: punch and kick hitboxes are now the same size
		PunchHitboxWidth:  28,
		PunchHitboxHeight: 20,
		KickHitboxWidth:   28,
		KickHitboxHeight:  20,

		HitboxLifetime:  10,
		ChargeBonusRate: 0, // Calculated dynamically in code, but good to have here
		MaxChargeTime:   60,

		// Extended from 30 to prevent stun-lock
		PlayerInvulnFrames: 45,
		EnemyInvulnFrames:  30,

		// Knockback tuning - increased for better ledge knockback
		KnockbackUpwardForce: -4.0,

		// Jump kick rotation in radians (~20 degrees)
		JumpKickRotation: 0.35,

		HealthBarDuration: 180,

		// Flash effects
		HitFlashFrames:    3,
		DamageFlashFrames: 5,
	}

	// Pause Config
	Pause = PauseConfig{
		OverlayColor:      BlackOverlay,
		TextColorNormal:   White,
		TextColorSelected: BrightOrange,
		MenuItemHeight:    30,
		MenuItemGap:       15,
		MenuOptions:       []string{"Resume", "Settings", "Exit"},
	}

	// Menu Config
	Menu = MenuConfig{
		BackgroundColor:      color.RGBA{R: 15, G: 25, B: 50, A: 255},
		TitleColor:           Orange,
		TextColorNormal:      White,
		TextColorSelected:    BrightOrange,
		TitleY:               50,
		MenuStartY:           100,
		MenuItemHeight:       30,
		MenuItemGap:          12,
		MenuOptions:          []string{"Start", "Continue", "Settings", "Exit"},
		ConfirmDialogMessage: "Overwrite existing save?",
		ConfirmDialogYes:     "Yes",
		ConfirmDialogNo:      "No",
	}

	// Game Over Config
	GameOver = GameOverConfig{
		BackgroundColor:   color.RGBA{R: 40, G: 10, B: 10, A: 255},
		TitleColor:        LightRed,
		TextColorNormal:   White,
		TextColorSelected: BrightOrange,
		TitleY:            100,
		MenuStartY:        160,
		MenuItemHeight:    30,
		MenuItemGap:       15,
		MenuOptions:       []string{"Retry", "Main Menu"},
	}

	// Screen Shake Config
	ScreenShake = ScreenShakeConfig{
		MeleeIntensity:        2.0,
		MeleeDuration:         5,
		PlayerDamageIntensity: 4.0,
		PlayerDamageDuration:  8,
		BoomerangIntensity:    4.0,
		BoomerangDuration:     6,
	}

	// Squash/Stretch Config
	SquashStretch = SquashStretchConfig{
		JumpScaleX: 0.7,
		JumpScaleY: 1.5,
		LandScaleX: 1.5,
		LandScaleY: 0.6,
		LerpSpeed:  0.10,
	}

	// Death Zone Config
	DeathZone = DeathZoneConfig{
		RespawnDelayFrames:   15,  // ~0.25s at 60fps
		ScreenShakeIntensity: 8.0, // Strong shake for death
		ScreenShakeDuration:  12,
		ExplosionScale:       1.2,
	}

	// Debug Config (defaults, can be overridden by CLI flags)
	Debug = DebugConfig{
		SkipMenu:        false,
		StartCheckpoint: -1,
	}

	// Message Config
	Message = MessageConfig{
		ActivationRadius: 100.0, // Large radius to catch jumping players
		DisplayDuration:  300,   // 5 seconds at 60fps
		BoxPadding:       8.0,
		BoxColor:         color.RGBA{R: 0, G: 0, B: 0, A: 200},
		TextColor:        color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TopMargin:        30.0,

		Messages: map[float64]string{
			1.0: "{move} to move, {jump} to jump",
			1.1: "{attack} to strike",
			1.2: "Hold toward wall + {jump} to wall slide and jump",
			1.3: "{boomerang} to throw boomerang - hold to charge for more damage",
			1.4: "Run + {down} to slide",
			1.5: "Hold a direction while throwing to aim",
		},

		KeyboardLabels: map[string]string{
			"jump": "X", "attack": "Z", "boomerang": "SPACE",
			"move": "Arrow Keys", "up": "UP", "down": "DOWN",
		},
		XboxLabels: map[string]string{
			"jump": "A", "attack": "X", "boomerang": "B",
			"move": "Left Stick", "up": "D-Pad Up", "down": "D-Pad Down",
		},
		PlayStationLabels: map[string]string{
			"jump": "Cross", "attack": "Square", "boomerang": "Circle",
			"move": "Left Stick", "up": "D-Pad Up", "down": "D-Pad Down",
		},
	}

	// Level Complete Config
	LevelComplete = LevelCompleteConfig{
		OverlayColor: BlackOverlay,
		TitleColor:   BrightGreen,
		TextColor:    White,
		HintColor:    White,
		TitleY:       80,
		MessageY:     140,
		HintY:        280,
		Title:        "Level Complete!",
		Message:      "Thanks for playing Level 1! More levels coming, stay tuned...",
		ContinueHint: "Press ENTER to exit",
	}

	// Camera Config
	Camera = CameraConfig{
		FollowSmoothing:         0.1,  // Current hardcoded value
		LookAheadDistanceX:      60.0, // ~10% of 640px screen width
		LookAheadSmoothing:      0.05, // Slower than follow for smooth feel
		LookAheadMovingScale:    1.0,
		LookAheadSpeedThreshold: 0.1, // Minimum speed to update look-ahead
	}
}
