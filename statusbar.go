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
	result.Update(window)
	return
}

func (bar *StatusBar) Update(window *sdl.Rect) {
	bar.Rect = sdl.Rect{
		X: 0,
		Y: window.H - 22,
		W: window.W,
		H: 22,
	}
	bar.RemainingRect = bar.Rect
}

func (bar *StatusBar) Begin(renderer *sdl.Renderer, theme *StatusBarTheme) {
	bar.RemainingRect = bar.Rect
	DrawRect(renderer, &bar.Rect, theme.BackgroundColor)
}

func (bar *StatusBar) RenderMode(renderer *sdl.Renderer, mode Mode, font *Font, theme *StatusBarTheme) {
	color := theme.GetColorForMode(mode)

	width := font.GetStringWidth(string(mode))
	bgrect := bar.getRectLeft(width + 16 + bar.TriangeImage.Width)
	bgrect.W -= bar.TriangeImage.Width
	DrawRect(renderer, &bgrect, color)
	bar.TriangeImage.Render(renderer, sdl.Point{X: bgrect.X + bgrect.W, Y: bgrect.Y}, color)

	textColor := theme.GetTextColorForMode(mode)
	txtrect := sdl.Rect{
		X: bgrect.X + 8,
		Y: bgrect.Y + (bgrect.H-int32(font.Size))/2,
		W: width,
		H: int32(font.Size),
	}
	DrawText(renderer, font, string(mode), &txtrect, textColor)
}

func (bar *StatusBar) RenderSubmode(renderer *sdl.Renderer, submode Submode, font *Font, theme *StatusBarTheme) {
	if submode == Submode_None {
		return
	}

	width := font.GetStringWidth(string(submode))
	rect := bar.getRectLeft(width + 5)
	rect.X += 5
	rect.Y += (rect.H - int32(font.Size)) / 2
	rect.W = width
	rect.H = int32(font.Size)
	DrawText(renderer, font, string(submode), &rect, theme.TextColor)
}

func (bar *StatusBar) RenderProject(renderer *sdl.Renderer, projectname string, filename string, dirty bool, font *Font, theme *StatusBarTheme) {
	if filename == "" {
		filename = "[untitled]"
	}

	var sb strings.Builder
	sb.WriteString(filename)
	if projectname != "" {
		sb.WriteString(fmt.Sprintf(" - %s", projectname))
	}

	txt := sb.String()
	width := font.GetStringWidth(txt)
	rect := sdl.Rect{
		X: bar.Rect.W/2 - width/2,
		Y: bar.Rect.Y + (bar.Rect.H-int32(font.Size))/2,
		W: width,
		H: int32(font.Size),
	}

	color := theme.TextColor
	if dirty {
		color = theme.DirtyColor
	}

	DrawText(renderer, font, txt, &rect, color)
}

func (bar *StatusBar) RenderLineCount(renderer *sdl.Renderer, text string, font *Font, theme *StatusBarTheme) {
	width := font.GetStringWidth(text)
	rect := bar.getRectRight(width + 8)
	rect.Y += (rect.H - int32(font.Size)) / 2
	rect.W = width
	rect.H = int32(font.Size)

	DrawText(renderer, font, text, &rect, theme.TextColor)
}

func (bar *StatusBar) RenderCaps(renderer *sdl.Renderer, text string, font *Font, theme *StatusBarTheme) {
	width := font.GetStringWidth(text)
	rect := bar.getRectRight(width + 8)
	rect.Y += (rect.H - int32(font.Size)) / 2
	rect.W = width
	rect.H = int32(font.Size)

	DrawText(renderer, font, text, &rect, theme.DirtyColor)
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
