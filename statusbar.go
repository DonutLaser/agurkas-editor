package main

import "github.com/veandco/go-sdl2/sdl"

type StatusBarCompnent struct {
	Width          int32
	RenderFunction func(sdl.Rect)
}

type StatusBar struct {
	Layout Layout

	TriangleImage Image

	ComponentsLeft  []StatusBarCompnent
	ComponentsRight []StatusBarCompnent
}

func CreateStatusBar(rect sdl.Rect, renderer *sdl.Renderer) (result StatusBar) {
	result.Layout = CreateLayout(rect)
	result.TriangleImage = LoadImage("./assets/images/status_bar_triangle.png", renderer)
	return
}

func (bar *StatusBar) RegisterComponentLeft(component StatusBarCompnent) {
	bar.ComponentsLeft = append(bar.ComponentsLeft, component)
}

func (bar *StatusBar) RegisterComponentRight(component StatusBarCompnent) {
	bar.ComponentsRight = append(bar.ComponentsRight, component)
}

func (bar *StatusBar) Render(renderer *sdl.Renderer) {
	DrawRect(renderer, &bar.Layout.Rect, sdl.Color{R: 48, G: 48, B: 48, A: 255})
}

func (bar *StatusBar) RenderMode(renderer *sdl.Renderer, mode string, color sdl.Color, font *Font) {
	width, height := font.GetStringSize(mode)

	bgRect := bar.Layout.ConsumeSpaceLeft(width + 16)
	DrawRect(renderer, &bgRect, color)

	modeRect := sdl.Rect{
		X: bgRect.X + 8,
		Y: bgRect.Y + (bgRect.H-height)/2,
		W: width,
		H: height,
	}
	DrawText(renderer, font, mode, &modeRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	// @TODO (!important) color constants

	triangleRect := bar.Layout.ConsumeSpaceLeft(bar.TriangleImage.Width)
	bar.TriangleImage.Render(renderer, sdl.Point{X: triangleRect.X, Y: triangleRect.Y}, color)
}

func (bar *StatusBar) RenderSubmode(renderer *sdl.Renderer, mode string, font *Font) {
	width, height := font.GetStringSize(mode)

	rect := bar.Layout.ConsumeSpaceLeft(width + 5)
	rect.X += 5
	rect.Y = rect.Y + (rect.H-height)/2
	rect.W = width
	DrawText(renderer, font, mode, &rect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}

func (bar *StatusBar) RenderLineCount(renderer *sdl.Renderer, text string, font *Font) {
	width, height := font.GetStringSize(text)

	rect := bar.Layout.ConsumeSpaceRight(width + 8)
	rect.Y = rect.Y + (rect.H-height)/2
	rect.W = width
	rect.H = height
	DrawText(renderer, font, text, &rect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}
