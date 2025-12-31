package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/config"
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
	// Track distance
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
		// Owner dead or gone? Destroy boomerang
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

	// Normalize
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist > 0 {
		dirX := dx / dist
		dirY := dy / dist

		// Apply Return Speed
		returnSpeed := config.Boomerang.ReturnSpeed
		physics.SpeedX = dirX * returnSpeed
		physics.SpeedY = dirY * returnSpeed
	}
}

func SwitchToInbound(b *components.BoomerangData, physics *components.PhysicsData) {
	if b.State == components.BoomerangInbound {
		return
	}
	b.State = components.BoomerangInbound
	physics.Gravity = 0 // Disable gravity for homing return
	// Reset hit enemies so we can hit them again on return
	b.HitEnemies = b.HitEnemies[:0]
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
				handleEnemyCollision(e, b, physics, enemyObj)
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

func handleEnemyCollision(boomerangEntry *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, enemyObj *resolv.Object) {
	// Use Data field for O(1) lookup
	enemyEntry, ok := enemyObj.Data.(*donburi.Entry)
	if !ok || enemyEntry == nil || !enemyEntry.Valid() {
		return
	}

	alreadyHit := false
	for _, hit := range b.HitEnemies {
		if hit == enemyEntry {
			alreadyHit = true
			break
		}
	}

	if !alreadyHit {
		// Apply Damage
		if health := components.Health.Get(enemyEntry); health != nil {
			health.Current -= b.Damage

			// Knockback (simplified)
			if enemyPhysics := components.Physics.Get(enemyEntry); enemyPhysics != nil {
				// Knockback away from boomerang
				if b.State == components.BoomerangOutbound {
					enemyPhysics.SpeedX = 2.0
				} else {
					enemyPhysics.SpeedX = -2.0
				}
			}
		}

		// Visual Feedback
		if enemyComp := components.Enemy.Get(enemyEntry); enemyComp != nil {
			enemyComp.InvulnFrames = 15 // Flash for 15 frames
		}

		// Add to hit list
		b.HitEnemies = append(b.HitEnemies, enemyEntry)

		// Short Return Rule
		if b.State == components.BoomerangOutbound {
			newMax := b.DistanceTraveled + b.PierceDistance
			if newMax < b.MaxRange {
				b.MaxRange = newMax
			}
		}
	}
}

func catchBoomerang(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData) {
	// Clear active boomerang on player
	if b.Owner != nil && b.Owner.Valid() {
		if b.Owner.HasComponent(components.Player) {
			player := components.Player.Get(b.Owner)
			player.ActiveBoomerang = nil
		}
	}

	destroyBoomerang(ecs, e, components.Object.Get(e))
}

func destroyBoomerang(ecs *ecs.ECS, e *donburi.Entry, obj *components.ObjectData) {
	// Remove from space
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		if obj != nil && obj.Object != nil {
			components.Space.Get(spaceEntry).Remove(obj.Object)
		}
	}

	// Destroy entity
	ecs.World.Remove(e.Entity())
}
