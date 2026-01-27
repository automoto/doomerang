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

// SceneChanger allows scenes to trigger transitions
type SceneChanger interface {
	ChangeScene(scene interface{})
}

// MenuScene displays the main menu
type MenuScene struct {
	ecs          *ecs.ECS
	sceneChanger SceneChanger
	once         sync.Once
}

// NewMenuScene creates a new menu scene
func NewMenuScene(sc SceneChanger) *MenuScene {
	return &MenuScene{sceneChanger: sc}
}

func (ms *MenuScene) Update() {
	ms.once.Do(ms.configure)
	ms.ecs.Update()
}

func (ms *MenuScene) Draw(screen *ebiten.Image) {
	// Always clear screen to prevent white flashes from OS window background
	screen.Fill(color.Black)

	if ms.ecs == nil {
		return
	}
	ms.ecs.Draw(screen)
}

func (ms *MenuScene) configure() {
	ms.ecs = ecs.NewECS(donburi.NewWorld())

	// Create platformer scene factory that captures the scene changer
	createPlatformerScene := func() interface{} {
		return NewPlatformerScene(ms.sceneChanger)
	}

	// Audio system (runs first to initialize audio context)
	ms.ecs.AddSystem(systems.UpdateAudio)

	// Minimal systems for menu
	ms.ecs.AddSystem(systems.UpdateInput)
	ms.ecs.AddSystem(systems.NewUpdateMenu(ms.sceneChanger, createPlatformerScene))
	ms.ecs.AddSystem(systems.UpdateSettingsMenu)

	// Renderers (settings draws on top of menu)
	ms.ecs.AddRenderer(cfg.Default, systems.DrawMenu)
	ms.ecs.AddRenderer(cfg.Default, systems.DrawSettingsMenu)

	// Start menu music
	systems.PlayMusic(ms.ecs, cfg.Sound.MenuMusic)
}
