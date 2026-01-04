package main

import (
	_ "embed"
	"image"
	"log"

	"github.com/automoto/doomerang/config"
	"github.com/automoto/doomerang/fonts"
	"github.com/automoto/doomerang/scenes"
	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/fonts/excel.ttf
var excelFont []byte

type Scene interface {
	Update()
	Draw(screen *ebiten.Image)
}

type Game struct {
	bounds image.Rectangle
	scene  Scene
}

// ChangeScene switches to a new scene
func (g *Game) ChangeScene(scene interface{}) {
	g.scene = scene.(Scene)
}

func NewGame() *Game {
	fonts.LoadFont(fonts.Excel, excelFont)
	fonts.LoadFontWithSize(fonts.ExcelBold, excelFont, 20)
	fonts.LoadFontWithSize(fonts.ExcelTitle, excelFont, 32)

	g := &Game{
		bounds: image.Rectangle{},
	}
	g.scene = scenes.NewMenuScene(g)

	return g
}

func (g *Game) Update() error {
	g.scene.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
	g.bounds = image.Rect(0, 0, width, height)
	return width, height
}

func main() {
	// Start pprof server for memory profiling
	// Usage: go tool pprof http://localhost:6060/debug/pprof/heap
	// go func() {
	// 	log.Println("pprof server running on http://localhost:6060/debug/pprof/")
	// 	if err := http.ListenAndServe("localhost:6060", nil); err != nil {
	// 		log.Printf("pprof server error: %v", err)
	// 	}
	// }()

	ebiten.SetWindowSize(config.C.Width, config.C.Height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewGame()); err != nil {
		log.Fatal(err)
	}
}
