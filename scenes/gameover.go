package scenes

import (
	"sync"

	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// GameOverScene displays the game over screen
type GameOverScene struct {
	ecs          *ecs.ECS
	sceneChanger SceneChanger
	once         sync.Once
}

// NewGameOverScene creates a new game over scene
func NewGameOverScene(sc SceneChanger) *GameOverScene {
	return &GameOverScene{sceneChanger: sc}
}

func (gs *GameOverScene) Update() {
	gs.once.Do(gs.configure)
	gs.ecs.Update()
}

func (gs *GameOverScene) Draw(screen *ebiten.Image) {
	if gs.ecs == nil {
		return
	}
	gs.ecs.Draw(screen)
}

func (gs *GameOverScene) configure() {
	gs.ecs = ecs.NewECS(donburi.NewWorld())

	// Scene factories
	createPlatformerScene := func() interface{} {
		return NewPlatformerScene(gs.sceneChanger)
	}
	createMenuScene := func() interface{} {
		return NewMenuScene(gs.sceneChanger)
	}

	// Minimal systems for game over
	gs.ecs.AddSystem(systems.UpdateInput)
	gs.ecs.AddSystem(systems.NewUpdateGameOver(gs.sceneChanger, createPlatformerScene, createMenuScene))

	// Renderer
	gs.ecs.AddRenderer(cfg.Default, systems.DrawGameOver)
}
