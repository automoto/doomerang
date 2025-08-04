package components

import "github.com/yohamta/donburi"

// DeathData marks an entity that has started its death sequence.
// Timer counts down each frame; when it reaches 0, the entity should be
// removed from the world.
type DeathData struct {
	Timer int
}

var Death = donburi.NewComponentType[DeathData]()
