package main

import (
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type StatusBar struct {
	Rect          sdl.Rect
	RemainingRect sdl.Rect
	TriangeImage  Image
}

func CreateStatusBar(renderer *sdl.Renderer, window *sdl.Rect) (result StatusBar) {
	result.TriangeImage = LoadImage("./assets/images/status_bar_triangle.png", renderer)
	result.Rect = sdl.Rect{
		X: 0,
		Y: window.H - 22,
		W: window.W,
		H: 22,
	}
	result.RemainingRect = result.Rect

	return
}

func (bar *StatusBar) Begin(renderer *sdl.Renderer) {
	bar.RemainingRect = bar.Rect
	DrawRect(renderer, &bar.Rect, sdl.Color{R: 48, G: 48, B: 48, A: 255})
}

func (bar *StatusBar) RenderMode(renderer *sdl.Renderer, mode Mode, color sdl.Color, font *Font) {
	width, height := font.GetStringSize(string(mode))
	bgrect := bar.getRectLeft(width + 16)
	DrawRect(renderer, &bgrect, color)
	bar.TriangeImage.Render(renderer, sdl.Point{X: bgrect.X + bgrect.W, Y: bgrect.Y}, color)

	txtrect := sdl.Rect{
		X: bgrect.X + 8,
		Y: bgrect.Y + (bgrect.H-height)/2,
		W: width,
		H: height,
	}
	DrawText(renderer, font, string(mode), &txtrect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}

func (bar *StatusBar) RenderSubmode(renderer *sdl.Renderer, submode Submode, font *Font) {
	if submode == Submode_None {
		return
	}

	width, height := font.GetStringSize(string(submode))
	rect := bar.getRectLeft(width + 5)
	rect.X += 5
	rect.Y += (rect.H - height) / 2
	rect.W = width
	rect.H = height
	DrawText(renderer, font, string(submode), &rect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}

func (bar *StatusBar) RenderProject(renderer *sdl.Renderer, projectname string, filename string, dirty bool, font *Font) {
	if filename == "" {
		filename = "[untitled]"
	}

	var sb strings.Builder
	sb.WriteString(filename)
	if projectname != "" {
		sb.WriteString(fmt.Sprintf(" - %s", projectname))
	}

	txt := sb.String()
	width, height := font.GetStringSize(txt)
	rect := sdl.Rect{
		X: bar.Rect.W/2 - width/2,
		Y: bar.Rect.Y + (bar.Rect.H-height)/2,
		W: width,
		H: height,
	}

	color := sdl.Color{R: 255, G: 255, B: 255, A: 255}
	if dirty {
		color.R = 213
		color.G = 41
		color.B = 65
	}

	DrawText(renderer, font, txt, &rect, color)
}

func (bar *StatusBar) RenderLineCount(renderer *sdl.Renderer, text string, font *Font) {
	width, height := font.GetStringSize(text)
	rect := bar.getRectRight(width + 8)
	rect.Y += (rect.H - height) / 2
	rect.W = width
	rect.H = height
	DrawText(renderer, font, text, &rect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}

func (bar *StatusBar) getRectLeft(width int32) (result sdl.Rect) {
	result.X = bar.RemainingRect.X
	result.Y = bar.RemainingRect.Y
	result.W = width
	result.H = bar.RemainingRect.H

	bar.RemainingRect.X += width
	bar.RemainingRect.W -= width

	return
}

func (bar *StatusBar) getRectRight(width int32) (result sdl.Rect) {
	result.X = bar.RemainingRect.X + bar.RemainingRect.W - width
	result.Y = bar.RemainingRect.Y
	result.W = width
	result.H = bar.RemainingRect.H

	bar.RemainingRect.W -= width

	return
}
