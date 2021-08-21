package main

import "github.com/veandco/go-sdl2/sdl"

type Cursor struct {
	Line        int32
	Column      int32
	WidthNormal int32
	WidthInsert int32
	Height      int32
}

func (cursor *Cursor) Render(renderer *sdl.Renderer, mode Mode, characterWidth int32, gutterWidth int32) {
	width := cursor.WidthNormal
	if mode != Mode_Normal {
		width = cursor.WidthInsert
	}

	cursorRect := sdl.Rect{
		X: gutterWidth + 10 + cursor.Column*characterWidth,
		Y: 10 + cursor.Line*cursor.Height,
		W: width,
		H: cursor.Height,
	}
	DrawRect(renderer, &cursorRect, sdl.Color{R: 255, G: 218, B: 4, A: 255})
}
