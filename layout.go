package main

import "github.com/veandco/go-sdl2/sdl"

type Layout struct {
	Rect      sdl.Rect
	Remaining sdl.Rect
}

func CreateLayout(rect sdl.Rect) Layout {
	return Layout{
		Rect:      rect,
		Remaining: rect,
	}
}

func (layout *Layout) Update(rect sdl.Rect) {
	layout.Rect = rect
	layout.Remaining = rect
}

func (layout *Layout) Reset() {
	layout.Remaining = layout.Rect
}

func (layout *Layout) ConsumeSpaceLeft(width int32) (result sdl.Rect) {
	result.X = layout.Remaining.X
	result.Y = layout.Remaining.Y
	result.W = width
	result.H = layout.Remaining.H

	layout.Remaining.X += width
	layout.Remaining.W -= width

	return
}

func (layout *Layout) ConsumeSpaceRight(width int32) (result sdl.Rect) {
	result.X = layout.Remaining.X + layout.Remaining.W - width
	result.Y = layout.Remaining.Y
	result.W = width
	result.H = layout.Remaining.H

	layout.Remaining.W -= width

	return
}

func (layout *Layout) ConsumeSpaceTop(height int32) (result sdl.Rect) {
	result.X = layout.Remaining.X
	result.Y = layout.Remaining.Y
	result.W = layout.Remaining.W
	result.H = height

	layout.Remaining.Y += height
	layout.Remaining.H -= height

	return
}

func (layout *Layout) ConsumeSpaceBottom(height int32) (result sdl.Rect) {
	result.X = layout.Remaining.X
	result.Y = layout.Remaining.Y + layout.Remaining.H - height
	result.W = layout.Remaining.W
	result.H = height

	layout.Remaining.H -= height

	return
}
