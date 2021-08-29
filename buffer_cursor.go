package main

import "github.com/veandco/go-sdl2/sdl"

type BufferCursor struct {
	Line        int32
	Column      int32
	WidthNormal int32
	WidthInsert int32
	Height      int32
	Advance     int32
}

func (cursor *BufferCursor) Render(renderer *sdl.Renderer, mode Mode, color sdl.Color, gutterWidth int32, windowWidth int32, scrollOffsetY int32) {
	lineHighlightRect := sdl.Rect{
		X: gutterWidth,
		Y: cursor.Line*cursor.Height + scrollOffsetY,
		W: windowWidth - gutterWidth,
		H: cursor.Height,
	}
	DrawRect(renderer, &lineHighlightRect, sdl.Color{R: 34, G: 35, B: 38, A: 255})

	width := cursor.WidthNormal
	if mode != Mode_Normal {
		width = cursor.WidthInsert
	}

	cursorRect := sdl.Rect{
		X: gutterWidth + 5 + cursor.Column*cursor.Advance,
		Y: cursor.Line*cursor.Height + scrollOffsetY,
		W: width,
		H: cursor.Height,
	}
	DrawRect(renderer, &cursorRect, color)
}

func (cursor *BufferCursor) GetBottom() int32 {
	return cursor.Line*cursor.Height + cursor.Height
}

func (cursor *BufferCursor) GetTop() int32 {
	return cursor.Line * cursor.Height
}
