package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type CommandPalette struct {
	Input strings.Builder

	Cursor      InputCursor
	Width       int32
	LineHeight  int32
	LineSpacing int32
	Font        *Font

	CloseCallback func(string)
}

func CreateCommandPalette(lineHeight int32, font *Font) (result CommandPalette) {
	result.Cursor = CreateInputCursor(lineHeight, int32(font.CharacterWidth))
	result.Width = 500
	result.LineHeight = lineHeight
	result.LineSpacing = (lineHeight - int32(font.Size)) / 2
	result.Font = font

	return
}

func (cp *CommandPalette) Open(onClose func(string)) {
	cp.Cursor.Column = 0
	cp.Input.Reset()

	cp.CloseCallback = onClose
}

func (cp *CommandPalette) Close() {
	cp.CloseCallback("")
}

func (cp *CommandPalette) Submit() {
	cp.CloseCallback(cp.Input.String())
}

func (cp *CommandPalette) Tick(input Input) {
	if input.Escape {
		cp.Close()
		return
	}

	if input.Backspace {
		if input.Ctrl {
			cp.Cursor.Column = 0
			cp.Input.Reset()
		} else if cp.Cursor.Column > 0 {
			cp.Cursor.Column -= 1
			str := cp.Input.String()
			cp.Input.Reset()

			str = str[:len(str)-1]
			cp.Input.WriteString(str)
		}

		return
	}

	if input.TypedCharacter != 0 {
		if input.TypedCharacter == '\n' {
			cp.Submit()
		} else {
			cp.Input.WriteByte(input.TypedCharacter)
			cp.Cursor.Column += 1
		}

		return
	}
}

func (cp *CommandPalette) Render(renderer *sdl.Renderer, parentRect *sdl.Rect, theme *FileSearchTheme) {
	inputRect := sdl.Rect{
		X: parentRect.W/2 - cp.Width/2,
		Y: parentRect.Y + int32(float32(parentRect.H)*0.15),
		W: cp.Width,
		H: cp.LineHeight + 10,
	}

	borderRect := expandRect(inputRect, 1)

	DrawRect(renderer, &borderRect, theme.BorderColor)
	DrawRect(renderer, &inputRect, theme.InputBackgroundColor)
	cp.Cursor.Render(renderer, inputRect, theme.CursorColor)

	command := cp.Input.String()
	color := theme.InputTextColor
	if len(command) == 0 {
		command = "Command"
		color = theme.ResultPathColor
	}

	width := cp.Font.GetStringWidth(command)
	commandRect := sdl.Rect{
		X: inputRect.X + 5,
		Y: inputRect.Y + (inputRect.H-int32(cp.Font.Size))/2,
		W: width,
		H: int32(cp.Font.Size),
	}
	DrawText(renderer, cp.Font, command, &commandRect, color)
}
