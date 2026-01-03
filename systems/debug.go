package systems

import (
	"image/color"

	"github.com/automoto/doomerang/components"
	cfg "github.com/automoto/doomerang/config"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/solarlune/resolv"
	"github.com/yohamta/donburi/ecs"
)

type debugCamera struct {
	offsetX, offsetY float64
}

func DrawDebug(ecs *ecs.ECS, screen *ebiten.Image) {
	settings := GetOrCreateSettings(ecs)
	if !settings.Debug {
		return
	}

	cam, ok := getDebugCamera(ecs, screen)
	if !ok {
		return
	}

	drawCollisionGrid(ecs, screen, cam)
	drawLevelCollisions(ecs, screen, cam)
	drawEntityCollisions(ecs, screen, cam)
}

func getDebugCamera(ecs *ecs.ECS, screen *ebiten.Image) (debugCamera, bool) {
	cameraEntry, ok := components.Camera.First(ecs.World)
	if !ok {
		return debugCamera{}, false
	}

	camera := components.Camera.Get(cameraEntry)
	width, height := screen.Bounds().Dx(), screen.Bounds().Dy()

	return debugCamera{
		offsetX: float64(width)/2 - camera.Position.X,
		offsetY: float64(height)/2 - camera.Position.Y,
	}, true
}

func drawCollisionGrid(ecs *ecs.ECS, screen *ebiten.Image, cam debugCamera) {
	spaceEntry, ok := components.Space.First(ecs.World)
	if !ok {
		return
	}

	space := components.Space.Get(spaceEntry)
	cw := float64(space.CellWidth)
	ch := float64(space.CellHeight)

	for y := 0; y < space.Height(); y++ {
		for x := 0; x < space.Width(); x++ {
			cell := space.Cell(x, y)
			cx := float64(cell.X)*cw + cam.offsetX
			cy := float64(cell.Y)*ch + cam.offsetY

			c := cfg.Debug.GridEmpty
			if cell.Occupied() {
				c = cfg.Debug.GridOccupied
			}

			drawGridCell(screen, cx, cy, cw, ch, c)
		}
	}
}

func drawGridCell(screen *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	fx, fy := float32(x), float32(y)
	fw, fh := float32(w), float32(h)

	vector.StrokeLine(screen, fx, fy, fx+fw, fy, 1, c, false)
	vector.StrokeLine(screen, fx+fw, fy, fx+fw, fy+fh, 1, c, false)
	vector.StrokeLine(screen, fx+fw, fy+fh, fx, fy+fh, 1, c, false)
	vector.StrokeLine(screen, fx, fy+fh, fx, fy, 1, c, false)
}

func drawLevelCollisions(ecs *ecs.ECS, screen *ebiten.Image, cam debugCamera) {
	levelEntry, ok := components.Level.First(ecs.World)
	if !ok {
		return
	}

	levelData := components.Level.Get(levelEntry)
	if levelData.CurrentLevel == nil {
		return
	}

	for _, path := range levelData.CurrentLevel.Paths {
		if len(path.Points) < 2 {
			continue
		}

		p1, p2 := path.Points[0], path.Points[1]
		rectX := float32(p1.X + cam.offsetX)
		rectY := float32(p1.Y + cam.offsetY)
		rectW := float32(p2.X - p1.X)
		rectH := float32(p2.Y - p1.Y)

		vector.DrawFilledRect(screen, rectX, rectY, rectW, rectH, cfg.Debug.CollisionRect, false)
	}
}

func drawEntityCollisions(ecs *ecs.ECS, screen *ebiten.Image, cam debugCamera) {
	spaceEntry, ok := components.Space.First(ecs.World)
	if !ok {
		return
	}

	space := components.Space.Get(spaceEntry)
	for _, obj := range space.Objects() {
		x := obj.X + cam.offsetX
		y := obj.Y + cam.offsetY
		c := getEntityDebugColor(obj)

		drawRectOutline(screen, x, y, obj.W, obj.H, c)
	}
}

func getEntityDebugColor(obj *resolv.Object) color.RGBA {
	switch {
	case obj.HasTags("solid"):
		return cfg.Debug.EntitySolid
	case obj.HasTags("Player"):
		return cfg.Debug.EntityPlayer
	case obj.HasTags("Enemy"):
		return cfg.Debug.EntityEnemy
	case obj.HasTags("Boomerang"):
		return cfg.Debug.EntityBoomerang
	default:
		return cfg.Debug.EntityDefault
	}
}

func drawRectOutline(screen *ebiten.Image, x, y, w, h float64, c color.RGBA) {
	fx, fy := float32(x), float32(y)
	fw, fh := float32(w), float32(h)

	vector.DrawFilledRect(screen, fx, fy, fw, 1, c, false)       // Top
	vector.DrawFilledRect(screen, fx, fy+fh-1, fw, 1, c, false)  // Bottom
	vector.DrawFilledRect(screen, fx, fy, 1, fh, c, false)       // Left
	vector.DrawFilledRect(screen, fx+fw-1, fy, 1, fh, c, false)  // Right
}
