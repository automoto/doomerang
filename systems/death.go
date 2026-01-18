package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

func UpdateDeaths(ecs *ecs.ECS) {
	components.Death.Each(ecs.World, func(e *donburi.Entry) {
		death := components.Death.Get(e)
		death.Timer--
		if death.Timer <= 0 {
			// Handle player death differently - respawn or game over
			if e.HasComponent(tags.Player) {
				handlePlayerDeath(ecs, e)
				return
			}

			// Non-player entity: remove from world
			spaceEntry, _ := components.Space.First(e.World)
			space := components.Space.Get(spaceEntry)
			if obj := components.Object.Get(e); obj != nil {
				space.Remove(obj.Object)
			}
			ecs.World.Remove(e.Entity())
		}
	})
}

// Game over delay in frames (30 frames = 0.5 seconds at 60fps)
const gameOverDelayFrames = 30

// handlePlayerDeath respawns the player or sets up game over state
func handlePlayerDeath(ecs *ecs.ECS, e *donburi.Entry) {
	lives := components.Lives.Get(e)

	// If lives already 0, this is the game over delay expiring - remove player and trigger game over
	if lives.Lives <= 0 {
		// Remove player from physics world and ECS
		spaceEntry, _ := components.Space.First(e.World)
		space := components.Space.Get(spaceEntry)
		if obj := components.Object.Get(e); obj != nil {
			space.Remove(obj.Object)
		}
		ecs.World.Remove(e.Entity())
		return
	}

	lives.Lives--

	if lives.Lives <= 0 {
		// Last life lost - keep Death component for game over delay
		// Reset timer to show death state a bit longer before game over
		death := components.Death.Get(e)
		death.Timer = gameOverDelayFrames
		return
	}

	// Remove death component and respawn
	donburi.Remove[components.DeathData](e, components.Death)
	RespawnPlayer(ecs, e)
}

// RespawnPlayer resets the player to spawn point with full health.
// Exported so it can be called from collision.go for dead zone deaths.
func RespawnPlayer(ecs *ecs.ECS, e *donburi.Entry) {
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		return
	}
	levelData := components.Level.Get(levelEntry)

	// Determine spawn position - use checkpoint if available, else default spawn
	var spawnX, spawnY float64
	if levelData.ActiveCheckpoint != nil {
		spawnX = levelData.ActiveCheckpoint.SpawnX
		spawnY = levelData.ActiveCheckpoint.SpawnY
	} else {
		if len(levelData.CurrentLevel.PlayerSpawns) == 0 {
			return
		}
		spawn := levelData.CurrentLevel.PlayerSpawns[0]
		spawnX = spawn.X
		spawnY = spawn.Y
	}

	// Reset position
	obj := components.Object.Get(e)
	obj.X = spawnX
	obj.Y = spawnY

	// Reset physics
	physics := components.Physics.Get(e)
	physics.SpeedX = 0
	physics.SpeedY = 0
	physics.OnGround = nil
	physics.WallSliding = nil
	physics.IgnorePlatform = nil

	// Grant invulnerability
	player := components.Player.Get(e)
	player.InvulnFrames = cfg.Player.RespawnInvulnFrames

	// Reset state
	state := components.State.Get(e)
	state.CurrentState = cfg.Idle
	state.StateTimer = 0

	// Reset health to full
	health := components.Health.Get(e)
	health.Current = health.Max
}
