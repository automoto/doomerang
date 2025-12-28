package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func init() {
	cfg.Physics = cfg.PhysicsConfig{
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
}

func UpdatePhysics(ecs *ecs.ECS) {
	components.Physics.Each(ecs.World, func(e *donburi.Entry) {
		physics := components.Physics.Get(e)

		// Apply friction and horizontal speed limiting.
		friction := physics.Friction
		if e.HasComponent(components.MeleeAttack) {
			if melee := components.MeleeAttack.Get(e); melee.IsAttacking {
				friction = physics.AttackFriction
			}
		}

		if physics.SpeedX > friction {
			physics.SpeedX -= friction
		} else if physics.SpeedX < -friction {
			physics.SpeedX += friction
		} else {
			physics.SpeedX = 0
		}

		if physics.SpeedX > physics.MaxSpeed {
			physics.SpeedX = physics.MaxSpeed
		} else if physics.SpeedX < -physics.MaxSpeed {
			physics.SpeedX = -physics.MaxSpeed
		}

		// Apply gravity
		physics.SpeedY += physics.Gravity
		if physics.WallSliding != nil && physics.SpeedY > cfg.Physics.WallSlideSpeed {
			physics.SpeedY = cfg.Physics.WallSlideSpeed
		}
	})
}
