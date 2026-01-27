package fonts

import (
	"fmt"

	"github.com/golang/freetype/truetype"
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

var (
	fonts = map[FontName]font.Face{}
)

func LoadFont(name FontName, ttf []byte) {
	LoadFontWithSize(name, ttf, 10)
}

func LoadFontWithSize(name FontName, ttf []byte, size float64) {
	fontData, _ := truetype.Parse(ttf)
	fonts[name] = truetype.NewFace(fontData, &truetype.Options{Size: size})
}

func getFont(name FontName) font.Face {
	f, ok := fonts[name]
	if !ok {
		panic(fmt.Sprintf("Font %s not found", name))
	}
	return f
}
