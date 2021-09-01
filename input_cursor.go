package main

import "github.com/veandco/go-sdl2/sdl"

type InputCursor struct {
	Column  int32
	Width   int32
	Height  int32
	Advance int32
}

func (cursor *InputCursor) Render(renderer *sdl.Renderer, inputRect sdl.Rect, color sdl.Color) {
	rect := sdl.Rect{
		X: inputRect.X + 5 + cursor.Column*cursor.Advance,
		Y: inputRect.Y + 5,
		W: cursor.Width,
		H: cursor.Height,
	}
	DrawRect(renderer, &rect, color)
}
