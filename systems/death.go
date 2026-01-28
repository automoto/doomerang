package systems

import (
	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/tags"
	"github.com/solarlune/resolv"
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

func handlePlayerDeath(ecs *ecs.ECS, e *donburi.Entry) {
	lives := components.Lives.Get(e)
	death := components.Death.Get(e)

	// Game over delay expired - remove player
	if lives.Lives <= 0 {
		spaceEntry, _ := components.Space.First(e.World)
		space := components.Space.Get(spaceEntry)
		if obj := components.Object.Get(e); obj != nil {
			space.Remove(obj.Object)
		}
		ecs.World.Remove(e.Entity())
		return
	}

	// Death zone already decremented lives at collision time
	if !death.IsDeathZone {
		lives.Lives--
	}

	if lives.Lives <= 0 {
		// Last life lost - delay before game over
		death.Timer = gameOverDelayFrames
		return
	}

	donburi.Remove[components.DeathData](e, components.Death)
	RespawnPlayerNearDeath(ecs, e)
}

// RespawnPlayer resets the player to checkpoint with full health and lives.
func RespawnPlayer(ecs *ecs.ECS, e *donburi.Entry) {
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		return
	}
	levelData := components.Level.Get(levelEntry)

	var spawnX, spawnY float64
	if levelData.ActiveCheckpoint != nil {
		spawnX = levelData.ActiveCheckpoint.SpawnX
		spawnY = levelData.ActiveCheckpoint.SpawnY
	} else if len(levelData.CurrentLevel.PlayerSpawns) > 0 {
		spawn := levelData.CurrentLevel.PlayerSpawns[0]
		spawnX = spawn.X
		spawnY = spawn.Y
	} else {
		return
	}

	resetPlayerAtPosition(e, spawnX, spawnY)

	lives := components.Lives.Get(e)
	lives.Lives = lives.MaxLives

	// Reset message state so messages can be shown again
	ResetMessageState(ecs)
}

// RespawnPlayerNearDeath respawns the player near their death location on safe ground.
func RespawnPlayerNearDeath(ecs *ecs.ECS, e *donburi.Entry) {
	spaceEntry, ok := components.Space.First(ecs.World)
	if !ok {
		RespawnPlayer(ecs, e)
		return
	}

	player := components.Player.Get(e)
	obj := components.Object.Get(e)
	space := components.Space.Get(spaceEntry)

	spawnX, spawnY := player.LastSafeX, player.LastSafeY
	if spawnX == 0 && spawnY == 0 {
		spawnX, spawnY = obj.X, obj.Y
	}

	if !isPositionSafe(space, spawnX, spawnY, obj.W, obj.H) {
		x, y, found := findNearestSafeGround(space, obj.X, obj.Y, obj.W, obj.H)
		if !found {
			RespawnPlayer(ecs, e)
			return
		}
		spawnX, spawnY = x, y
	}

	resetPlayerAtPosition(e, spawnX, spawnY)

	// Reset message state so messages can be shown again
	ResetMessageState(ecs)
}

func resetPlayerAtPosition(e *donburi.Entry, spawnX, spawnY float64) {
	obj := components.Object.Get(e)
	obj.X = spawnX
	obj.Y = spawnY

	physics := components.Physics.Get(e)
	physics.SpeedX = 0
	physics.SpeedY = 0
	physics.OnGround = nil
	physics.WallSliding = nil
	physics.IgnorePlatform = nil

	player := components.Player.Get(e)
	player.InvulnFrames = cfg.Player.RespawnInvulnFrames

	state := components.State.Get(e)
	state.CurrentState = cfg.Idle
	state.StateTimer = 0

	health := components.Health.Get(e)
	health.Current = health.Max
}

func isPositionSafe(space *resolv.Space, x, y, w, h float64) bool {
	tempObj := resolv.NewObject(x, y, w, h)
	space.Add(tempObj)
	defer space.Remove(tempObj)

	if tempObj.Check(0, 0, tags.ResolvDeadZone) != nil {
		return false
	}
	return tempObj.Check(0, 2, tags.ResolvSolid, "platform", tags.ResolvRamp) != nil
}

func findNearestSafeGround(space *resolv.Space, startX, startY, w, h float64) (x, y float64, found bool) {
	const searchStep = 32.0
	const maxSearchDist = 512.0

	// Reuse a single object for all checks to avoid allocations
	tempObj := resolv.NewObject(startX, startY, w, h)
	space.Add(tempObj)
	defer space.Remove(tempObj)

	checkSafe := func(checkX, checkY float64) bool {
		tempObj.X = checkX
		tempObj.Y = checkY
		if tempObj.Check(0, 0, tags.ResolvDeadZone) != nil {
			return false
		}
		return tempObj.Check(0, 2, tags.ResolvSolid, "platform", tags.ResolvRamp) != nil
	}

	// Search left then right
	for _, dir := range []float64{-1, 1} {
		for dist := searchStep; dist <= maxSearchDist; dist += searchStep {
			checkX := startX + dist*dir
			for checkY := startY - 64; checkY <= startY+128; checkY += 16 {
				if checkSafe(checkX, checkY) {
					return checkX, checkY, true
				}
			}
		}
	}

	return 0, 0, false
}
