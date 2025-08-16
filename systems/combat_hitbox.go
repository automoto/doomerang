package systems

import (
	"image/color"

	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// Combat damage values
const (
	punchDamage    = 15  // Base punch damage
	kickDamage     = 22  // Kicks do more damage
	punchKnockback = 3.0 // Punch knockback force
	kickKnockback  = 5.0 // Kick knockback force
	invulnFrames   = 30  // 30 frames of invincibility after hit
)

// Hitbox sizes
const (
	punchHitboxWidth  = 20
	punchHitboxHeight = 16
	kickHitboxWidth   = 28  // Kicks have larger hitboxes
	kickHitboxHeight  = 20
)

func UpdateCombatHitboxes(ecs *ecs.ECS) {
	// Create hitboxes for attacking players
	createPlayerHitboxes(ecs)
	
	// Create hitboxes for attacking enemies
	createEnemyHitboxes(ecs)
	
	// Update existing hitboxes and check for collisions
	updateHitboxes(ecs)
	
	// Clean up expired hitboxes
	cleanupHitboxes(ecs)
}

func createPlayerHitboxes(ecs *ecs.ECS) {
	tags.Player.Each(ecs.World, func(playerEntry *donburi.Entry) {
		player := components.Player.Get(playerEntry)
		playerObject := cfg.GetObject(playerEntry)
		
		// Check if player is in attack state and at the right frame
		shouldCreateHitbox := false
		attackType := ""
		
		switch player.CurrentState {
		case cfg.Punch01, cfg.Punch02, cfg.Punch03:
			// Create hitbox at frame 10-15 of punch animation
			if player.StateTimer >= 10 && player.StateTimer <= 15 {
				shouldCreateHitbox = true
				attackType = "punch"
			}
		case cfg.Kick01:
			// Create hitbox at frame 15-20 of kick animation
			if player.StateTimer >= 15 && player.StateTimer <= 20 {
				shouldCreateHitbox = true
				attackType = "kick"
			}
		}
		
		if shouldCreateHitbox {
			// Check if hitbox already exists for this attack
			if !hasActiveHitbox(ecs, playerEntry) {
				createHitbox(ecs, playerEntry, playerObject, attackType, true)
			}
		}
	})
}

func createEnemyHitboxes(ecs *ecs.ECS) {
	tags.Enemy.Each(ecs.World, func(enemyEntry *donburi.Entry) {
		enemy := components.Enemy.Get(enemyEntry)
		enemyObject := cfg.GetObject(enemyEntry)
		
		// Enemies only punch for now
		if enemy.CurrentState == "attack" && enemy.StateTimer >= 10 && enemy.StateTimer <= 15 {
			if !hasActiveHitbox(ecs, enemyEntry) {
				createHitbox(ecs, enemyEntry, enemyObject, "punch", false)
			}
		}
	})
}

func hasActiveHitbox(ecs *ecs.ECS, owner *donburi.Entry) bool {
	hasHitbox := false
	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		if hitbox.OwnerEntity == owner {
			hasHitbox = true
		}
	})
	return hasHitbox
}

func createHitbox(ecs *ecs.ECS, owner *donburi.Entry, ownerObject *resolv.Object, attackType string, isPlayer bool) {
	hitbox := archetypes.Hitbox.Spawn(ecs)
	
	// Determine hitbox size and damage
	var width, height float64
	var damage int
	var knockback float64
	
	if attackType == "kick" {
		width = kickHitboxWidth
		height = kickHitboxHeight
		damage = kickDamage
		knockback = kickKnockback
	} else { // punch
		width = punchHitboxWidth
		height = punchHitboxHeight
		damage = punchDamage
		knockback = punchKnockback
	}
	
	// Position hitbox in front of attacker
	var hitboxX, hitboxY float64
	
	if isPlayer {
		player := components.Player.Get(owner)
		if player.Direction.X > 0 {
			hitboxX = ownerObject.X + ownerObject.W
		} else {
			hitboxX = ownerObject.X - width
		}
	} else {
		enemy := components.Enemy.Get(owner)
		if enemy.Direction.X > 0 {
			hitboxX = ownerObject.X + ownerObject.W
		} else {
			hitboxX = ownerObject.X - width
		}
	}
	
	hitboxY = ownerObject.Y + (ownerObject.H - height) / 2 // Center vertically
	
	// Create hitbox object
	hitboxObject := resolv.NewObject(hitboxX, hitboxY, width, height)
	hitboxObject.SetShape(resolv.NewRectangle(0, 0, width, height))
	cfg.SetObject(hitbox, hitboxObject)
	
	// Set hitbox data
	components.Hitbox.SetValue(hitbox, components.HitboxData{
		OwnerEntity:    owner,
		Damage:         damage,
		KnockbackForce: knockback,
		LifeTime:       10, // Hitbox lasts 10 frames
		HitEntities:    make(map[*donburi.Entry]bool),
		AttackType:     attackType,
	})
}

func updateHitboxes(ecs *ecs.ECS) {
	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		hitboxObject := cfg.GetObject(hitboxEntry)
		
		// Check for collisions with targets
		checkHitboxCollisions(ecs, hitboxEntry, hitbox, hitboxObject)
		
		// Decrease lifetime
		hitbox.LifeTime--
	})
}

func checkHitboxCollisions(ecs *ecs.ECS, hitboxEntry *donburi.Entry, hitbox *components.HitboxData, hitboxObject *resolv.Object) {
	// Determine if owner is player or enemy
	isPlayerAttack := hitbox.OwnerEntity.HasComponent(components.Player)
	
	if isPlayerAttack {
		// Player hitbox - check collision with enemies
		tags.Enemy.Each(ecs.World, func(enemyEntry *donburi.Entry) {
			if shouldHitTarget(hitbox, enemyEntry, hitboxObject, cfg.GetObject(enemyEntry)) {
				applyHitToEnemy(enemyEntry, hitbox)
			}
		})
	} else {
		// Enemy hitbox - check collision with player
		tags.Player.Each(ecs.World, func(playerEntry *donburi.Entry) {
			if shouldHitTarget(hitbox, playerEntry, hitboxObject, cfg.GetObject(playerEntry)) {
				applyHitToPlayer(playerEntry, hitbox)
			}
		})
	}
}

func shouldHitTarget(hitbox *components.HitboxData, target *donburi.Entry, hitboxObject, targetObject *resolv.Object) bool {
	// Don't hit if already hit this target
	if hitbox.HitEntities[target] {
		return false
	}
	
	// Don't hit if target is invulnerable
	if target.HasComponent(components.Player) {
		// Players don't have invuln frames yet - we'll add this when we integrate with existing combat
	} else if target.HasComponent(components.Enemy) {
		enemy := components.Enemy.Get(target)
		if enemy.InvulnFrames > 0 {
			return false
		}
	}
	
	// Check collision by testing overlap
	return hitboxObject.Shape.Intersection(0, 0, targetObject.Shape) != nil
}

func applyHitToEnemy(enemyEntry *donburi.Entry, hitbox *components.HitboxData) {
	enemy := components.Enemy.Get(enemyEntry)
	enemyObject := cfg.GetObject(enemyEntry)
	
	// Mark as hit
	hitbox.HitEntities[enemyEntry] = true
	
	// Apply damage
	donburi.Add(enemyEntry, components.DamageEvent, &components.DamageEventData{
		Amount: hitbox.Damage,
	})
	
	// Apply knockback
	applyKnockback(enemyEntry, hitbox, enemyObject, true)
	
	// Set invulnerability frames
	enemy.InvulnFrames = invulnFrames
}

func applyHitToPlayer(playerEntry *donburi.Entry, hitbox *components.HitboxData) {
	playerObject := cfg.GetObject(playerEntry)
	
	// Mark as hit
	hitbox.HitEntities[playerEntry] = true
	
	// Apply damage
	donburi.Add(playerEntry, components.DamageEvent, &components.DamageEventData{
		Amount: hitbox.Damage,
	})
	
	// Apply knockback
	applyKnockback(playerEntry, hitbox, playerObject, false)
	
	// TODO: Set player invulnerability frames (would need to add to PlayerData)
}

func applyKnockback(targetEntry *donburi.Entry, hitbox *components.HitboxData, targetObject *resolv.Object, isEnemy bool) {
	ownerObject := cfg.GetObject(hitbox.OwnerEntity)
	
	// Determine knockback direction
	knockbackDirection := 1.0
	if targetObject.X < ownerObject.X {
		knockbackDirection = -1.0
	}
	
	// Apply knockback force
	if isEnemy {
		enemy := components.Enemy.Get(targetEntry)
		enemy.SpeedX = knockbackDirection * hitbox.KnockbackForce
		enemy.SpeedY = -2.0 // Small upward knockback
	} else {
		player := components.Player.Get(targetEntry)
		player.SpeedX = knockbackDirection * hitbox.KnockbackForce
		player.SpeedY = -2.0 // Small upward knockback
	}
}

func cleanupHitboxes(ecs *ecs.ECS) {
	var toRemove []*donburi.Entry
	
	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		if hitbox.LifeTime <= 0 {
			toRemove = append(toRemove, hitboxEntry)
		}
	})
	
	for _, hitboxEntry := range toRemove {
		ecs.World.Remove(hitboxEntry.Entity())
	}
}

func DrawHitboxes(ecs *ecs.ECS, screen *ebiten.Image) {
	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		o := cfg.GetObject(hitboxEntry)
		
		// Different colors for different attack types
		hitboxColor := color.RGBA{255, 255, 0, 100} // Yellow for punch
		if hitbox.AttackType == "kick" {
			hitboxColor = color.RGBA{255, 128, 0, 100} // Orange for kick
		}
		
		// Apply camera offset
		screenX := float32(o.X + float64(width)/2 - camera.Position.X)
		screenY := float32(o.Y + float64(height)/2 - camera.Position.Y)
		
		vector.DrawFilledRect(screen, screenX, screenY, float32(o.W), float32(o.H), hitboxColor, false)
	})
}