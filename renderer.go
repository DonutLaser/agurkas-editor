package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func DrawText(renderer *sdl.Renderer, font *Font, text string, rect *sdl.Rect, color sdl.Color) {
	surface, err := font.Data.RenderUTF8Blended(text, color)
	checkError(err)
	defer surface.Free()

	texture, err := renderer.CreateTextureFromSurface(surface)
	checkError(err)
	defer texture.Destroy()

	renderer.Copy(texture, nil, rect)
}

func DrawRect(renderer *sdl.Renderer, rect *sdl.Rect, color sdl.Color) {
	renderer.SetDrawColor(color.R, color.G, color.B, color.A)
	renderer.FillRect(rect)
}

func DrawImage(renderer *sdl.Renderer, texture *sdl.Texture, rect sdl.Rect, color sdl.Color) {
	texture.SetColorMod(color.R, color.G, color.B)
	renderer.Copy(texture, nil, &rect)
}
