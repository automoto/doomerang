package fonts

import (
	"fmt"

	"github.com/golang/freetype/truetype"
	textv2 "github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
)

type FontName string

const (
	Excel      FontName = "excel"
	ExcelBold  FontName = "excel-bold"
	ExcelTitle FontName = "excel-title"
	ExcelSmall FontName = "excel-small"
)

func (f FontName) Get() font.Face {
	return getFont(f)
}

func (f FontName) GetV2() *textv2.GoXFace {
	return getGoXFace(f)
}

var (
	fonts    = map[FontName]font.Face{}
	goXFaces = map[FontName]*textv2.GoXFace{}
)

func LoadFont(name FontName, ttf []byte) {
	LoadFontWithSize(name, ttf, 10)
}

func LoadFontWithSize(name FontName, ttf []byte, size float64) {
	fontData, _ := truetype.Parse(ttf)
	face := truetype.NewFace(fontData, &truetype.Options{Size: size})
	fonts[name] = face
	goXFaces[name] = textv2.NewGoXFace(face)
}

func getFont(name FontName) font.Face {
	f, ok := fonts[name]
	if !ok {
		panic(fmt.Sprintf("Font %s not found", name))
	}
	return f
}

func getGoXFace(name FontName) *textv2.GoXFace {
	f, ok := goXFaces[name]
	if !ok {
		panic(fmt.Sprintf("GoXFace %s not found", name))
	}
	return f
}
