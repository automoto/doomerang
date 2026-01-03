package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateCombat handles damage events and keeps health values within their valid range.
func UpdateCombat(ecs *ecs.ECS) {
	processDamageEvents(ecs)
	clampHealthValues(ecs)
}

func processDamageEvents(ecs *ecs.ECS) {
	for e := range components.DamageEvent.Iter(ecs.World) {
		dmg := components.DamageEvent.Get(e)

		if isPlayerInvulnerable(e) {
			donburi.Remove[components.DamageEventData](e, components.DamageEvent)
			continue
		}

		applyDamage(e, dmg)
		showEnemyHealthBar(e)
		applyKnockbackAndState(e, dmg)

		donburi.Remove[components.DamageEventData](e, components.DamageEvent)
	}
}

func isPlayerInvulnerable(e *donburi.Entry) bool {
	if !e.HasComponent(components.Player) {
		return false
	}
	return components.Player.Get(e).InvulnFrames > 0
}

func applyDamage(e *donburi.Entry, dmg *components.DamageEventData) {
	hp := components.Health.Get(e)
	hp.Current -= dmg.Amount
}

func showEnemyHealthBar(e *donburi.Entry) {
	if !e.HasComponent(tags.Enemy) {
		return
	}
	donburi.Add(e, components.HealthBar, &components.HealthBarData{
		TimeToLive: cfg.Combat.HealthBarDuration,
	})
}

func applyKnockbackAndState(e *donburi.Entry, dmg *components.DamageEventData) {
	if !e.HasComponent(components.Physics) {
		return
	}

	physics := components.Physics.Get(e)
	if dmg.KnockbackX != 0 || dmg.KnockbackY != 0 {
		physics.SpeedX = dmg.KnockbackX
		physics.SpeedY = dmg.KnockbackY
	}

	applyHitState(e)
}

func applyHitState(e *donburi.Entry) {
	if !e.HasComponent(components.State) {
		return
	}

	state := components.State.Get(e)
	state.StateTimer = 0

	if e.HasComponent(tags.Enemy) {
		state.CurrentState = cfg.Hit
		return
	}

	state.CurrentState = cfg.Stunned
	setPlayerInvulnFrames(e)
	resetMeleeAttackState(e)
}

func setPlayerInvulnFrames(e *donburi.Entry) {
	if !e.HasComponent(components.Player) {
		return
	}
	components.Player.Get(e).InvulnFrames = cfg.Combat.PlayerInvulnFrames
}

func resetMeleeAttackState(e *donburi.Entry) {
	if !e.HasComponent(components.MeleeAttack) {
		return
	}
	melee := components.MeleeAttack.Get(e)
	melee.IsCharging = false
	melee.IsAttacking = false
	melee.HasSpawnedHitbox = false
}

func clampHealthValues(ecs *ecs.ECS) {
	for e := range components.Health.Iter(ecs.World) {
		hp := components.Health.Get(e)

		if hp.Current < 0 {
			hp.Current = 0
		}
		if hp.Current > hp.Max {
			hp.Current = hp.Max
		}

		if hp.Current == 0 && !e.HasComponent(components.Death) {
			startDeathSequence(e)
		}
	}
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
