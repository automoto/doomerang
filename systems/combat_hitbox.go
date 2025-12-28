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
	kickHitboxWidth   = 28 // Kicks have larger hitboxes
	kickHitboxHeight  = 20
)

type HitboxConfig struct {
	Width    float64
	Height   float64
	OffsetX  float64
	OffsetY  float64
	Damage   int
	Knockback float64
}

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
		state := components.State.Get(playerEntry)
		playerObject := cfg.GetObject(playerEntry)

		// Check if player is in attack state and at the right frame
		shouldCreateHitbox := false
		attackType := ""

		switch state.CurrentState {
		case cfg.StateAttackingPunch:
			shouldCreateHitbox = true
			attackType = "punch"
		case cfg.StateAttackingKick:
			shouldCreateHitbox = true
			attackType = "kick"
		case cfg.StateAttackingJump:
			shouldCreateHitbox = true
			attackType = "jump_kick"
		}

		if shouldCreateHitbox {
			// Check if hitbox already exists for this attack
			if !hasActiveHitbox(ecs, playerEntry) {
				CreateHitbox(ecs, playerEntry, playerObject, attackType, true)
			}
		}
	})
}

func createEnemyHitboxes(ecs *ecs.ECS) {
	tags.Enemy.Each(ecs.World, func(enemyEntry *donburi.Entry) {
		state := components.State.Get(enemyEntry)
		enemyObject := cfg.GetObject(enemyEntry)

		// Enemies only punch for now
		if state.CurrentState == "attack" && state.StateTimer >= 10 && state.StateTimer <= 15 {
			if !hasActiveHitbox(ecs, enemyEntry) {
				CreateHitbox(ecs, enemyEntry, enemyObject, "punch", false)
			}
		}
	})
}

func hasActiveHitbox(ecs *ecs.ECS, owner *donburi.Entry) bool {
	if owner.HasComponent(components.Player) {
		return components.MeleeAttack.Get(owner).ActiveHitbox != nil
	}
	if owner.HasComponent(components.Enemy) {
		return components.Enemy.Get(owner).ActiveHitbox != nil
	}
	return false
}

func CreateHitbox(ecs *ecs.ECS, owner *donburi.Entry, ownerObject *resolv.Object, attackType string, isPlayer bool) {
	var configs []HitboxConfig

	switch attackType {
	case "punch":
		configs = []HitboxConfig{
			{Width: punchHitboxWidth, Height: punchHitboxHeight, OffsetX: 0, OffsetY: 0, Damage: punchDamage, Knockback: punchKnockback},
		}
	case "kick":
		configs = []HitboxConfig{
			{Width: kickHitboxWidth, Height: kickHitboxHeight, OffsetX: 0, OffsetY: 0, Damage: kickDamage, Knockback: kickKnockback},
		}
	case "jump_kick":
		configs = []HitboxConfig{
			// Main horizontal kick
			{Width: kickHitboxWidth, Height: kickHitboxHeight, OffsetX: 0, OffsetY: 0, Damage: kickDamage, Knockback: kickKnockback},
			// Diagonal hitbox
			{Width: 16, Height: 16, OffsetX: 10, OffsetY: 10, Damage: kickDamage, Knockback: kickKnockback},
			// Downward hitbox
			{Width: 12, Height: 24, OffsetX: 0, OffsetY: 20, Damage: kickDamage, Knockback: kickKnockback},
		}
	}

	for _, config := range configs {
		hitbox := archetypes.Hitbox.Spawn(ecs)

		// Apply charge bonus
		if isPlayer {
			melee := components.MeleeAttack.Get(owner)
			chargeBonus := 1.0 + (melee.ChargeTime / 60.0) // Add 1% bonus for every frame charged
			config.Damage = int(float64(config.Damage) * chargeBonus)
			config.Knockback *= chargeBonus
			config.Width *= chargeBonus
			config.Height *= chargeBonus
			melee.ChargeTime = 0 // Reset charge time
		}

		// Position hitbox in front of attacker
		var hitboxX, hitboxY float64

		if isPlayer {
			player := components.Player.Get(owner)
			if player.Direction.X > 0 {
				hitboxX = ownerObject.X + ownerObject.W + config.OffsetX
			} else {
				hitboxX = ownerObject.X - config.Width - config.OffsetX
			}
		} else {
			enemy := components.Enemy.Get(owner)
			if enemy.Direction.X > 0 {
				hitboxX = ownerObject.X + ownerObject.W + config.OffsetX
			} else {
				hitboxX = ownerObject.X - config.Width - config.OffsetX
			}
		}

		hitboxY = ownerObject.Y + (ownerObject.H-config.Height)/2 + config.OffsetY // Center vertically

		// Create hitbox object
		hitboxObject := resolv.NewObject(hitboxX, hitboxY, config.Width, config.Height)
		hitboxObject.SetShape(resolv.NewRectangle(0, 0, config.Width, config.Height))
		cfg.SetObject(hitbox, hitboxObject)

		// Set hitbox data
		components.Hitbox.SetValue(hitbox, components.HitboxData{
			OwnerEntity:    owner,
			Damage:         config.Damage,
			KnockbackForce: config.Knockback,
			LifeTime:       10, // Hitbox lasts 10 frames
			HitEntities:    make(map[*donburi.Entry]bool),
			AttackType:     attackType,
		})

		// Set active hitbox reference on owner
		if isPlayer {
			components.MeleeAttack.Get(owner).ActiveHitbox = hitbox
		} else {
			components.Enemy.Get(owner).ActiveHitbox = hitbox
		}
	}
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
	// Don't hit the owner of the hitbox
	if hitbox.OwnerEntity == target {
		return false
	}

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
	applyKnockback(enemyEntry, hitbox, enemyObject)

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
	applyKnockback(playerEntry, hitbox, playerObject)

	// TODO: Set player invulnerability frames (would need to add to PlayerData)
}

func applyKnockback(targetEntry *donburi.Entry, hitbox *components.HitboxData, targetObject *resolv.Object) {
	ownerObject := cfg.GetObject(hitbox.OwnerEntity)

	// Determine knockback direction
	knockbackDirection := 1.0
	if targetObject.X < ownerObject.X {
		knockbackDirection = -1.0
	}

	// Apply knockback force
	physics := components.Physics.Get(targetEntry)
	physics.SpeedX = knockbackDirection * hitbox.KnockbackForce
	physics.SpeedY = -2.0 // Small upward knockback
}

func cleanupHitboxes(ecs *ecs.ECS) {
	var toRemove []*donburi.Entry

	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		if hitbox.LifeTime <= 0 {
			toRemove = append(toRemove, hitboxEntry)
			
			// Clear active hitbox reference on owner
			owner := hitbox.OwnerEntity
			if owner != nil && owner.Valid() {
				if owner.HasComponent(components.Player) {
					melee := components.MeleeAttack.Get(owner)
					if melee.ActiveHitbox == hitboxEntry {
						melee.ActiveHitbox = nil
					}
				} else if owner.HasComponent(components.Enemy) {
					enemy := components.Enemy.Get(owner)
					if enemy.ActiveHitbox == hitboxEntry {
						enemy.ActiveHitbox = nil
					}
				}
			}
		}
	})

	for _, hitboxEntry := range toRemove {
		ecs.World.Remove(hitboxEntry.Entity())
	}
}

func DrawHitboxes(ecs *ecs.ECS, screen *ebiten.Image) {
	settings := GetOrCreateSettings(ecs)
	if !settings.Debug {
		return
	}

	// Get camera
	cameraEntry, _ := components.Camera.First(ecs.World)
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		o := cfg.GetObject(hitboxEntry)

		// Different colors for different attack types
		var hitboxColor color.RGBA
		switch hitbox.AttackType {
		case "punch":
			hitboxColor = color.RGBA{255, 255, 0, 100} // Yellow
		case "kick":
			hitboxColor = color.RGBA{255, 128, 0, 100} // Orange
		case "jump_kick":
			hitboxColor = color.RGBA{0, 255, 0, 100} // Green
		default:
			hitboxColor = color.RGBA{255, 255, 255, 100} // White
		}

		// Apply camera offset
		screenX := float32(o.X + float64(width)/2 - camera.Position.X)
		screenY := float32(o.Y + float64(height)/2 - camera.Position.Y)

		vector.DrawFilledRect(screen, screenX, screenY, float32(o.W), float32(o.H), hitboxColor, false)
	})
}
