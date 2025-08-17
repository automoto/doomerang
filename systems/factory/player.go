package factory

import (
	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/assets/animations"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	_ "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

const (
	playerFrameWidth      = 96
	playerFrameHeight     = 84
	playerCollisionWidth  = 16
	playerCollisionHeight = 40 // Fixed: matches actual character height
)

const (
	playerDir = "player"
)

func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

func CreatePlayer(ecs *ecs.ECS) *donburi.Entry {
	player := archetypes.Player.Spawn(ecs)

	// Calculate spawn position so that the bottom of the collision box aligns
	// with where the bottom of the full 96x84 sprite would have been previously.
	spawnX := 32.0
	spawnY := 128.0 + float64(playerFrameHeight-playerCollisionHeight)

	obj := resolv.NewObject(spawnX, spawnY, playerCollisionWidth, playerCollisionHeight)
	cfg.SetObject(player, obj)
	obj.AddTags("character")
	components.Player.SetValue(player, components.PlayerData{
		Direction:    components.Vector{X: 1, Y: 0},
		ComboCounter: 0,
	})
	components.State.SetValue(player, components.StateData{
		CurrentState: cfg.Idle,
		StateTimer:   0,
	})
	components.Physics.SetValue(player, components.PhysicsData{
		Gravity:        0.75,
		Friction:       0.5,
		AttackFriction: 0.2,
		MaxSpeed:       6.0,
	})
	components.Health.SetValue(player, components.HealthData{
		Current: 100,
		Max:     100,
	})

	obj.SetShape(resolv.NewRectangle(0, 0, playerCollisionWidth, playerCollisionHeight))

	// Load sprite sheets
	animData := GeneratePlayerAnimations()
	animData.CurrentAnimation = animData.Animations[cfg.Idle]
	components.Animation.Set(player, animData)

	return player
}

func GeneratePlayerAnimations() *components.AnimationData {
	crouchSprite := assets.GetSheet(playerDir, cfg.Crouch)
	dieSprite := assets.GetSheet(playerDir, cfg.Die)
	guardSprite := assets.GetSheet(playerDir, cfg.Guard)
	guardImpactSprite := assets.GetSheet(playerDir, cfg.GuardImpact)
	hitSprite := assets.GetSheet(playerDir, cfg.Hit)
	idleSprite := assets.GetSheet(playerDir, cfg.Idle)
	jumpSprite := assets.GetSheet(playerDir, cfg.Jump)
	kick01Sprite := assets.GetSheet(playerDir, cfg.Kick01)
	kick02Sprite := assets.GetSheet(playerDir, cfg.Kick02)
	kick03Sprite := assets.GetSheet(playerDir, cfg.Kick03)
	knockbackSprite := assets.GetSheet(playerDir, cfg.Knockback)
	ledgeSprite := assets.GetSheet(playerDir, cfg.Ledge)
	ledgeGrabSprite := assets.GetSheet(playerDir, cfg.LedgeGrab)
	punch01Sprite := assets.GetSheet(playerDir, cfg.Punch01)
	punch02Sprite := assets.GetSheet(playerDir, cfg.Punch02)
	punch03Sprite := assets.GetSheet(playerDir, cfg.Punch03)
	runningSprite := assets.GetSheet(playerDir, cfg.Running)
	stunnedSprite := assets.GetSheet(playerDir, cfg.Stunned)
	throwSprite := assets.GetSheet(playerDir, cfg.Throw)
	walkSprite := assets.GetSheet(playerDir, cfg.Walk)
	wallSlideSprite := assets.GetSheet(playerDir, cfg.WallSlide)

	// Set up animations
	animData := &components.AnimationData{
		SpriteSheets: map[string]*ebiten.Image{
			cfg.Crouch:      crouchSprite,
			cfg.Die:         dieSprite,
			cfg.Guard:       guardSprite,
			cfg.GuardImpact: guardImpactSprite,
			cfg.Hit:         hitSprite,
			cfg.Idle:        idleSprite,
			cfg.Jump:        jumpSprite,
			cfg.Kick01:      kick01Sprite,
			cfg.Kick02:      kick02Sprite,
			cfg.Kick03:      kick03Sprite,
			cfg.Knockback:   knockbackSprite,
			cfg.Ledge:       ledgeSprite,
			cfg.LedgeGrab:   ledgeGrabSprite,
			cfg.Punch01:     punch01Sprite,
			cfg.Punch02:     punch02Sprite,
			cfg.Punch03:     punch03Sprite,
			cfg.Running:     runningSprite,
			cfg.Stunned:     stunnedSprite,
			cfg.Throw:       throwSprite,
			cfg.Walk:        walkSprite,
			cfg.WallSlide:   wallSlideSprite,
		},
		CurrentSheet: cfg.Idle,
		FrameWidth:   playerFrameWidth,
		FrameHeight:  playerFrameHeight,
		Animations: map[string]*animations.Animation{
			cfg.Crouch:      animations.NewAnimation(0, 5, 1, 5),
			cfg.Die:         animations.NewAnimation(0, 8, 1, 5),
			cfg.Guard:       animations.NewAnimation(0, 0, 1, 10),
			cfg.GuardImpact: animations.NewAnimation(0, 2, 1, 5),
			cfg.Hit:         animations.NewAnimation(0, 2, 1, 5),
			cfg.Idle:        animations.NewAnimation(0, 6, 1, 5),
			cfg.Jump:        animations.NewAnimation(0, 2, 1, 10),
			cfg.Kick01:      animations.NewAnimation(0, 8, 1, 5),
			cfg.Kick02:      animations.NewAnimation(0, 7, 1, 5),
			cfg.Kick03:      animations.NewAnimation(0, 8, 1, 5),
			cfg.Knockback:   animations.NewAnimation(0, 5, 1, 5),
			cfg.Ledge:       animations.NewAnimation(0, 7, 1, 5),
			cfg.LedgeGrab:   animations.NewAnimation(0, 4, 1, 5),
			cfg.Punch01:     animations.NewAnimation(0, 5, 1, 5),
			cfg.Punch02:     animations.NewAnimation(0, 3, 1, 5),
			cfg.Punch03:     animations.NewAnimation(0, 6, 1, 5),
			cfg.Running:     animations.NewAnimation(0, 7, 1, 5),
			cfg.Stunned:     animations.NewAnimation(0, 6, 1, 5),
			cfg.Throw:       animations.NewAnimation(0, 4, 1, 5),
			cfg.Walk:        animations.NewAnimation(0, 7, 1, 5),
			cfg.WallSlide:   animations.NewAnimation(0, 5, 1, 5),
		},
	}
	return animData
}
