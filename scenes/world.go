package scenes

import (
	"errors"
	"image/color"
	"sync"

	cfg "github.com/automoto/doomerang/config"
	factory2 "github.com/automoto/doomerang/systems/factory"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/systems"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

type PlatformerScene struct {
	ecs  *ecs.ECS
	once sync.Once
}

func (ps *PlatformerScene) Update() {
	ps.once.Do(ps.configure)
	ps.ecs.Update()
}

func (ps *PlatformerScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{20, 20, 40, 255})
	ps.ecs.Draw(screen)
}

func (ps *PlatformerScene) configure() {
	ecs := ecs.NewECS(donburi.NewWorld())

	// Add systems
	ecs.AddSystem(systems.UpdatePlayer)
	ecs.AddSystem(systems.UpdateEnemies)
	ecs.AddSystem(systems.UpdateStates)
	ecs.AddSystem(systems.UpdatePhysics)
	ecs.AddSystem(systems.UpdateCollisions)
	ecs.AddSystem(systems.UpdateObjects)
	ecs.AddSystem(systems.UpdateBoomerang)
	ecs.AddSystem(systems.UpdateCombat)
	ecs.AddSystem(systems.UpdateCombatHitboxes)
	ecs.AddSystem(systems.UpdateDeaths)
	ecs.AddSystem(systems.UpdateSettings)
	ecs.AddSystem(systems.UpdateCamera)

	// Add renderers
	ecs.AddRenderer(cfg.Default, systems.DrawLevel)
	ecs.AddRenderer(cfg.Default, systems.DrawAnimated)
	ecs.AddRenderer(cfg.Default, systems.DrawSprites)
	ecs.AddRenderer(cfg.Default, systems.DrawHealthBars)
	ecs.AddRenderer(cfg.Default, systems.DrawHitboxes)
	ecs.AddRenderer(cfg.Default, systems.DrawHUD)
	ecs.AddRenderer(cfg.Default, systems.DrawDebug)

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

	// Create collision objects from the level
	for _, path := range levelData.CurrentLevel.Paths {
		// Create a wall object for each ground object
		// The path points represent the top-left and bottom-right corners
		width := path.Points[1].X - path.Points[0].X
		height := path.Points[1].Y - path.Points[0].Y

		// Create a solid wall object using the factory
		factory2.CreateWall(ps.ecs,
			path.Points[0].X,
			path.Points[0].Y,
			width,
			height,
		)
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
