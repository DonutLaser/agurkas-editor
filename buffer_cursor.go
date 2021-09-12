package main

import "github.com/veandco/go-sdl2/sdl"

type BufferCursor struct {
	Line       int32
	Column     int32
	LastColumn int32

	WidthWide int32
	WidthSlim int32
	Height    int32

	Advance int32

	Color sdl.Color
}

func CreateBufferCursor(height int32, advance int32) (result BufferCursor) {
	result.Line = 0
	result.Column = 0
	result.LastColumn = 0

	result.WidthWide = 8
	result.WidthSlim = 2
	result.Height = height

	result.Advance = advance

	return
}

func (cursor *BufferCursor) Render(renderer *sdl.Renderer, mode Mode, gutterWidth int32, windowWidth int32, scrollOffsetY int32) {
	lineHighlightRect := sdl.Rect{
		X: gutterWidth,
		Y: cursor.Line*cursor.Height + scrollOffsetY,
		W: windowWidth - gutterWidth,
		H: cursor.Height,
	}
	DrawRect(renderer, &lineHighlightRect, sdl.Color{R: 34, G: 35, B: 38, A: 255})

	width := cursor.WidthWide
	if mode != Mode_Normal {
		width = cursor.WidthSlim
	}

	cursorRect := sdl.Rect{
		X: gutterWidth + 5 + cursor.Column*cursor.Advance,
		Y: cursor.Line*cursor.Height + scrollOffsetY,
		W: width,
		H: cursor.Height,
	}
	DrawRect(renderer, &cursorRect, cursor.Color)
}

func (cursor *BufferCursor) GetBottom() int32 {
	return cursor.Line*cursor.Height + cursor.Height
}

func (cursor *BufferCursor) GetTop() int32 {
	return cursor.Line * cursor.Height
}
