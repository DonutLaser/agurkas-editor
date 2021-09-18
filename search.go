package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type Search struct {
	Input strings.Builder

	Cursor      InputCursor
	Width       int32
	LineHeight  int32
	LineSpacing int32
	Font        *Font

	CloseCallback func(string)
}

func CreateSearch(lineHeight int32, font *Font) (result Search) {
	result.Cursor = CreateInputCursor(lineHeight, int32(font.CharacterWidth))
	result.Width = 500
	result.LineHeight = lineHeight
	result.LineSpacing = (lineHeight - int32(font.Size)) / 2
	result.Font = font

	return
}

func (search *Search) Open(closeCallback func(string)) {
	search.Cursor.Column = 0
	search.Input.Reset()

	search.CloseCallback = closeCallback
}

func (search *Search) Close() {
	search.CloseCallback(search.Input.String())
}

func (search *Search) Tick(input Input) {
	if input.Escape {
		search.Close()
		return
	}

	if input.Backspace {
		if input.Ctrl {
			search.Cursor.Column = 0
			search.Input.Reset()

		} else if search.Cursor.Column > 0 {
			search.Cursor.Column -= 1
			str := search.Input.String()
			search.Input.Reset()

			str = str[:len(str)-1]
			search.Input.WriteString(str)
		}

		return
	}

	if input.TypedCharacter != 0 {
		if input.TypedCharacter == '\n' {
			search.Close()
		} else {
			search.Input.WriteByte(input.TypedCharacter)
			search.Cursor.Column += 1
		}

		return
	}
}

func (search *Search) Render(renderer *sdl.Renderer, parentRect *sdl.Rect, theme *FileSearchTheme) {
	inputRect := sdl.Rect{
		X: parentRect.W - search.Width - 20,
		Y: parentRect.Y + parentRect.H - search.LineHeight - 20,
		W: search.Width,
		H: search.LineHeight + 10,
	}

	borderRect := expandRect(inputRect, 1)

	DrawRect(renderer, &borderRect, theme.BorderColor)
	DrawRect(renderer, &inputRect, theme.InputBackgroundColor)
	search.Cursor.Render(renderer, inputRect, theme.CursorColor)

	query := search.Input.String()
	if len(query) > 0 {
		width := search.Font.GetStringWidth(query)
		queryRect := sdl.Rect{
			X: inputRect.X + 5,
			Y: inputRect.Y + (inputRect.H-int32(search.Font.Size))/2,
			W: width,
			H: int32(search.Font.Size),
		}
		DrawText(renderer, search.Font, query, &queryRect, theme.InputTextColor)
	}
}
