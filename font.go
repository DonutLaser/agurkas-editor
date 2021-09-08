package main

import (
	"github.com/veandco/go-sdl2/ttf"
)

type Font struct {
	Data           *ttf.Font
	Size           int
	CharacterWidth int
}

func LoadFont(path string, size int) (result Font) {
	font, err := ttf.OpenFont(path, size)
	checkError(err)

	// We assume that the font is going to always be monospaced
	metrics, err := font.GlyphMetrics('m')
	checkError(err)

	result.Data = font
	result.Size = size
	result.CharacterWidth = metrics.Advance

	return
}

func (font *Font) GetStringWidth(text string) int32 {
	return int32(len(text) * font.CharacterWidth)
}

func (font *Font) Unload() {
	font.Data.Close()
}
