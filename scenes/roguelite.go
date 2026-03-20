package scenes

import (
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/automoto/doomerang/assets"
	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/procgen"
	factory2 "github.com/automoto/doomerang/systems/factory"

	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/systems"
	"github.com/automoto/doomerang/tags"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// RogueliteScene generates and plays a procedurally generated level
type RogueliteScene struct {
	ecs          *ecs.ECS
	sceneChanger SceneChanger
	once         sync.Once
	seed         int64
}

// NewRogueliteScene creates a new roguelite scene with a random seed
func NewRogueliteScene(sc SceneChanger) *RogueliteScene {
	return &RogueliteScene{
		sceneChanger: sc,
		seed:         time.Now().UnixNano(),
	}
}

func (rs *RogueliteScene) Update() {
	rs.once.Do(rs.configure)
	rs.ecs.Update()

	if rs.checkGameOver() {
		retryFactory := func() interface{} {
			return NewRogueliteScene(rs.sceneChanger)
		}
		rs.sceneChanger.ChangeScene(NewGameOverScene(rs.sceneChanger, retryFactory))
	}
}

func (rs *RogueliteScene) checkGameOver() bool {
	if rs.ecs == nil {
		return false
	}
	_, ok := tags.Player.First(rs.ecs.World)
	return !ok
}

func (rs *RogueliteScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	if rs.ecs == nil {
		return
	}
	rs.ecs.Draw(screen)
}

func (rs *RogueliteScene) configure() {
	systems.PreloadAllSFX()
	assets.PreloadAllAnimations()

	// Generate the procedural level
	level, result, err := rs.generateLevel()
	if err != nil {
		log.Printf("Procgen failed: %v, falling back to campaign", err)
		rs.sceneChanger.ChangeScene(NewPlatformerScene(rs.sceneChanger))
		return
	}
	// Build room boundaries: right-edge X of each placed chunk
	roomBoundaries := make([]float64, len(result.PlacedChunks))
	for i, pc := range result.PlacedChunks {
		roomBoundaries[i] = pc.OffsetX + float64(pc.Chunk.Width)
	}

	e := ecs.NewECS(donburi.NewWorld())

	// Closure that snapshots run stats and creates the summary scene
	createSummaryScene := func() interface{} {
		snap := systems.SnapshotRunStats(e)
		return NewRunSummaryScene(rs.sceneChanger, snap)
	}

	// Same systems as PlatformerScene
	e.AddSystem(systems.UpdateAudio)
	e.AddSystem(systems.UpdateInput)
	e.AddSystem(systems.UpdatePause)
	e.AddSystem(systems.WithGameplayChecks(systems.UpdatePlayer))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateEnemies))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateStates))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdatePhysics))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateCollisions))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateObjects))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateBoomerang))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateKnives))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateCombat))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateCombatHitboxes))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateDeaths))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateRunStats))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateCheckpoints))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateFire))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateEffects))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateMessage))
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateFinishLine))
	e.AddSystem(systems.NewUpdateRogueliteFinish(rs.sceneChanger, createSummaryScene))
	e.AddSystem(systems.UpdateSettings)
	e.AddSystem(systems.UpdateSettingsMenu)
	e.AddSystem(systems.WithGameplayChecks(systems.UpdateCamera))

	e.AddRenderer(cfg.Default, systems.DrawLevel)
	e.AddRenderer(cfg.Default, systems.DrawAnimated)
	e.AddRenderer(cfg.Default, systems.DrawSprites)
	e.AddRenderer(cfg.Default, systems.DrawHealthBars)
	e.AddRenderer(cfg.Default, systems.DrawHitboxes)
	e.AddRenderer(cfg.Default, systems.DrawHUD)
	e.AddRenderer(cfg.Default, systems.DrawMessage)
	e.AddRenderer(cfg.Default, systems.DrawDebug)
	e.AddRenderer(cfg.Default, systems.DrawPause)
	e.AddRenderer(cfg.Default, systems.DrawSettingsMenu)
	e.AddRenderer(cfg.Default, systems.DrawRogueliteFinish)

	rs.ecs = e

	// Create RunStats entity for this run
	runStatsEntry := e.World.Entry(e.Create(cfg.Default, components.RunStats))
	components.RunStats.SetValue(runStatsEntry, components.RunStatsData{
		Seed:           rs.seed,
		TotalRooms:     len(result.PlacedChunks),
		RoomBoundaries: roomBoundaries,
	})

	// Create level entity with procgen level
	levelEntry := archetypeSpawnLevel(e)
	levelData := &components.LevelData{
		Levels:       []assets.Level{*level},
		LevelIndex:   0,
		CurrentLevel: level,
	}
	components.Level.Set(levelEntry, levelData)

	// Create space
	spaceEntry := factory2.CreateSpace(e, level.Width, level.Height, 16, 16)
	space := components.Space.Get(spaceEntry)

	// Create camera
	factory2.CreateCamera(e)

	// Create collision objects from solid tiles
	for _, tile := range level.SolidTiles {
		if tile.SlopeType != "" {
			factory2.CreateSlopeWall(e, tile.X, tile.Y, tile.Width, tile.Height, tile.SlopeType)
		} else {
			factory2.CreateWall(e, tile.X, tile.Y, tile.Width, tile.Height)
		}
	}

	// Create dead zones
	for _, dz := range level.DeadZones {
		factory2.CreateDeadZone(e, dz.X, dz.Y, dz.Width, dz.Height)
	}

	// Create checkpoints
	for _, ckp := range level.Checkpoints {
		factory2.CreateCheckpoint(e, ckp.X, ckp.Y, ckp.Width, ckp.Height, ckp.CheckpointID)
	}

	// Create fires
	for _, fire := range level.Fires {
		factory2.CreateFire(e, fire.X, fire.Y, fire.FireType, fire.Direction)
	}

	// Create finish lines
	for _, fl := range level.FinishLines {
		factory2.CreateFinishLine(e, fl.X, fl.Y, fl.Width, fl.Height)
	}

	// Spawn player
	spawn := level.PlayerSpawns[0]
	player := factory2.CreatePlayer(e, spawn.X, spawn.Y)
	playerObj := components.Object.Get(player)
	space.Add(playerObj.Object)

	// Snap camera to player
	if cameraEntry, ok := components.Camera.First(e.World); ok {
		camera := components.Camera.Get(cameraEntry)
		camera.Position.X = spawn.X
		camera.Position.Y = spawn.Y
	}

	// Spawn enemies
	for _, es := range level.EnemySpawns {
		enemyType := es.EnemyType
		if enemyType == "" {
			enemyType = "Guard"
		}
		enemy := factory2.CreateEnemy(e, es.X, es.Y, es.PatrolPath, enemyType)
		enemyObj := components.Object.Get(enemy)
		space.Add(enemyObj.Object)
	}

	if len(cfg.Sound.RogueliteMusic) > 0 {
		musicRng := rand.New(rand.NewSource(rs.seed))
		track := cfg.Sound.RogueliteMusic[musicRng.Intn(len(cfg.Sound.RogueliteMusic))]
		systems.PlayMusic(e, track)
	}
}

func (rs *RogueliteScene) generateLevel() (*assets.Level, *procgen.GenerationResult, error) {
	loader := procgen.NewChunkLoader()
	chunks, err := loader.LoadAllChunks("chunks")
	if err != nil {
		return nil, nil, err
	}

	rng := rand.New(rand.NewSource(rs.seed))

	// Pick a single biome for the entire run based on seed
	biome := "cyberpunk"
	if len(cfg.Procgen.Biomes) > 0 {
		biome = cfg.Procgen.Biomes[rng.Intn(len(cfg.Procgen.Biomes))]
	}

	// Generate concept graph with pacing rules
	graph := procgen.GenerateGraph(rng, cfg.Procgen.DefaultRunLength, []string{biome})
	procgen.ValidateGraph(graph)

	// Generate chunks with solvability validation (retries up to 5 times)
	generator := procgen.NewChunkGenerator(rs.seed)
	result, err := procgen.ValidateAndRemediate(generator, chunks, graph, 5)
	if err != nil {
		return nil, nil, err
	}

	// Derive decorative variation (background image + color tint) from seed + biome
	decoration := procgen.DeriveDecoration(rs.seed, biome)
	defer func() {
		if decoration.BackgroundImage != nil {
			decoration.BackgroundImage.Dispose()
		}
	}()

	// Compile base level with decoration
	compiler := procgen.NewCompiler()
	level, err := compiler.Compile(result, &decoration)
	if err != nil {
		return nil, nil, err
	}

	// Dynamic enemy placement
	enemyPlacer := procgen.NewEnemyPlacer(rng)
	for i, pc := range result.PlacedChunks {
		if i < len(graph.Nodes) && (graph.Nodes[i].Type == procgen.NodeCombat || graph.Nodes[i].Type == procgen.NodeArena) {
			spawns, patrolPaths := enemyPlacer.PlaceEnemies(pc, graph.Nodes[i].Difficulty)
			level.EnemySpawns = append(level.EnemySpawns, spawns...)
			for name, path := range patrolPaths {
				level.PatrolPaths[name] = path
			}
		}
	}

	// Dynamic hazard placement
	hazardPlacer := procgen.NewHazardPlacer(rng)
	for i, pc := range result.PlacedChunks {
		diff := 1
		if i < len(graph.Nodes) {
			diff = graph.Nodes[i].Difficulty
		}
		deadZones, fires := hazardPlacer.PlaceHazards(pc, diff)
		level.DeadZones = append(level.DeadZones, deadZones...)
		level.Fires = append(level.Fires, fires...)
	}

	// Auto-place checkpoints at break rooms
	checkpointID := 1.0
	for i, pc := range result.PlacedChunks {
		if i < len(graph.Nodes) && graph.Nodes[i].Type == procgen.NodeBreakRoom {
			level.Checkpoints = append(level.Checkpoints, assets.CheckpointSpawn{
				X:            pc.OffsetX + float64(pc.Chunk.Width)/2,
				Y:            pc.OffsetY + float64(pc.Chunk.Height) - 80,
				Width:        32,
				Height:       48,
				CheckpointID: checkpointID,
			})
			checkpointID++
		}
	}

	return level, result, nil
}

// archetypeSpawnLevel creates a level entity (avoids importing archetypes to prevent cycles)
func archetypeSpawnLevel(e *ecs.ECS) *donburi.Entry {
	return e.World.Entry(e.Create(cfg.Default, components.Level))
}
