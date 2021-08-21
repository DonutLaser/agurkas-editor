package main

import (
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

// @TODO (!important) write tests for this

type Buffer struct {
	Data     []byte
	GapStart int
	GapEnd   int

	Cursor Cursor
}

func CreateBuffer(lineHeight int32) Buffer {
	result := Buffer{
		Data:     make([]byte, 16),
		GapStart: 0,
		GapEnd:   15,
		Cursor: Cursor{
			WidthNormal: 8,
			WidthInsert: 2,
			Height:      lineHeight,
		},
	}

	return result
}

func (buffer *Buffer) Insert(char byte) {
	buffer.Data[buffer.GapStart] = char
	buffer.GapStart += 1

	if char == '\n' {
		buffer.Cursor.Column = 0
		buffer.Cursor.Line += 1
	} else {
		buffer.Cursor.Column += 1
	}

	if buffer.GapEnd-buffer.GapStart == 1 {
		buffer.expand()
	}
}

func (buffer *Buffer) RemoveBefore() {
	if buffer.GapStart == 0 {
		return
	}

	char := buffer.Data[buffer.GapStart-1]
	buffer.Data[buffer.GapStart-1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.GapStart -= 1

	if char == '\n' {
		buffer.Cursor.Line -= 1
		buffer.Cursor.Column = buffer.getCurrentLineSize()
	} else {
		buffer.Cursor.Column -= 1
	}
}

func (buffer *Buffer) RemoveAfter() {
	if buffer.GapEnd == len(buffer.Data)-1 {
		return
	}

	buffer.Data[buffer.GapEnd] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.GapEnd += 1
}

func (buffer *Buffer) MoveLeft() {
	if buffer.GapStart == 0 || buffer.Data[buffer.GapStart-1] == '\n' {
		return
	}

	buffer.moveLeftInternal()
}

func (buffer *Buffer) MoveRight() {
	if buffer.GapEnd == len(buffer.Data)-1 || buffer.Data[buffer.GapEnd+1] == '\n' {
		return
	}

	buffer.moveRightInternal()
}

func (buffer *Buffer) MoveUp() {
	endColumn := buffer.Cursor.Column

	for buffer.Cursor.Column > 0 {
		buffer.moveLeftInternal()
	}

	if buffer.Cursor.Line > 0 {
		buffer.moveLeftInternal() // Move over new line symbol
		buffer.moveLeftInternal() // Move into the previous line to get its size correctly

		// @TODO (!important) do something better here
		buffer.Cursor.Column = buffer.getCurrentLineSize() - 1
		buffer.Cursor.Line -= 1

		for buffer.Cursor.Column > endColumn {
			buffer.moveLeftInternal()
		}
	}
}

func (buffer *Buffer) MoveDown() {
	endColumn := buffer.Cursor.Column
	lineSize := buffer.getCurrentLineSize() // @TODO (!important) this might not be needed, we can just check if the next symbol will be newline

	for buffer.Cursor.Column < lineSize {
		buffer.moveRightInternal()
	}

	if buffer.GapEnd != len(buffer.Data)-1 {
		buffer.moveRightInternal() // Move over the new line symbol

		buffer.Cursor.Column = 0
		buffer.Cursor.Line += 1

		fmt.Println(string(buffer.Data[buffer.GapStart-1]))
		fmt.Println(string(buffer.Data[buffer.GapEnd+1]))

		for buffer.Cursor.Column < endColumn {
			buffer.moveRightInternal()
		}
	}
}

func (buffer *Buffer) GetText() []string {
	// @TODO (!important) it is possible to cache the text lines if the text did not change between frames
	var sb strings.Builder

	for i := 0; i < buffer.GapStart; i += 1 {
		sb.WriteByte(buffer.Data[i])
	}

	for i := buffer.GapEnd + 1; i < len(buffer.Data); i += 1 {
		sb.WriteByte(buffer.Data[i])
	}

	return strings.Split(sb.String(), "\n")
}

func (buffer *Buffer) Render(renderer *sdl.Renderer, font *ttf.Font, mode Mode, characterWidth int32) {
	buffer.Cursor.Render(renderer, mode, characterWidth)

	text := buffer.GetText()
	for index, line := range text {
		if len(line) == 0 {
			continue
		}

		width, height := GetStringSize(font, line)
		rect := sdl.Rect{
			X: 10,
			Y: 10 + int32(index)*height,
			W: width,
			H: height,
		}

		DrawText(renderer, font, line, &rect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}
}

func (buffer *Buffer) expand() {
	newSize := len(buffer.Data) * 2
	newData := make([]byte, newSize)

	for i := 0; i < buffer.GapStart; i += 1 {
		newData[i] = buffer.Data[i]
	}

	postSize := len(buffer.Data) - 1 - buffer.GapEnd
	newGapEnd := len(newData) - postSize - 1
	for i := 0; i < postSize; i += 1 {
		newData[i+newGapEnd+1] = buffer.Data[i+buffer.GapEnd+1]
	}

	buffer.Data = newData
	buffer.GapEnd = newGapEnd
}

func (buffer *Buffer) moveLeftInternal() {
	char := buffer.Data[buffer.GapStart-1]
	buffer.Data[buffer.GapStart-1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.Data[buffer.GapEnd] = char

	buffer.GapStart -= 1
	buffer.GapEnd -= 1

	buffer.Cursor.Column -= 1
}

func (buffer *Buffer) moveRightInternal() {
	char := buffer.Data[buffer.GapEnd+1]
	buffer.Data[buffer.GapEnd+1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.Data[buffer.GapStart] = char

	buffer.GapStart += 1
	buffer.GapEnd += 1

	buffer.Cursor.Column += 1
}

func (buffer *Buffer) getCurrentLineSize() int32 {
	// @TODO (!important) can use one size variable here
	var preSize int32
	preIndex := buffer.GapStart - 1
	for preIndex >= 0 && buffer.Data[preIndex] != '\n' {
		preSize += 1
		preIndex -= 1
	}

	var postSize int32
	postIndex := buffer.GapEnd + 1
	for postIndex != len(buffer.Data) && buffer.Data[postIndex] != '\n' {
		postSize += 1
		postIndex += 1
	}

	return preSize + postSize
}
