package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/config"
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
		obj.Object.X += physics.SpeedX
		obj.Object.Y += physics.SpeedY
		
		// Update shape position for collision check
		obj.Object.Update()

		// 4. Collision Check
		checkCollisions(ecs, e, b, obj)
	})
}

func updateOutbound(e *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, obj *components.ObjectData) {
	// Physics (Gravity) is handled by systems.UpdatePhysics which runs separately?
	// If UpdatePhysics runs, SpeedY is already updated with Gravity.
	// But wait, UpdatePhysics iterates components.Physics. Boomerang has Physics.
	// So Gravity is applied automatically.

	// Track distance
	speed := math.Sqrt(physics.SpeedX*physics.SpeedX + physics.SpeedY*physics.SpeedY)
	b.DistanceTraveled += speed

	// Check Max Range
	if b.DistanceTraveled >= b.MaxRange {
		SwitchToInbound(b)
	}
}

func updateInbound(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData, physics *components.PhysicsData, obj *components.ObjectData) {
	// Homing Logic
	if b.Owner == nil || !b.Owner.Valid() {
		// Owner dead or gone? Destroy boomerang?
		ecs.World.Remove(e.Entity())
		return
	}

	ownerObj := components.Object.Get(b.Owner).Object
	
	// Target center of owner
	targetX := ownerObj.X + ownerObj.W/2
	targetY := ownerObj.Y + ownerObj.H/2
	
	currentX := obj.Object.X + obj.Object.W/2
	currentY := obj.Object.Y + obj.Object.H/2

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
	
	// Disable Gravity for Inbound (override SpeedY every frame)
	// UpdatePhysics will apply gravity again next frame, but we override it here.
	// To be safe, we could set Gravity to 0 in Physics component when switching state.
}

func SwitchToInbound(b *components.BoomerangData) {
	if b.State == components.BoomerangInbound {
		return
	}
	b.State = components.BoomerangInbound
	// Reset hit enemies so we can hit them again on return
	b.HitEnemies = make([]*donburi.Entry, 0)
}

func checkCollisions(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData, obj *components.ObjectData) {
	// Check for collision with anything
	if check := obj.Object.Check(0, 0, "solid", "Enemy", "Player"); check != nil {
		
		// Wall Collision
		if solids := check.ObjectsByTags("solid"); len(solids) > 0 {
			SwitchToInbound(b)
		}

		// Enemy Collision
		// Note: "Enemy" tag is a donburi tag, resolv tag needs to be consistent. 
		// Assuming resolv object for enemy has "Enemy" tag.
		if enemies := check.ObjectsByTags("Enemy"); len(enemies) > 0 {
			for _, enemyObj := range enemies {
				// Find enemy entry from object? 
				// We need a way to map resolv.Object back to donburi.Entry or Component.
				// Usually done by iterating enemies and checking overlap, or storing Entry in Data.
				// But we are iterating Boomerang.
				
				// Simplified: We assume we can get Entity ID from Object tags or UserData?
				// resolv.Object doesn't store Entity ID by default.
				// However, we can iterate all Enemies and check collision with THIS boomerang?
				// Or better, let's assume we can't easily get the Entry from resolv object here without a map.
				
				// WORKAROUND: Iterate all enemies in the world and check overlap with this boomerang.
				// This is inefficient but works. Better approach: Collision system handles this.
				// But since we are here:
				// Actually, check.ObjectsByTags returns resolv Objects.
				// We need to act on them.
				
				// For now, let's just implement the "Short Return Rule" logic which doesn't need the enemy Entry.
				// But we need to avoid multi-hits on the same enemy.
				// So we really need the Enemy Entity/ID.
				
				// Let's defer damage application to a proper Collision System or do the iteration method.
				handleEnemyCollision(e, b, enemyObj)
			}
		}

		// Player Collision (Catch)
		if b.State == components.BoomerangInbound {
			if players := check.ObjectsByTags("Player"); len(players) > 0 {
				// Check if it's the owner
				// We assume only 1 player or we check if this object matches owner's object
				ownerObj := components.Object.Get(b.Owner).Object
				for _, pObj := range players {
					if pObj == ownerObj {
						catchBoomerang(ecs, e, b)
						return
					}
				}
			}
		}
	}
}

func handleEnemyCollision(boomerangEntry *donburi.Entry, b *components.BoomerangData, enemyObj *resolv.Object) {
	// We need to identify the enemy to track hits.
	// Since we can't easily get the Entry from resolv.Object, we'll use the pointer to the object as ID for now.
	// It's not persistent across saves but fine for runtime.
	
	alreadyHit := false
	for _, hit := range b.HitEnemies {
		// This check is tricky because HitEnemies stores *donburi.Entry
		// We need to compare objects.
		if hit.Valid() {
			hitObj := components.Object.Get(hit).Object
			if hitObj == enemyObj {
				alreadyHit = true
				break
			}
		}
	}

	if !alreadyHit {
		// Apply Damage Logic (Placeholder)
		// To apply damage, we need the Entry.
		// So we MUST find the entry.
		// We can do a reverse lookup if we have a map, or scan all enemies.
		// Scanning is O(N) where N is enemies. acceptable for now.
		var enemyEntry *donburi.Entry
		components.Enemy.Each(boomerangEntry.World, func(e *donburi.Entry) {
			if components.Object.Get(e).Object == enemyObj {
				enemyEntry = e
			}
		})

		if enemyEntry != nil {
			// Apply Damage
			if health := components.Health.Get(enemyEntry); health != nil {
				health.Current -= b.Damage
				
				// Knockback (simplified)
				if physics := components.Physics.Get(enemyEntry); physics != nil {
					// Knockback away from boomerang
					if b.State == components.BoomerangOutbound {
						physics.SpeedX = 2.0 // Just a value
					} else {
						physics.SpeedX = -2.0
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
}

func catchBoomerang(ecs *ecs.ECS, e *donburi.Entry, b *components.BoomerangData) {
	// Clear active boomerang on player
	if b.Owner != nil && b.Owner.Valid() {
		if b.Owner.HasComponent(components.Player) {
			player := components.Player.Get(b.Owner)
			player.ActiveBoomerang = nil
		}
	}
	
	// Destroy entity
	ecs.World.Remove(e.Entity())
}
