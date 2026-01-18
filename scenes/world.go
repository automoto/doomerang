package scenes

import (
	"errors"
	"image/color"
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

// checkGameOver returns true if the player entity has been removed (after death sequence completes)
func (ps *PlatformerScene) checkGameOver() bool {
	if ps.ecs == nil {
		return false
	}

	// Player entity is removed when game over delay expires
	_, ok := tags.Player.First(ps.ecs.World)
	return !ok
}

func (ps *PlatformerScene) Draw(screen *ebiten.Image) {
	if ps.ecs == nil {
		return // Skip ECS draw until initialized
	}
	
	// Clear screen to black to prevent white flashes from OS window background
	screen.Fill(color.Black)
	
	ps.ecs.Draw(screen)
}

func (ps *PlatformerScene) configure() {
	ecs := ecs.NewECS(donburi.NewWorld())

	// Audio system (runs first, even when paused for menu sounds)
	ecs.AddSystem(systems.UpdateAudio)

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
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateCheckpoints))
	ecs.AddSystem(systems.WithPauseCheck(systems.UpdateEffects))

	// Systems that run even when paused
	ecs.AddSystem(systems.UpdateSettings)
	ecs.AddSystem(systems.UpdateSettingsMenu)
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
	ecs.AddRenderer(cfg.Default, systems.DrawSettingsMenu)

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

	// Create checkpoints from the level
	for _, ckp := range levelData.CurrentLevel.Checkpoints {
		factory2.CreateCheckpoint(ps.ecs, ckp.X, ckp.Y, ckp.Width, ckp.Height, ckp.CheckpointID)
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

	// Snap camera to player start position to prevent panning from (0,0)
	if cameraEntry, ok := components.Camera.First(ps.ecs.World); ok {
		camera := components.Camera.Get(cameraEntry)
		camera.Position.X = playerSpawnX
		camera.Position.Y = playerSpawnY
	}

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

	// Start level music
	systems.PlayLevelMusic(ps.ecs, levelData.CurrentLevel.Name)
}
