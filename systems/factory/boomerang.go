package factory

import (
	"math"

	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// CreateBoomerang spawns a new boomerang entity with the given aim direction.
func CreateBoomerang(ecs *ecs.ECS, owner *donburi.Entry, chargeFrames float64, aimX, aimY float64) *donburi.Entry {
	b := archetypes.Boomerang.Spawn(ecs)

	// Get owner position and facing
	ownerObj := components.Object.Get(owner).Object
	ownerPlayer := components.Player.Get(owner)
	facingX := ownerPlayer.Direction.X

	// Determine start position (offset from player)
	startX := ownerObj.X + ownerObj.W/2
	if facingX > 0 {
		startX += 10
	} else {
		startX -= 10
	}
	startY := ownerObj.Y + ownerObj.H/2

	// Create Physics Object (Hitbox)
	// Using a smaller hitbox for the boomerang
	width, height := 12.0, 12.0
	obj := resolv.NewObject(startX, startY, width, height, tags.ResolvBoomerang)
	obj.Data = b
	components.Object.Set(b, &components.ObjectData{
		Object: obj,
	})

	// Add to space
	components.Space.Get(components.Space.MustFirst(ecs.World)).Add(obj)

	// Physics
	// Calculate initial velocity based on charge
	chargeRatio := chargeFrames / float64(config.Boomerang.MaxChargeTime)
	if chargeRatio > 1.0 {
		chargeRatio = 1.0
	}

	// Speed scaling: Base speed + bonus from charge (simple linear for now)
	speed := config.Boomerang.ThrowSpeed * (1.0 + chargeRatio*0.5)

	// Normalize aim vector and apply speed
	length := math.Sqrt(aimX*aimX + aimY*aimY)
	if length > 0 {
		aimX /= length
		aimY /= length
	}
	velocityX := speed * aimX
	velocityY := speed * aimY

	// Add upward lift to all throws for a nice arc
	velocityY -= 3.0

	components.Physics.Set(b, &components.PhysicsData{
		SpeedX:   velocityX,
		SpeedY:   velocityY,
		Gravity:  config.Boomerang.Gravity,
		Friction: 0,
		MaxSpeed: speed * 2, // Allow high speed
	})

	// Boomerang Logic
	maxRange := config.Boomerang.BaseRange + (config.Boomerang.MaxChargeRange-config.Boomerang.BaseRange)*chargeRatio

	components.Boomerang.Set(b, &components.BoomerangData{
		Owner:            owner,
		State:            components.BoomerangOutbound,
		DistanceTraveled: 0,
		MaxRange:         maxRange,
		PierceDistance:   config.Boomerang.PierceDistance,
		HitEnemies:       make(map[*donburi.Entry]struct{}),
		Damage:           config.Boomerang.BaseDamage + int(float64(config.Boomerang.MaxChargeDamageBonus)*chargeRatio),
		ChargeRatio:      chargeRatio, // Store for scaled effects
	})

	// Sprite
	img := assets.GetObjectImage("boom_green.png")
	components.Sprite.Set(b, &components.SpriteData{
		Image:    img,
		Rotation: 0,
		PivotX:   float64(img.Bounds().Dx()) / 2,
		PivotY:   float64(img.Bounds().Dy()) / 2,
	})

	// Track active boomerang on player
	if owner.HasComponent(components.Player) {
		ownerPlayer.ActiveBoomerang = b
	}

	// Spawn gunshot muzzle flash effect in the throw direction
	gunshotOffset := 45.0
	gunshotX := ownerObj.X + ownerObj.W/2 + gunshotOffset*aimX
	gunshotY := ownerObj.Y + ownerObj.H/2 + gunshotOffset*aimY
	SpawnGunshot(ecs, gunshotX, gunshotY, aimX)

	return b
}
