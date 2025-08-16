package scenes

import (
	"image/color"
	"sync"

	dresolv "github.com/automoto/doomerang/config"
	factory2 "github.com/automoto/doomerang/systems/factory"

	"github.com/automoto/doomerang/assets"
	"github.com/automoto/doomerang/components"
	"github.com/automoto/doomerang/layers"
	"github.com/automoto/doomerang/systems"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/solarlune/resolv"
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
	ecs.AddSystem(systems.UpdateCombat)
	ecs.AddSystem(systems.UpdateCombatHitboxes)
	ecs.AddSystem(systems.UpdateDeaths)
	ecs.AddSystem(systems.UpdateSettings)
	ecs.AddSystem(systems.UpdateCamera)

	// Add renderers
	ecs.AddRenderer(layers.Default, systems.DrawLevel)
	ecs.AddRenderer(layers.Default, systems.DrawAnimated)
	// ecs.AddRenderer(layers.Default, systems.DrawHitboxes) // Disabled - ugly debug rendering
	ecs.AddRenderer(layers.Default, systems.DrawHUD)
	ecs.AddRenderer(layers.Default, systems.DrawDebug)
	ecs.AddRenderer(layers.Default, systems.DrawHelp)

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

		// Create a solid wall object
		wall := factory2.CreateWall(ps.ecs, resolv.NewObject(
			path.Points[0].X,
			path.Points[0].Y,
			width,
			height,
			"solid",
		))

		// Add the wall to the collision space
		wallObj := dresolv.GetObject(wall)
		space.Add(wallObj)
	}

	// Create the player at a safe starting position
	player := factory2.CreatePlayer(ps.ecs)
	playerObj := dresolv.GetObject(player)

	// Position the player at a safe starting point
	if len(levelData.CurrentLevel.Paths) > 0 {
		// Find the main ground platform (largest platform by area)
		var mainPlatform assets.Path
		maxArea := 0.0
		for _, path := range levelData.CurrentLevel.Paths {
			width := path.Points[1].X - path.Points[0].X
			height := path.Points[1].Y - path.Points[0].Y
			area := width * height
			if area > maxArea {
				maxArea = area
				mainPlatform = path
			}
		}

		// Position player well above the main platform to avoid any embedding issues
		playerObj.X = mainPlatform.Points[0].X + 200              // Start 200 pixels from the left edge (away from left wall)
		playerObj.Y = mainPlatform.Points[0].Y - playerObj.H - 20 // 20 pixels above the platform for safety
	} else {
		// If no platforms found, place the player at a default position
		playerObj.X = 200
		playerObj.Y = 100
	}

	space.Add(playerObj)

	// Create test enemy
	enemy := factory2.CreateTestEnemy(ps.ecs)
	enemyObj := dresolv.GetObject(enemy)
	space.Add(enemyObj)
}
