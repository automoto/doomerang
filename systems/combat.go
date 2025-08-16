package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateCombat handles damage events, debug damage input and keeps health
// values within their valid range.
func UpdateCombat(ecs *ecs.ECS) {
	// --------------------------------------------------------------------
	// 1. Process queued damage events (generic for any entity with Health)
	// --------------------------------------------------------------------
	for e := range components.DamageEvent.Iter(ecs.World) {
		dmg := components.DamageEvent.Get(e)
		hp := components.Health.Get(e)
		hp.Current -= dmg.Amount
		// Remove the damage event component so it is processed only once.
		donburi.Remove[components.DamageEventData](e, components.DamageEvent)
	}

	// --------------------------------------------------------------------
	// 2. Debug: press H to hurt the player by 10 HP
	// --------------------------------------------------------------------
	if inpututil.IsKeyJustPressed(ebiten.KeyH) {
		tags.Player.Each(ecs.World, func(e *donburi.Entry) {
			donburi.Add(e, components.DamageEvent, &components.DamageEventData{Amount: 10})
		})
	}

	// --------------------------------------------------------------------
	// 3. Clamp health ranges (0..Max)
	// --------------------------------------------------------------------
	for e := range components.Health.Iter(ecs.World) {
		hp := components.Health.Get(e)
		if hp.Current < 0 {
			hp.Current = 0
		}
		if hp.Current > hp.Max {
			hp.Current = hp.Max
		}

		// Trigger death sequence if HP reached 0 and not already dying.
		if hp.Current == 0 && !hbEntryHasDeathComponent(e) {
			startDeathSequence(e)
		}
	}
}

// hbEntryHasDeathComponent is a small helper to avoid duplicate death components.
func hbEntryHasDeathComponent(e *donburi.Entry) bool {
	return e.HasComponent(components.Death)
}

func startDeathSequence(e *donburi.Entry) {
	// Add DeathData component with a 60-frame timer.
	donburi.Add(e, components.Death, &components.DeathData{Timer: 60})

	// Switch to die animation if entity has one.
	if e.HasComponent(components.Animation) {
		anim := components.Animation.Get(e)
		anim.SetAnimation(cfg.Die)
	}

	// Zero out movement if it has PlayerData.
	if e.HasComponent(components.Player) {
		physics := components.Physics.Get(e)
		physics.SpeedX = 0
		physics.SpeedY = 0
	}
}
