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

func NewGame() *Game {
    fonts.LoadFont(fonts.Excel, excelFont)

    g := &Game{
        bounds: image.Rectangle{},
        scene:  &scenes.PlatformerScene{},
    }

    return g
}

func (g *Game) Update() error {
    g.scene.Update()
    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    screen.Clear()
    g.scene.Draw(screen)
}

func (g *Game) Layout(width, height int) (int, int) {
    g.bounds = image.Rect(0, 0, width, height)
    return width, height
}

func main() {
    ebiten.SetWindowSize(config.C.Width, config.C.Height)
    ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
    if err := ebiten.RunGame(NewGame()); err != nil {
        log.Fatal(err)
    }
}
