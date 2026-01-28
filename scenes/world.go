package scenes

import (
	"errors"
	"image/color"
	"log"
	"sync"

	"github.com/automoto/doomerang/assets"
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
	// Always clear screen to prevent white flashes from OS window background
	screen.Fill(color.Black)

	if ps.ecs == nil {
		return
	}
	ps.ecs.Draw(screen)
}

func (ps *PlatformerScene) configure() {
	// Preload assets to avoid lag on first use (important for WASM)
	systems.PreloadAllSFX()
	assets.PreloadAllAnimations()

	ecs := ecs.NewECS(donburi.NewWorld())

	// Audio system (runs first, even when paused for menu sounds)
	ecs.AddSystem(systems.UpdateAudio)

	// Systems that always run
	ecs.AddSystem(systems.UpdateInput)
	ecs.AddSystem(systems.UpdatePause)

	// Game systems wrapped with pause and level complete checks
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdatePlayer))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateEnemies))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateStates))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdatePhysics))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateCollisions))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateObjects))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateBoomerang))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateKnives))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateCombat))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateCombatHitboxes))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateDeaths))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateCheckpoints))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateFire))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateEffects))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateMessage))
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateFinishLine))
	ecs.AddSystem(systems.UpdateLevelComplete)

	// Systems that run even when paused
	ecs.AddSystem(systems.UpdateSettings)
	ecs.AddSystem(systems.UpdateSettingsMenu)
	ecs.AddSystem(systems.WithGameplayChecks(systems.UpdateCamera))

	// Add renderers
	ecs.AddRenderer(cfg.Default, systems.DrawLevel)
	ecs.AddRenderer(cfg.Default, systems.DrawAnimated)
	ecs.AddRenderer(cfg.Default, systems.DrawSprites)
	ecs.AddRenderer(cfg.Default, systems.DrawHealthBars)
	ecs.AddRenderer(cfg.Default, systems.DrawHitboxes)
	ecs.AddRenderer(cfg.Default, systems.DrawHUD)
	ecs.AddRenderer(cfg.Default, systems.DrawMessage)
	ecs.AddRenderer(cfg.Default, systems.DrawDebug)
	ecs.AddRenderer(cfg.Default, systems.DrawPause)
	ecs.AddRenderer(cfg.Default, systems.DrawSettingsMenu)
	ecs.AddRenderer(cfg.Default, systems.DrawLevelComplete)

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

	// Create fire obstacles from the level
	for _, fire := range levelData.CurrentLevel.Fires {
		factory2.CreateFire(ps.ecs, fire.X, fire.Y, fire.FireType, fire.Direction)
	}

	// Create message points from the level
	for _, msg := range levelData.CurrentLevel.Messages {
		factory2.CreateMessagePoint(ps.ecs, msg.X, msg.Y, msg.MessageID)
	}

	// Create finish lines from the level
	for _, fl := range levelData.CurrentLevel.FinishLines {
		factory2.CreateFinishLine(ps.ecs, fl.X, fl.Y, fl.Width, fl.Height)
	}

	// Determine player spawn position
	var playerSpawnX, playerSpawnY float64
	var foundCheckpoint bool

	// Load saved checkpoint progress
	if progress, _ := systems.LoadGameProgress(); progress != nil && progress.LevelIndex == levelData.LevelIndex {
		levelData.ActiveCheckpoint = &components.ActiveCheckpointData{
			SpawnX:       progress.CheckpointSpawnX,
			SpawnY:       progress.CheckpointSpawnY,
			CheckpointID: progress.CheckpointID,
		}
		playerSpawnX = progress.CheckpointSpawnX
		playerSpawnY = progress.CheckpointSpawnY
		foundCheckpoint = true
	}

	// Check if we should spawn at a specific checkpoint (debug/testing) - overrides saved progress
	if cfg.Debug.StartCheckpoint >= 0 {
		for _, ckp := range levelData.CurrentLevel.Checkpoints {
			if ckp.CheckpointID == cfg.Debug.StartCheckpoint {
				playerSpawnX = ckp.X + ckp.Width/2
				playerSpawnY = ckp.Y + ckp.Height/2
				foundCheckpoint = true

				// Set active checkpoint so respawns work correctly
				levelData.ActiveCheckpoint = &components.ActiveCheckpointData{
					SpawnX:       playerSpawnX,
					SpawnY:       playerSpawnY,
					CheckpointID: ckp.CheckpointID,
				}
				break
			}
		}
		if !foundCheckpoint {
			log.Printf("Warning: Checkpoint %.0f not found, using default spawn", cfg.Debug.StartCheckpoint)
		}
	}

	// Fallback to default spawn
	if !foundCheckpoint {
		if len(levelData.CurrentLevel.PlayerSpawns) <= 0 {
			err := errors.New("no player spawn points defined in Map")
			panic(err)
		}
		spawn := levelData.CurrentLevel.PlayerSpawns[0]
		playerSpawnX = spawn.X
		playerSpawnY = spawn.Y
	}

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
