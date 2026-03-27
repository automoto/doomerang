package systems

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
)

// drawText draws text using text/v2, preserving the same baseline-relative positioning
// as the deprecated text.Draw(dst, str, face, x, y, clr).
func drawText(dst *ebiten.Image, str string, face *textv2.GoXFace, x, y int, clr color.Color) {
	op := &textv2.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y)-face.Metrics().HAscent)
	op.ColorScale.ScaleWithColor(clr)
	textv2.Draw(dst, str, face, op)
}

// centerTextX calculates the X position to horizontally center text on screen.
func centerTextX(s string, face *textv2.GoXFace, screenWidth float64) int {
	w, _ := textv2.Measure(s, face, 0)
	return int((screenWidth - w) / 2)
}

// measureTextWidth returns the pixel width of the given string with the given face.
func measureTextWidth(s string, face *textv2.GoXFace) int {
	w, _ := textv2.Measure(s, face, 0)
	return int(w)
}

// measureText returns the pixel width and height of the given string with the given face.
func measureText(s string, face *textv2.GoXFace) (int, int) {
	w, h := textv2.Measure(s, face, 0)
	return int(w), int(h)
}
