package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func GetCharacterWidth(font *ttf.Font) int32 {
	// We assume that the font is going to always be monospaced
	metrics, err := font.GlyphMetrics('m')
	checkError(err)
	return int32(metrics.Advance)
}

func GetStringSize(font *ttf.Font, text string) (int32, int32) {
	width, height, err := font.SizeUTF8(text)
	checkError(err)

	return int32(width), int32(height)
}

func DrawText(renderer *sdl.Renderer, font *ttf.Font, text string, rect *sdl.Rect, color sdl.Color) {
	surface, err := font.RenderUTF8Blended(text, color)
	checkError(err)

	texture, err := renderer.CreateTextureFromSurface(surface)
	checkError(err)

	renderer.Copy(texture, nil, rect)

	surface.Free()
}

func DrawRect(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	renderer.FillRect(rect)
}
