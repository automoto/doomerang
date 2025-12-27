package systems

import (
	"github.com/automoto/doomerang/components"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

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
		if physics.WallSliding != nil && physics.SpeedY > 1 {
			physics.SpeedY = 1
		}
	})
}
