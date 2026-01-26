package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateFire handles dynamic hitboxes and player collision with fire obstacles
func UpdateFire(ecs *ecs.ECS) {
	// Update all fire entities
	tags.Fire.Each(ecs.World, func(e *donburi.Entry) {
		// Update animation
		if !e.HasComponent(components.Animation) {
			return
		}
		anim := components.Animation.Get(e)
		if anim.CurrentAnimation == nil {
			return
		}
		anim.CurrentAnimation.Update()

		fire := components.Fire.Get(e)

		// Only fire with hitbox phases has dynamic hitboxes
		if fire.HitboxPhases == nil {
			return
		}

		// Get current animation frame and calculate hitbox scale
		frame := anim.CurrentAnimation.Frame()
		scale := getFireHitboxScale(frame, fire.HitboxPhases)

		if scale == 0 {
			fire.Active = false
			return
		}

		fire.Active = true
		updateFireHitbox(e, fire, scale)
	})

	// Check player collision with active fire
	playerEntry, ok := tags.Player.First(ecs.World)
	if !ok {
		return
	}

	player := components.Player.Get(playerEntry)
	playerObj := components.Object.Get(playerEntry)

	// Check collision with fire
	check := playerObj.Check(0, 0, tags.ResolvFire)
	if check == nil {
		return
	}

	fireObjs := check.ObjectsByTags(tags.ResolvFire)
	for _, fireResolv := range fireObjs {
		fireEntry, ok := fireResolv.Data.(*donburi.Entry)
		if !ok || fireEntry == nil {
			continue
		}

		fire := components.Fire.Get(fireEntry)
		if !fire.Active {
			continue // Fire is in inactive phase
		}

		// Calculate knockback direction (push away from fire center)
		fireObj := components.Object.Get(fireEntry)
		knockbackX := calculateFireKnockbackDirection(playerObj.Object, fireObj.Object) * fire.KnockbackForce

		// Always apply knockback to prevent walking through fire
		physics := components.Physics.Get(playerEntry)
		physics.SpeedX = knockbackX
		physics.SpeedY = cfg.Combat.KnockbackUpwardForce

		// Only apply damage if not invulnerable
		if player.InvulnFrames == 0 {
			donburi.Add(playerEntry, components.DamageEvent, &components.DamageEventData{
				Amount:     fire.Damage,
				KnockbackX: knockbackX,
				KnockbackY: cfg.Combat.KnockbackUpwardForce,
			})
			TriggerDamageFlash(playerEntry)
			PlaySFX(ecs, cfg.SoundHit)
		}

		break // Only process one fire collision per frame
	}
}

// calculateFireKnockbackDirection returns -1 or 1 based on player position relative to fire
func calculateFireKnockbackDirection(player, fire *resolv.Object) float64 {
	playerCenterX := player.X + player.W/2
	fireCenterX := fire.X + fire.W/2

	if playerCenterX < fireCenterX {
		return -1.0
	}
	return 1.0
}

// getFireHitboxScale returns the hitbox scale (0.0-1.0) for the given animation frame
// Returns 0 if the frame is not in any phase (hitbox disabled)
func getFireHitboxScale(frame int, phases []cfg.FireHitboxPhase) float64 {
	for _, phase := range phases {
		if frame >= phase.StartFrame && frame <= phase.EndFrame {
			// Linear interpolation between start and end scale
			frameRange := float64(phase.EndFrame - phase.StartFrame)
			if frameRange == 0 {
				return phase.StartScale
			}
			progress := float64(frame-phase.StartFrame) / frameRange
			return phase.StartScale + (phase.EndScale-phase.StartScale)*progress
		}
	}
	return 0 // Frame not in any phase = no hitbox
}

// updateFireHitbox resizes the fire's collision hitbox based on scale
// Hitbox is always centered on sprite center (pre-calculated in FireData)
func updateFireHitbox(e *donburi.Entry, fire *components.FireData, scale float64) {
	obj := components.Object.Get(e)

	newW := fire.BaseWidth * scale
	newH := fire.BaseHeight * scale

	// Center hitbox on pre-calculated sprite center
	obj.X = fire.SpriteCenterX - newW/2
	obj.Y = fire.SpriteCenterY - newH/2
	obj.W = newW
	obj.H = newH
}
