package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdatePhysics(ecs *ecs.ECS) {
	components.Physics.Each(ecs.World, func(e *donburi.Entry) {
		// Skip physics for dying player - freeze in place during death delay
		if e.HasComponent(components.Player) && e.HasComponent(components.Death) {
			return
		}

		physics := components.Physics.Get(e)

		// Determine friction based on state
		friction := physics.Friction
		if e.HasComponent(components.State) {
			state := components.State.Get(e)
			if state.CurrentState == cfg.StateSliding {
				friction = cfg.Player.SlideFriction
			}
		}
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

		// Track last safe ground position for player respawn
		if e.HasComponent(components.Player) && physics.OnGround != nil {
			obj := components.Object.Get(e)
			if obj.Check(0, 0, tags.ResolvDeadZone) == nil {
				player := components.Player.Get(e)
				player.LastSafeX = obj.X
				player.LastSafeY = obj.Y
			}
		}
	})
}
