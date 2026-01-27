package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems/factory"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateBoomerang(ecs *ecs.ECS) {
	components.Boomerang.Each(ecs.World, func(e *donburi.Entry) {
		b := components.Boomerang.Get(e)
		physics := components.Physics.Get(e)
		obj := components.Object.Get(e)
		sprite := components.Sprite.Get(e)

		// 1. Update Rotation
		sprite.Rotation += 0.3 // Constant spin

		// 2. State Logic
		switch b.State {
		case components.BoomerangOutbound:
			updateOutbound(e, b, physics, obj)
		case components.BoomerangInbound:
			updateInbound(ecs, e, b, physics, obj)
		}

		// 3. Update Position (Manual movement, ignoring standard collision system for now)
		obj.X += physics.SpeedX
		obj.Y += physics.SpeedY

		// Update shape position for collision check
		obj.Update()

		// 4. Collision Check
		checkCollisions(ecs, e, b, physics, obj)
	})
}

func updateOutbound(e *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, obj *components.ObjectData) {
	// Track distance - we still need sqrt here for accurate range tracking 
	// because speed varies, but we could optimize this if needed.
	speed := math.Sqrt(physics.SpeedX*physics.SpeedX + physics.SpeedY*physics.SpeedY)
	b.DistanceTraveled += speed

	// Check Max Range
	if b.DistanceTraveled >= b.MaxRange {
		SwitchToInbound(b, physics)
	}
}

func updateInbound(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, obj *components.ObjectData) {
	// Homing Logic
	if b.Owner == nil || !b.Owner.Valid() {
		destroyBoomerang(ecs, e, obj)
		return
	}

	ownerObj := components.Object.Get(b.Owner)

	// Target center of owner
	targetX := ownerObj.X + ownerObj.W/2
	targetY := ownerObj.Y + ownerObj.H/2

	currentX := obj.X + obj.W/2
	currentY := obj.Y + obj.H/2

	dx := targetX - currentX
	dy := targetY - currentY

	// Squared distance for proximity check
	distSq := dx*dx + dy*dy

	// Only normalize and update velocity if we aren't already "at" the target
	// and to avoid division by zero.
	if distSq > 0.01 {
		dist := math.Sqrt(distSq)
		dirX := dx / dist
		dirY := dy / dist

		// Apply Return Speed
		returnSpeed := cfg.Boomerang.ReturnSpeed
		physics.SpeedX = dirX * returnSpeed
		physics.SpeedY = dirY * returnSpeed
	} else {
		// Stop moving if we are exactly at the player center 
		// (collision check will catch it)
		physics.SpeedX = 0
		physics.SpeedY = 0
	}
}

func SwitchToInbound(b *components.BoomerangData, physics *components.PhysicsData) {
	if b.State == components.BoomerangInbound {
		return
	}
	b.State = components.BoomerangInbound
	physics.Gravity = 0 // Disable gravity for homing return
	// Keep HitEnemies - each enemy should only be hit once per throw
}

func checkCollisions(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, obj *components.ObjectData) {
	// Check for collision with anything
	if check := obj.Check(0, 0, tags.ResolvSolid, tags.ResolvEnemy, tags.ResolvPlayer); check != nil {

		// Wall Collision
		if solids := check.ObjectsByTags(tags.ResolvSolid); len(solids) > 0 {
			SwitchToInbound(b, physics)
		}

		// Enemy Collision
		if enemies := check.ObjectsByTags(tags.ResolvEnemy); len(enemies) > 0 {
			for _, enemyObj := range enemies {
				handleEnemyCollision(ecs, e, b, physics, enemyObj)
			}
		}

		// Player Collision (Catch)
		if b.State == components.BoomerangInbound {
			if players := check.ObjectsByTags(tags.ResolvPlayer); len(players) > 0 {
				ownerObj := components.Object.Get(b.Owner)
				for _, pObj := range players {
					if pObj == ownerObj.Object {
						catchBoomerang(ecs, e, b)
						return
					}
				}
			}
		}
	}
}

func handleEnemyCollision(ecs *ecs.ECS, boomerangEntry *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, enemyObj *resolv.Object) {
	enemyEntry, ok := enemyObj.Data.(*donburi.Entry)
	if !ok || enemyEntry == nil || !enemyEntry.Valid() {
		return
	}

	// O(1) map lookup
	if _, alreadyHit := b.HitEnemies[enemyEntry]; alreadyHit {
		return
	}

	// Play impact sound
	PlaySFX(ecs, cfg.SoundBoomerangImpact)

	// Visual effects: hit flash on enemy
	TriggerHitFlash(enemyEntry)

	// Impact effect at enemy position - scale based on charge
	// Quick throw (0.0) = 50% size, fully charged (1.0) = 100% size
	impactX := enemyObj.X + enemyObj.W/2
	impactY := enemyObj.Y + enemyObj.H/2
	explosionScale := 0.5 + b.ChargeRatio*0.5
	factory.SpawnExplosion(ecs, impactX, impactY, explosionScale)
	TriggerScreenShake(ecs, cfg.ScreenShake.BoomerangIntensity, cfg.ScreenShake.BoomerangDuration)

	// Apply Damage
	if health := components.Health.Get(enemyEntry); health != nil {
		health.Current -= b.Damage

		// Show health bar on hit
		if !enemyEntry.HasComponent(components.HealthBar) {
			donburi.Add(enemyEntry, components.HealthBar, &components.HealthBarData{
				TimeToLive: cfg.Combat.HealthBarDuration,
			})
		} else {
			// Reset timer if health bar already visible
			hb := components.HealthBar.Get(enemyEntry)
			hb.TimeToLive = cfg.Combat.HealthBarDuration
		}

	}

	// Set Hit state to trigger knockback animation and prevent AI override
	if enemyEntry.HasComponent(components.State) {
		state := components.State.Get(enemyEntry)
		state.CurrentState = cfg.Hit
		state.StateTimer = 0
	}

	// Apply knockback similar to melee attacks
	if enemyPhysics := components.Physics.Get(enemyEntry); enemyPhysics != nil {
		boomerangObj := components.Object.Get(boomerangEntry).Object
		boomerangCenterX := boomerangObj.X + boomerangObj.W/2
		enemyCenterX := enemyObj.X + enemyObj.W/2

		// Knockback pushes enemy AWAY from boomerang
		knockbackDirection := 1.0
		if enemyCenterX < boomerangCenterX {
			knockbackDirection = -1.0
		}

		enemyPhysics.SpeedX = knockbackDirection * cfg.Boomerang.HitKnockback
		enemyPhysics.SpeedY = cfg.Combat.KnockbackUpwardForce
	}

	// Visual Feedback - use half the normal enemy invuln frames for boomerang hits
	if enemyComp := components.Enemy.Get(enemyEntry); enemyComp != nil {
		enemyComp.InvulnFrames = cfg.Combat.EnemyInvulnFrames / 2
	}

	// Add to hit map
	b.HitEnemies[enemyEntry] = struct{}{}

	// Short Return Rule
	if b.State == components.BoomerangOutbound {
		newMax := b.DistanceTraveled + b.PierceDistance
		if newMax < b.MaxRange {
			b.MaxRange = newMax
		}
	}
}

func catchBoomerang(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData) {
	// Play catch sound
	PlaySFX(ecs, cfg.SoundBoomerangCatch)

	if b.Owner != nil && b.Owner.Valid() {
		if b.Owner.HasComponent(components.Player) {
			player := components.Player.Get(b.Owner)
			player.ActiveBoomerang = nil
		}
	}

	destroyBoomerang(ecs, e, components.Object.Get(e))
}

func destroyBoomerang(ecs *ecs.ECS, e *donburi.Entry, obj *components.ObjectData) {
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		if obj != nil && obj.Object != nil {
			components.Space.Get(spaceEntry).Remove(obj.Object)
		}
	}
	ecs.World.Remove(e.Entity())
}
