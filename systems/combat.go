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
		// If the entity is the player, check for invulnerability.
		if e.HasComponent(components.Player) {
			player := components.Player.Get(e)
			if player.InvulnFrames > 0 {
				donburi.Remove[components.DamageEventData](e, components.DamageEvent)
				continue // Skip the rest of the loop for this entity
			}
		}

		hp := components.Health.Get(e)
		hp.Current -= dmg.Amount

		// If the entity is an enemy, show the health bar.
		if e.HasComponent(tags.Enemy) {
			donburi.Add(e, components.HealthBar, &components.HealthBarData{
				TimeToLive: cfg.Combat.HealthBarDuration,
			})
		}

		// Apply knockback if the entity has a physics component and knockback values are set.
		// Note: Knockback from hitbox collisions is already applied in applyHitToEnemy/applyHitToPlayer,
		// so we only apply here if explicit knockback values are provided in the DamageEvent.
		if e.HasComponent(components.Physics) {
			physics := components.Physics.Get(e)
			if dmg.KnockbackX != 0 || dmg.KnockbackY != 0 {
				physics.SpeedX = dmg.KnockbackX
				physics.SpeedY = dmg.KnockbackY
			}

			// Set the entity's state to knockback if it has a state component.
			if e.HasComponent(components.State) {
				state := components.State.Get(e)
				if e.HasComponent(tags.Enemy) {
					// Enemies have a specific hit state
					state.CurrentState = cfg.Hit
				} else {
					state.CurrentState = cfg.Stunned
					if e.HasComponent(components.Player) {
						player := components.Player.Get(e)
						player.InvulnFrames = cfg.Combat.PlayerInvulnFrames
					}
					// Reset melee attack state when hit to prevent getting stuck in charging state
					if e.HasComponent(components.MeleeAttack) {
						melee := components.MeleeAttack.Get(e)
						melee.IsCharging = false
						melee.IsAttacking = false
						melee.HasSpawnedHitbox = false
					}
				}
				state.StateTimer = 0 // Reset state timer
			}
		}

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
			startDeathSequence(ecs, e)
		}
	}
}

// hbEntryHasDeathComponent is a small helper to avoid duplicate death components.
func hbEntryHasDeathComponent(e *donburi.Entry) bool {
	return e.HasComponent(components.Death)
}

func startDeathSequence(ecs *ecs.ECS, e *donburi.Entry) {
	// Play death sound
	PlaySFX(ecs, cfg.SoundDeath)

	// Remove visual effect components to prevent rendering artifacts
	if e.HasComponent(components.Flash) {
		e.RemoveComponent(components.Flash)
	}
	if e.HasComponent(components.SquashStretch) {
		e.RemoveComponent(components.SquashStretch)
	}

	// Add DeathData component with a 60-frame timer.
	donburi.Add(e, components.Death, &components.DeathData{Timer: 60})

	// Switch to die animation if entity has one.
	if e.HasComponent(components.Animation) {
		anim := components.Animation.Get(e)
		anim.SetAnimation(cfg.Die)
		// Freeze on last frame of death animation instead of looping
		if anim.CurrentAnimation != nil {
			anim.CurrentAnimation.FreezeOnComplete = true
		}
	}

	// Zero out movement if it has PlayerData.
	if e.HasComponent(components.Player) {
		physics := components.Physics.Get(e)
		physics.SpeedX = 0
		physics.SpeedY = 0
	}
}
