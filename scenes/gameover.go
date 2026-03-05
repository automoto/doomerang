package scenes

import (
	"image/color"
	"sync"

	cfg "github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/systems"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/yohamta/donburi"
	"github.com/yohamta/donburi/ecs"
)

// GameOverScene displays the game over screen
type GameOverScene struct {
	ecs              *ecs.ECS
	sceneChanger     SceneChanger
	once             sync.Once
	retrySceneFactory func() interface{} // creates the scene to retry (platformer or roguelite)
}

// NewGameOverScene creates a new game over scene.
// Optional retryFactory overrides the default retry behavior (platformer).
func NewGameOverScene(sc SceneChanger, retryFactory ...func() interface{}) *GameOverScene {
	gs := &GameOverScene{sceneChanger: sc}
	if len(retryFactory) > 0 {
		gs.retrySceneFactory = retryFactory[0]
	}
	return gs
}

func (gs *GameOverScene) Update() {
	gs.once.Do(gs.configure)
	gs.ecs.Update()
}

func (gs *GameOverScene) Draw(screen *ebiten.Image) {
	// Always clear screen to prevent white flashes from OS window background
	screen.Fill(color.Black)

	if gs.ecs == nil {
		return
	}
	gs.ecs.Draw(screen)
}

func (gs *GameOverScene) configure() {
	gs.ecs = ecs.NewECS(donburi.NewWorld())

	// Scene factories
	retryScene := gs.retrySceneFactory
	if retryScene == nil {
		retryScene = func() interface{} {
			return NewPlatformerScene(gs.sceneChanger)
		}
	}
	createMenuScene := func() interface{} {
		return NewMenuScene(gs.sceneChanger)
	}

	// Audio system
	gs.ecs.AddSystem(systems.UpdateAudio)

	// Minimal systems for game over
	gs.ecs.AddSystem(systems.UpdateInput)
	gs.ecs.AddSystem(systems.NewUpdateGameOver(gs.sceneChanger, retryScene, createMenuScene))

	// Renderer
	gs.ecs.AddRenderer(cfg.Default, systems.DrawGameOver)

	// Play menu music on game over screen
	systems.PlayMusic(gs.ecs, cfg.Sound.MenuMusic)
}
