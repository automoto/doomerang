package procgen

import (
	"math/rand"

	"github.com/automoto/doomerang/assets"
)

// HazardPlacer handles dynamic hazard placement within chunks
type HazardPlacer struct {
	rng *rand.Rand
}

// NewHazardPlacer creates a hazard placer with the given RNG
func NewHazardPlacer(rng *rand.Rand) *HazardPlacer {
	return &HazardPlacer{rng: rng}
}

// PlaceHazards generates dead zones and fire spawns for a placed chunk.
// Returns hazards in world-space coordinates.
func (hp *HazardPlacer) PlaceHazards(pc PlacedChunk, difficulty int) ([]assets.DeadZone, []assets.FireSpawn) {
	var deadZones []assets.DeadZone
	var fires []assets.FireSpawn

	chunk := pc.Chunk
	ox := pc.OffsetX
	oy := pc.OffsetY

	for _, slot := range chunk.HazardSlots {
		// Activation probability scales with difficulty
		prob := 0.3 + float64(difficulty)*0.15
		if prob > 1.0 {
			prob = 1.0
		}
		if hp.rng.Float64() > prob {
			continue
		}

		switch slot.SlotType {
		case "deadzone":
			deadZones = append(deadZones, assets.DeadZone{
				X:      slot.X + ox,
				Y:      slot.Y + oy,
				Width:  slot.Width,
				Height: slot.Height,
			})
		case "fire_pulsing", "fire_continuous":
			fires = append(fires, assets.FireSpawn{
				X:         slot.X + ox,
				Y:         slot.Y + oy,
				FireType:  slot.SlotType,
				Direction: "up",
			})
		}
	}

	return deadZones, fires
}
