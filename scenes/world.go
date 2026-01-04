package scenes

import (
	"errors"
	"sync"

	cfg "github.com/automoto/doomerang/config"
	factory2 "github.com/automoto/doomerang/systems/factory"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/systems"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type PlatformerScene struct {
	ecs          *ecs.ECS
	sceneChanger SceneChanger
	once         sync.Once
}

// NewPlatformerScene creates a new platformer scene
func NewPlatformerScene(sc SceneChanger) *PlatformerScene {
	return &PlatformerScene{sceneChanger: sc}
}

func (ps *PlatformerScene) Update() {
	ps.once.Do(ps.configure)
	ps.ecs.Update()

	// Check for game over (player has 0 lives)
	if ps.checkGameOver() {
		ps.sceneChanger.ChangeScene(NewGameOverScene(ps.sceneChanger))
	}
}

// checkGameOver returns true if the player has 0 or fewer lives
func (ps *PlatformerScene) checkGameOver() bool {
	if ps.ecs == nil {
		return false
	}

	playerEntry, ok := tags.Player.First(ps.ecs.World)
	if !ok {
		return false
	}

	lives := components.Lives.Get(playerEntry)
	return lives.Lives <= 0
}

func (ps *PlatformerScene) Draw(screen *ebiten.Image) {
	if ps.ecs == nil {
		return // Skip ECS draw until initialized
	}
	ps.ecs.Draw(screen)
}

func (ps *PlatformerScene) configure() {
	ecs := ecs.NewECS(donburi.NewWorld())

	// Systems that always run
	ecs.AddSystem(systems.UpdateInput)
	ecs.AddSystem(systems.UpdatePause)

	// Game systems wrapped with pause check
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdatePlayer))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateEnemies))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateStates))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdatePhysics))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateCollisions))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateObjects))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateBoomerang))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateCombat))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateCombatHitboxes))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateDeaths))

	// Systems that run even when paused
	ecs.AddSystem(systems.UpdateSettings)
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateCamera))

	// Add renderers
	ecs.AddRenderer(cfg.Default, systems.DrawLevel)
	ecs.AddRenderer(cfg.Default, systems.DrawAnimated)
	ecs.AddRenderer(cfg.Default, systems.DrawSprites)
	ecs.AddRenderer(cfg.Default, systems.DrawHealthBars)
	ecs.AddRenderer(cfg.Default, systems.DrawHitboxes)
	ecs.AddRenderer(cfg.Default, systems.DrawHUD)
	ecs.AddRenderer(cfg.Default, systems.DrawDebug)
	ecs.AddRenderer(cfg.Default, systems.DrawPause)

	ps.ecs = ecs

	// Create the level entity and load level data FIRST.
	level := factory2.CreateLevel(ps.ecs)
	levelData := components.Level.Get(level)

	// Now create the space for collision detection using the level's dimensions.
	spaceEntry := factory2.CreateSpace(ps.ecs,
		levelData.CurrentLevel.Width,
		levelData.CurrentLevel.Height,
		16, 16,
	)
	space := components.Space.Get(spaceEntry)

	// Create camera
	factory2.CreateCamera(ps.ecs)

	// Create collision objects from solid tiles
	for _, tile := range levelData.CurrentLevel.SolidTiles {
		if tile.SlopeType != "" {
			factory2.CreateSlopeWall(ps.ecs, tile.X, tile.Y, tile.Width, tile.Height, tile.SlopeType)
		} else {
			factory2.CreateWall(ps.ecs, tile.X, tile.Y, tile.Width, tile.Height)
		}
	}

	// Create dead zones from the level
	for _, dz := range levelData.CurrentLevel.DeadZones {
		factory2.CreateDeadZone(ps.ecs, dz.X, dz.Y, dz.Width, dz.Height)
	}

	// Determine player spawn position
	var playerSpawnX, playerSpawnY float64

	if len(levelData.CurrentLevel.PlayerSpawns) <= 0 {
		err := errors.New("no player spawn points defined in Map")
		panic(err)
	}

	// Use the first player spawn point defined in Tiled
	spawn := levelData.CurrentLevel.PlayerSpawns[0]
	playerSpawnX = spawn.X
	playerSpawnY = spawn.Y

	// Create the player at the determined position
	player := factory2.CreatePlayer(ps.ecs, playerSpawnX, playerSpawnY)
	playerObj := components.Object.Get(player)
	space.Add(playerObj.Object)

	// Spawn enemies for the current level
	for _, spawn := range levelData.CurrentLevel.EnemySpawns {
		// Use the enemy type from the spawn data, default to "Guard" if not specified
		enemyType := spawn.EnemyType
		if enemyType == "" {
			enemyType = "Guard"
		}
		enemy := factory2.CreateEnemy(ps.ecs, spawn.X, spawn.Y, spawn.PatrolPath, enemyType)
		enemyObj := components.Object.Get(enemy)
		space.Add(enemyObj.Object)
	}
}
