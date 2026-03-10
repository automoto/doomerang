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

// RunSummaryScene displays per-run statistics after completing a roguelite run.
type RunSummaryScene struct {
	ecs          *ecs.ECS
	sceneChanger SceneChanger
	once         sync.Once
	stats        systems.FinalRunStats
}

// NewRunSummaryScene creates a new run summary scene with the given stats.
func NewRunSummaryScene(sc SceneChanger, stats systems.FinalRunStats) *RunSummaryScene {
	return &RunSummaryScene{
		sceneChanger: sc,
		stats:        stats,
	}
}

func (rs *RunSummaryScene) Update() {
	rs.once.Do(rs.configure)
	rs.ecs.Update()
}

func (rs *RunSummaryScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)
	if rs.ecs == nil {
		return
	}
	rs.ecs.Draw(screen)
}

func (rs *RunSummaryScene) configure() {
	// Persist lifetime stats before anything else
	_ = systems.SaveRogueliteLifetimeStats(rs.stats)

	rs.ecs = ecs.NewECS(donburi.NewWorld())

	createRogueliteScene := func() interface{} {
		return NewRogueliteScene(rs.sceneChanger)
	}
	createMenuScene := NewMainMenuFactory(rs.sceneChanger)

	rs.ecs.AddSystem(systems.UpdateAudio)
	rs.ecs.AddSystem(systems.UpdateInput)
	rs.ecs.AddSystem(systems.NewUpdateRunSummary(rs.sceneChanger, rs.stats, createRogueliteScene, createMenuScene))
	rs.ecs.AddRenderer(cfg.Default, systems.DrawRunSummary(rs.stats))

	systems.PlayMusic(rs.ecs, cfg.Sound.MenuMusic)
}
