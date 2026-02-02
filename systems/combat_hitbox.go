package systems

import (
	"image/color"

	"github.com/automoto/doomerang/archetypes"
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems/factory"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type HitboxConfig struct {
	Width     float64
	Height    float64
	OffsetX   float64
	OffsetY   float64
	Damage    int
	Knockback float64
	Lifetime  int
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
		playerObject := components.Object.Get(playerEntry).Object

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
		enemyObject := components.Object.Get(enemyEntry).Object

		// Enemies only punch for now
		if state.CurrentState == cfg.StateAttackingPunch && state.StateTimer >= 10 && state.StateTimer <= 15 {
			if !hasActiveHitbox(ecs, enemyEntry) {
				CreateHitbox(ecs, enemyEntry, enemyObject, "punch", false)
			}
		}
	})
}

func hasActiveHitbox(ecs *ecs.ECS, owner *donburi.Entry) bool {
	if owner.HasComponent(components.Player) {
		melee := components.MeleeAttack.Get(owner)
		return melee.ActiveHitbox != nil || melee.HasSpawnedHitbox
	}
	if owner.HasComponent(components.Enemy) {
		return components.Enemy.Get(owner).ActiveHitbox != nil
	}
	return false
}

func CreateHitbox(ecs *ecs.ECS, owner *donburi.Entry, ownerObject *resolv.Object, attackType string, isPlayer bool) {
	// Play attack sound
	switch attackType {
	case "punch":
		PlaySFX(ecs, cfg.SoundPunch)
	case "kick", "jump_kick":
		PlaySFX(ecs, cfg.SoundKick)
	}

	var configs []HitboxConfig

	switch attackType {
	case "punch":
		configs = []HitboxConfig{
			{
				Width:     cfg.Combat.PunchHitboxWidth,
				Height:    cfg.Combat.PunchHitboxHeight,
				OffsetX:   0,
				OffsetY:   0,
				Damage:    cfg.Combat.PlayerPunchDamage,
				Knockback: cfg.Combat.PlayerPunchKnockback,
				Lifetime:  cfg.Combat.HitboxLifetime,
			},
		}
	case "kick":
		configs = []HitboxConfig{
			{
				Width:     cfg.Combat.KickHitboxWidth,
				Height:    cfg.Combat.KickHitboxHeight,
				OffsetX:   0,
				OffsetY:   0,
				Damage:    cfg.Combat.PlayerKickDamage,
				Knockback: cfg.Combat.PlayerKickKnockback,
				Lifetime:  cfg.Combat.HitboxLifetime,
			},
		}
	case "jump_kick":
		// Jump kick uses kick values but multiple hitboxes with longer lifetime
		configs = []HitboxConfig{
			// Main horizontal kick
			{
				Width:     cfg.Combat.KickHitboxWidth,
				Height:    cfg.Combat.KickHitboxHeight,
				OffsetX:   0,
				OffsetY:   0,
				Damage:    cfg.Combat.PlayerKickDamage,
				Knockback: cfg.Combat.PlayerKickKnockback,
				Lifetime:  cfg.Combat.HitboxLifetime * 2,
			},
			// Diagonal hitbox
			{
				Width:     16,
				Height:    16,
				OffsetX:   10,
				OffsetY:   10,
				Damage:    cfg.Combat.PlayerKickDamage,
				Knockback: cfg.Combat.PlayerKickKnockback,
				Lifetime:  cfg.Combat.HitboxLifetime * 2,
			},
			// Downward hitbox
			{
				Width:     12,
				Height:    24,
				OffsetX:   0,
				OffsetY:   20,
				Damage:    cfg.Combat.PlayerKickDamage,
				Knockback: cfg.Combat.PlayerKickKnockback,
				Lifetime:  cfg.Combat.HitboxLifetime * 2,
			},
		}
	}

	// Create shared hit map for all hitboxes in this attack
	sharedHitMap := make(map[*donburi.Entry]bool)

	// Calculate charge ratio before applying bonuses (for VFX scaling)
	var chargeRatio float64
	if isPlayer {
		melee := components.MeleeAttack.Get(owner)
		chargeRatio = float64(melee.ChargeTime) / float64(cfg.Combat.MaxChargeTime)
		if chargeRatio > 1.0 {
			chargeRatio = 1.0
		}
		melee.ChargeTime = 0 // Reset charge time
	}

	for _, config := range configs {
		hitbox := archetypes.Hitbox.Spawn(ecs)

		// Apply charge bonus
		if isPlayer {
			chargeBonus := 1.0 + chargeRatio
			config.Damage = int(float64(config.Damage) * chargeBonus)
			config.Knockback *= chargeBonus
			config.Width *= chargeBonus
			config.Height *= chargeBonus
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
		hitboxObject.Data = hitbox // Linked for O(1) lookup
		components.Object.SetValue(hitbox, components.ObjectData{Object: hitboxObject})

		// Add to space for collision detection
		if spaceEntry, ok := components.Space.First(ecs.World); ok {
			components.Space.Get(spaceEntry).Add(hitboxObject)
		}

		// Set hitbox data
		components.Hitbox.SetValue(hitbox, components.HitboxData{
			OwnerEntity:    owner,
			Damage:         config.Damage,
			KnockbackForce: config.Knockback,
			LifeTime:       config.Lifetime,
			HitEntities:    sharedHitMap,
			AttackType:     attackType,
			ChargeRatio:    chargeRatio,
		})

		// Set active hitbox reference on owner
		if isPlayer {
			melee := components.MeleeAttack.Get(owner)
			melee.ActiveHitbox = hitbox
			melee.HasSpawnedHitbox = true
		} else {
			components.Enemy.Get(owner).ActiveHitbox = hitbox
		}
	}
}

func updateHitboxes(ecs *ecs.ECS) {
	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		hitboxObject := components.Object.Get(hitboxEntry).Object

		// Update hitbox position to follow owner
		updateHitboxPosition(hitbox, hitboxObject)

		// Check for collisions with targets using Resolv space
		checkHitboxCollisions(ecs, hitboxEntry, hitbox, hitboxObject)

		// Decrease lifetime
		hitbox.LifeTime--
	})
}

func updateHitboxPosition(hitbox *components.HitboxData, hitboxObject *resolv.Object) {
	owner := hitbox.OwnerEntity
	if owner == nil || !owner.Valid() {
		return
	}

	ownerObject := components.Object.Get(owner).Object

	// Get facing direction based on owner type
	var directionX float64
	switch {
	case owner.HasComponent(components.Player):
		directionX = components.Player.Get(owner).Direction.X
	case owner.HasComponent(components.Enemy):
		directionX = components.Enemy.Get(owner).Direction.X
	default:
		return
	}

	// Position hitbox in front of owner based on facing direction
	var hitboxX float64
	if directionX > 0 {
		hitboxX = ownerObject.X + ownerObject.W
	} else {
		hitboxX = ownerObject.X - hitboxObject.W
	}
	hitboxY := ownerObject.Y + (ownerObject.H-hitboxObject.H)/2

	hitboxObject.X = hitboxX
	hitboxObject.Y = hitboxY
}

func checkHitboxCollisions(ecs *ecs.ECS, hitboxEntry *donburi.Entry, hitbox *components.HitboxData, hitboxObject *resolv.Object) {
	// Determine if owner is player or enemy
	isPlayerAttack := hitbox.OwnerEntity.HasComponent(components.Player)

	targetTag := tags.ResolvEnemy
	if !isPlayerAttack {
		targetTag = tags.ResolvPlayer
	}

	// Efficient collision check using resolv space
	if check := hitboxObject.Check(0, 0, targetTag); check != nil {
		for _, obj := range check.Objects {
			if targetEntry, ok := obj.Data.(*donburi.Entry); ok {
				if shouldHitTarget(hitbox, targetEntry, hitboxObject, obj) {
					if isPlayerAttack {
						applyHitToEnemy(ecs, targetEntry, hitbox)
					} else {
						applyHitToPlayer(ecs, targetEntry, hitbox)
					}
				}
			}
		}
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
		player := components.Player.Get(target)
		if player.InvulnFrames > 0 {
			return false
		}
	} else if target.HasComponent(components.Enemy) {
		enemy := components.Enemy.Get(target)
		if enemy.InvulnFrames > 0 {
			return false
		}
	}

	return true
}

func applyHitToEnemy(ecs *ecs.ECS, enemyEntry *donburi.Entry, hitbox *components.HitboxData) {
	enemy := components.Enemy.Get(enemyEntry)
	enemyObject := components.Object.Get(enemyEntry).Object

	// Mark as hit
	hitbox.HitEntities[enemyEntry] = true

	// Set Hit state immediately to prevent enemy AI from overriding knockback
	if enemyEntry.HasComponent(components.State) {
		state := components.State.Get(enemyEntry)
		state.CurrentState = cfg.Hit
		state.StateTimer = 0
	}

	// Play hit sound
	PlaySFX(ecs, cfg.SoundHit)

	// Visual effects: flash and screen shake (same for punch, kick, jump kick)
	TriggerHitFlash(enemyEntry)
	TriggerScreenShake(ecs, cfg.ScreenShake.MeleeIntensity, cfg.ScreenShake.MeleeDuration)

	// Spawn hit particles at enemy center, scaled by charge
	hitX := enemyObject.X + enemyObject.W/2
	hitY := enemyObject.Y + enemyObject.H/2
	explosionScale := 0.5 + hitbox.ChargeRatio*0.5 // 50% to 100% size
	factory.SpawnHitExplosion(ecs, hitX, hitY, explosionScale)

	// Apply damage
	donburi.Add(enemyEntry, components.DamageEvent, &components.DamageEventData{
		Amount: hitbox.Damage,
	})

	// Apply knockback
	applyKnockback(enemyEntry, hitbox, enemyObject)

	// Set invulnerability frames
	enemy.InvulnFrames = cfg.Combat.EnemyInvulnFrames
}

func applyHitToPlayer(ecs *ecs.ECS, playerEntry *donburi.Entry, hitbox *components.HitboxData) {
	playerObject := components.Object.Get(playerEntry).Object

	// Mark as hit
	hitbox.HitEntities[playerEntry] = true

	// Play hit sound
	PlaySFX(ecs, cfg.SoundHit)

	// Visual effects: damage flash and stronger screen shake
	TriggerDamageFlash(playerEntry)
	TriggerScreenShake(ecs, cfg.ScreenShake.PlayerDamageIntensity, cfg.ScreenShake.PlayerDamageDuration)

	// Apply damage
	donburi.Add(playerEntry, components.DamageEvent, &components.DamageEventData{
		Amount: hitbox.Damage,
	})

	// Apply knockback
	applyKnockback(playerEntry, hitbox, playerObject)
}

func applyKnockback(targetEntry *donburi.Entry, hitbox *components.HitboxData, targetObject *resolv.Object) {
	ownerObject := components.Object.Get(hitbox.OwnerEntity).Object

	// Determine knockback direction using center points for accuracy
	ownerCenterX := ownerObject.X + ownerObject.W/2
	targetCenterX := targetObject.X + targetObject.W/2

	// Knockback pushes target AWAY from owner
	knockbackDirection := 1.0
	if targetCenterX < ownerCenterX {
		// Target is to the left of owner, push left (negative)
		knockbackDirection = -1.0
	}
	// Otherwise target is to the right, push right (positive)

	// Apply knockback force
	physics := components.Physics.Get(targetEntry)
	physics.SpeedX = knockbackDirection * hitbox.KnockbackForce
	physics.SpeedY = cfg.Combat.KnockbackUpwardForce
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
		// Remove from resolv space
		if spaceEntry, ok := components.Space.First(ecs.World); ok {
			obj := components.Object.Get(hitboxEntry)
			components.Space.Get(spaceEntry).Remove(obj.Object)
		}
		ecs.World.Remove(hitboxEntry.Entity())
	}
}

func DrawHitboxes(ecs *ecs.ECS, screen *ebiten.Image) {
	settings := GetOrCreateSettings(ecs)
	if !settings.Debug {
		return
	}

	// Get camera
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return // No camera yet
	}
	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	// Culling bounds
	padding := 32.0
	minX := camera.Position.X - float64(width)/2 - padding
	maxX := camera.Position.X + float64(width)/2 + padding
	minY := camera.Position.Y - float64(height)/2 - padding
	maxY := camera.Position.Y + float64(height)/2 + padding

	tags.Hitbox.Each(ecs.World, func(hitboxEntry *donburi.Entry) {
		hitbox := components.Hitbox.Get(hitboxEntry)
		o := components.Object.Get(hitboxEntry).Object

		// Viewport Culling
		if o.X+o.W < minX || o.X > maxX || o.Y+o.H < minY || o.Y > maxY {
			return
		}

		// TODO: switch these colors to use the constants we defined in config
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
			hitboxColor = color.RGBA{255, 0, 255, 100} // Magenta (Debug)
		}

		// Apply camera offset
		screenX := float32(o.X + float64(width)/2 - camera.Position.X)
		screenY := float32(o.Y + float64(height)/2 - camera.Position.Y)

		vector.DrawFilledRect(screen, screenX, screenY, float32(o.W), float32(o.H), hitboxColor, false)
	})
}
