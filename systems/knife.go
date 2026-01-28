package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateKnives(ecs *ecs.ECS) {
	var toRemove []*donburi.Entry

	// Cache level dimensions outside the loop
	var levelWidth, levelHeight float64
	if level, hasLevel := components.Level.First(ecs.World); hasLevel {
		levelData := components.Level.Get(level)
		levelWidth = float64(levelData.CurrentLevel.Width * 16)   // tile width * tile size
		levelHeight = float64(levelData.CurrentLevel.Height * 16) // tile height * tile size
	}

	components.Knife.Each(ecs.World, func(e *donburi.Entry) {
		physics := components.Physics.Get(e)
		obj := components.Object.Get(e)

		// Update position (straight line movement, no gravity)
		obj.X += physics.SpeedX
		obj.Y += physics.SpeedY
		obj.Update()

		// Check if knife is off-screen (with buffer)
		if levelWidth > 0 {
			if obj.X < -100 || obj.X > levelWidth+100 ||
				obj.Y < -100 || obj.Y > levelHeight+100 {
				toRemove = append(toRemove, e)
				return
			}
		}

		// Check collisions
		if shouldDestroy := checkKnifeCollisions(ecs, e, obj); shouldDestroy {
			toRemove = append(toRemove, e)
		}
	})

	// Remove destroyed knives
	for _, knife := range toRemove {
		destroyKnife(ecs, knife)
	}
}

func checkKnifeCollisions(ecs *ecs.ECS, knifeEntry *donburi.Entry, obj *components.ObjectData) bool {
	knife := components.Knife.Get(knifeEntry)

	// Check for collisions with walls and player
	check := obj.Check(0, 0, tags.ResolvSolid, tags.ResolvPlayer)
	if check == nil {
		return false
	}

	// Wall collision - destroy knife
	if solids := check.ObjectsByTags(tags.ResolvSolid); len(solids) > 0 {
		return true
	}

	// Player collision - deal damage and destroy
	if players := check.ObjectsByTags(tags.ResolvPlayer); len(players) > 0 {
		for _, playerObj := range players {
			playerEntry, ok := playerObj.Data.(*donburi.Entry)
			if ok && playerEntry != nil && playerEntry.Valid() {
				handleKnifePlayerHit(ecs, knifeEntry, knife, playerEntry)
			}
		}
		return true
	}

	return false
}

func handleKnifePlayerHit(ecs *ecs.ECS, knifeEntry *donburi.Entry, knife *components.KnifeData, playerEntry *donburi.Entry) {
	// Check player invulnerability
	player := components.Player.Get(playerEntry)
	if player.InvulnFrames > 0 {
		return
	}

	// Calculate knockback direction (based on knife velocity)
	knifePhysics := components.Physics.Get(knifeEntry)
	knockbackX := cfg.Knife.KnockbackForce
	if knifePhysics.SpeedX < 0 {
		knockbackX = -knockbackX
	}

	// Apply damage via DamageEvent
	donburi.Add(playerEntry, components.DamageEvent, &components.DamageEventData{
		Amount:     knife.Damage,
		KnockbackX: knockbackX,
		KnockbackY: cfg.Combat.KnockbackUpwardForce,
	})

	// Visual feedback
	TriggerDamageFlash(playerEntry)
	TriggerScreenShake(ecs, cfg.ScreenShake.PlayerDamageIntensity, cfg.ScreenShake.PlayerDamageDuration)
	PlaySFX(ecs, cfg.SoundHit)
}

func destroyKnife(ecs *ecs.ECS, knifeEntry *donburi.Entry) {
	if spaceEntry, ok := components.Space.First(ecs.World); ok {
		obj := components.Object.Get(knifeEntry)
		if obj != nil && obj.Object != nil {
			components.Space.Get(spaceEntry).Remove(obj.Object)
		}
	}
	ecs.World.Remove(knifeEntry.Entity())
}
