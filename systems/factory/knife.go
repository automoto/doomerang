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

// CreateKnife spawns a knife projectile aimed at the target position.
func CreateKnife(ecs *ecs.ECS, owner *donburi.Entry, targetX, targetY float64) *donburi.Entry {
	k := archetypes.Knife.Spawn(ecs)

	// Get owner position
	ownerObj := components.Object.Get(owner).Object

	// Start position (center of enemy)
	startX := ownerObj.X + ownerObj.W/2
	startY := ownerObj.Y + ownerObj.H/2

	// Create collision object
	obj := resolv.NewObject(
		startX-config.Knife.Width/2,
		startY-config.Knife.Height/2,
		config.Knife.Width,
		config.Knife.Height,
		tags.ResolvKnife,
	)
	obj.Data = k
	components.Object.Set(k, &components.ObjectData{Object: obj})

	// Add to space
	components.Space.Get(components.Space.MustFirst(ecs.World)).Add(obj)

	dx := targetX - startX
	dy := targetY - startY

	// Clamp downward angle to prevent throwing into ground
	if dy > 0 && config.Knife.MaxDownwardAngle > 0 {
		angle := math.Atan2(dy, math.Abs(dx))
		if angle > config.Knife.MaxDownwardAngle {
			dy = math.Abs(dx) * math.Tan(config.Knife.MaxDownwardAngle)
		}
	}

	length := math.Sqrt(dx*dx + dy*dy)
	if length > 0 {
		dx /= length
		dy /= length
	}

	velocityX := config.Knife.Speed * dx
	velocityY := config.Knife.Speed * dy

	components.Physics.Set(k, &components.PhysicsData{
		SpeedX:   velocityX,
		SpeedY:   velocityY,
		Gravity:  0, // Knife travels in straight line
		Friction: 0,
		MaxSpeed: config.Knife.Speed * 2,
	})

	// Knife data
	components.Knife.Set(k, &components.KnifeData{
		Owner:  owner,
		Damage: config.Knife.Damage,
		Speed:  config.Knife.Speed,
	})

	// Sprite
	img := assets.GetObjectImage("knife-green.png")

	// Calculate rotation based on velocity direction
	rotation := math.Atan2(velocityY, velocityX)

	components.Sprite.Set(k, &components.SpriteData{
		Image:    img,
		Rotation: rotation,
		PivotX:   float64(img.Bounds().Dx()) / 2,
		PivotY:   float64(img.Bounds().Dy()) / 2,
	})

	return k
}
