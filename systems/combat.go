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

func init() {
	cfg.Combat = cfg.CombatConfig{
		// Player damage values
		PlayerPunchDamage:    10,
		PlayerKickDamage:     15,
		PlayerPunchKnockback: 3.0,
		PlayerKickKnockback:  5.0,

		// Hitbox sizes (these are just examples, adjust as needed)
		PunchHitboxWidth:  20.0,
		PunchHitboxHeight: 10.0,
		KickHitboxWidth:   25.0,
		KickHitboxHeight:  15.0,

		// Timing
		HitboxLifetime:  10,  // frames
		ChargeBonusRate: 0.1, // Bonus per frame charged
		MaxChargeTime:   30,  // frames

		// Invulnerability
		PlayerInvulnFrames: 30,
		EnemyInvulnFrames:  15,

		// Health bar display
		HealthBarDuration: 120, // frames
	}
}

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

		// Apply knockback if the entity has a physics component.
		if e.HasComponent(components.Physics) {
			physics := components.Physics.Get(e)
			physics.SpeedX = dmg.KnockbackX
			physics.SpeedY = dmg.KnockbackY

			// Set the entity's state to knockback if it has a state component.
			if e.HasComponent(components.State) {
				state := components.State.Get(e)
				if e.HasComponent(tags.Enemy) {
					// Enemies have a specific hit state
					// We would need to import the systems package to use enemyStateHit
					// but that would create a circular dependency.
					// So we just use the string "hit" for now.
					state.CurrentState = "hit"
				} else {
					state.CurrentState = cfg.Stunned
					if e.HasComponent(components.Player) {
						player := components.Player.Get(e)
						player.InvulnFrames = cfg.Combat.PlayerInvulnFrames
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
