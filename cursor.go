package main

import "github.com/veandco/go-sdl2/sdl"

type Cursor struct {
	Line        int32
	Column      int32
	WidthNormal int32
	WidthInsert int32
	Height      int32
}

func (cursor *Cursor) Render(renderer *sdl.Renderer, mode Mode, characterWidth int32, gutterWidth int32, windowWidth int32) {
	lineHighlightRect := sdl.Rect{
		X: gutterWidth,
		Y: cursor.Line * cursor.Height,
		W: windowWidth - gutterWidth,
		H: cursor.Height,
	}
	DrawRect(renderer, &lineHighlightRect, sdl.Color{R: 34, G: 35, B: 38, A: 255})

	width := cursor.WidthNormal
	if mode != Mode_Normal {
		width = cursor.WidthInsert
	}

	cursorRect := sdl.Rect{
		X: gutterWidth + 5 + cursor.Column*characterWidth,
		Y: cursor.Line * cursor.Height,
		W: width,
		H: cursor.Height,
	}
	DrawRect(renderer, &cursorRect, sdl.Color{R: 255, G: 218, B: 4, A: 255})
}
