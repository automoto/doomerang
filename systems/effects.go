package systems

import (
	"math"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/config"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// UpdateEffects processes visual effect components (flash, squash/stretch, auto-destroy)
func UpdateEffects(ecs *ecs.ECS) {
	updateFlashEffects(ecs)
	updateSquashStretchEffects(ecs)
	updateVFXAnimations(ecs)
	updateAutoDestroy(ecs)
}

// updateVFXAnimations advances animations for VFX entities (they don't have their own update system)
func updateVFXAnimations(ecs *ecs.ECS) {
	components.AutoDestroy.Each(ecs.World, func(e *donburi.Entry) {
		if !e.HasComponent(components.Animation) {
			return
		}
		anim := components.Animation.Get(e)
		if anim.CurrentAnimation != nil {
			anim.CurrentAnimation.Update()
		}
	})
}

// updateFlashEffects decrements flash timers and removes expired flashes
func updateFlashEffects(ecs *ecs.ECS) {
	components.Flash.Each(ecs.World, func(e *donburi.Entry) {
		flash := components.Flash.Get(e)
		if flash.Duration > 0 {
			flash.Duration--
		}
	})
}

// updateSquashStretchEffects lerps scale values toward target and removes when normalized
func updateSquashStretchEffects(ecs *ecs.ECS) {
	var toRemove []*donburi.Entry

	components.SquashStretch.Each(ecs.World, func(e *donburi.Entry) {
		ss := components.SquashStretch.Get(e)

		// Lerp toward target
		ss.ScaleX += (ss.TargetX - ss.ScaleX) * ss.LerpSpeed
		ss.ScaleY += (ss.TargetY - ss.ScaleY) * ss.LerpSpeed

		// Check if close enough to target (normalized)
		threshold := 0.01
		if math.Abs(ss.ScaleX-ss.TargetX) < threshold && math.Abs(ss.ScaleY-ss.TargetY) < threshold {
			toRemove = append(toRemove, e)
		}
	})

	for _, e := range toRemove {
		e.RemoveComponent(components.SquashStretch)
	}
}

// updateAutoDestroy handles entities that should be destroyed after duration or animation
func updateAutoDestroy(ecs *ecs.ECS) {
	var toDestroy []*donburi.Entry

	components.AutoDestroy.Each(ecs.World, func(e *donburi.Entry) {
		ad := components.AutoDestroy.Get(e)

		// Check animation loop condition
		if ad.DestroyOnAnimLoop && e.HasComponent(components.Animation) {
			anim := components.Animation.Get(e)
			if anim.CurrentAnimation != nil && anim.CurrentAnimation.Looped {
				toDestroy = append(toDestroy, e)
				return
			}
		}

		// Check frame countdown
		if ad.FramesRemaining > 0 {
			ad.FramesRemaining--
			if ad.FramesRemaining <= 0 {
				toDestroy = append(toDestroy, e)
			}
		}
	})

	for _, e := range toDestroy {
		// Remove from physics space if it has an object
		if e.HasComponent(components.Object) {
			obj := components.Object.Get(e)
			if obj.Space != nil {
				obj.Space.Remove(obj.Object)
			}
		}
		e.Remove()
	}
}

// TriggerSquashStretch adds a squash/stretch effect to an entity
func TriggerSquashStretch(entry *donburi.Entry, scaleX, scaleY float64) {
	if entry.HasComponent(components.SquashStretch) {
		ss := components.SquashStretch.Get(entry)
		ss.ScaleX = scaleX
		ss.ScaleY = scaleY
		ss.TargetX = 1.0
		ss.TargetY = 1.0
		ss.LerpSpeed = config.SquashStretch.LerpSpeed
	} else {
		entry.AddComponent(components.SquashStretch)
		components.SquashStretch.Set(entry, &components.SquashStretchData{
			ScaleX:    scaleX,
			ScaleY:    scaleY,
			TargetX:   1.0,
			TargetY:   1.0,
			LerpSpeed: config.SquashStretch.LerpSpeed,
		})
	}
}
