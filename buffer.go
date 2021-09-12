package main

import (
	"strconv"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type Direction uint8

const (
	Direction_Up Direction = iota
	Direction_Down
)

type Selection struct {
	Line  int32
	Start int32
	End   int32
}

type SelectionPoint struct {
	Column int32
	Line   int32
	Offset int32
}

type Buffer struct {
	Data                []byte
	GapStart            int
	GapEnd              int
	SelectionStartPoint SelectionPoint
	TotalLines          int

	Font *Font

	Cursor       BufferCursor
	Rect         sdl.Rect
	ScrollY      int32
	ScrollOffset int32
	Dirty        bool

	BookmarkLine  int32
	LineFindQuery byte

	Filepath        string
	HighlighterFunc func(line []byte, theme *SyntaxTheme) []TokenInfo
}

func CreateBuffer(lineHeight int32, font *Font, rect sdl.Rect) (result Buffer) {
	result.Data = make([]byte, 16)
	result.GapStart = 0
	result.GapEnd = 15
	result.SelectionStartPoint = SelectionPoint{Column: -1, Line: -1, Offset: 0}
	result.TotalLines = 1

	result.Font = font

	result.Cursor = CreateBufferCursor(lineHeight, int32(font.CharacterWidth))
	result.Rect = rect
	result.ScrollY = 0
	result.ScrollOffset = 8 // Line count
	result.Dirty = false

	result.BookmarkLine = 0
	result.LineFindQuery = 0

	result.Filepath = ""
	result.HighlighterFunc = nil

	return
}

// =============================================================
// PUBLIC
// =============================================================

func (buffer *Buffer) SetData(data []byte, filepath string) {
	cleaned := cleanText(data)

	buffer.Data = make([]byte, len(cleaned)+16) // 16 symbols for the gap
	buffer.GapStart = 0
	buffer.GapEnd = 15
	buffer.BookmarkLine = 0
	buffer.Filepath = filepath
	buffer.Dirty = false
	buffer.Cursor.Column = 0
	buffer.Cursor.Line = 0
	buffer.ScrollY = 0

	for i := 16; i < len(buffer.Data); i += 1 {
		buffer.Data[i] = cleaned[i-16]
	}

	text, _ := buffer.GetText()
	buffer.TotalLines = len(text)

	buffer.HighlighterFunc = nil
	if strings.HasSuffix(buffer.Filepath, ".go") {
		buffer.HighlighterFunc = HighlightLineGolang
	} else if strings.HasSuffix(buffer.Filepath, ".atheme") {
		buffer.HighlighterFunc = HighlightLineTheme
	}
}

func (buffer *Buffer) StartSelection() {
	buffer.SelectionStartPoint.Column = buffer.Cursor.Column
	buffer.SelectionStartPoint.Line = buffer.Cursor.Line
	buffer.SelectionStartPoint.Offset = int32(buffer.GapStart)
}

func (buffer *Buffer) StopSelection() {
	buffer.SelectionStartPoint.Column = -1
	buffer.SelectionStartPoint.Line = -1
	buffer.SelectionStartPoint.Offset = 0
}

func (buffer *Buffer) Insert(char byte) {
	prevChar := buffer.prevCharacter()
	nextChar := buffer.nextCharacter()

	if char == '\t' {
		// @TODO (!important) write tests for this
		count := 4 - buffer.Cursor.Column%4
		// @TODO (!important) temporary, should correctly handle tabs
		for i := 0; i < int(count); i += 1 {
			buffer.Insert(' ')
		}

		return
	} else {
		buffer.Data[buffer.GapStart] = char
		buffer.GapStart += 1
		buffer.Cursor.Column += 1

		if char == '\n' {
			buffer.Cursor.Column = 0
			buffer.Cursor.Line += 1
			buffer.TotalLines += 1

			pair := getSymbolPair(prevChar)
			if pair != 0 && pair != '"' && pair != '\'' && nextChar == pair {
				buffer.Data[buffer.GapStart] = char
				buffer.GapStart += 1

				buffer.Cursor.Line += 1
				buffer.TotalLines += 1

				buffer.MoveUp()
			}

			buffer.maybeScrollDown()
		} else {
			pair := getSymbolPair(char)
			if pair != 0 {
				buffer.Insert(pair)
				buffer.MoveLeft()
			}
		}
	}

	if buffer.GapEnd-buffer.GapStart == 1 {
		buffer.expand()
	}

	buffer.Dirty = true
}

// @TODO (!important) write tests for this
func (buffer *Buffer) ReplaceCurrentCharacter(char byte) {
	if char == '\n' {
		return
	}

	buffer.Data[buffer.GapEnd+1] = char

	buffer.Dirty = true
}

func (buffer *Buffer) RemoveBefore() {
	if buffer.GapStart == 0 {
		return
	}

	char := buffer.prevCharacter()
	nextChar := buffer.nextCharacter()

	buffer.Data[buffer.GapStart-1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.GapStart -= 1

	if char == '\n' {
		buffer.Cursor.Line -= 1
		buffer.Cursor.Column = buffer.currentLineSize()
		buffer.TotalLines -= 1
	} else {
		buffer.Cursor.Column -= 1

		pair := getSymbolPair(char)
		if pair != 0 && pair == nextChar {
			buffer.RemoveAfter()
		}
	}

	buffer.Dirty = true
}

func (buffer *Buffer) RemoveAfter() {
	if buffer.GapEnd == len(buffer.Data)-1 {
		return
	}

	if buffer.nextCharacter() == '\n' {
		buffer.TotalLines -= 1
	}

	buffer.Data[buffer.GapEnd] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.GapEnd += 1

	buffer.Dirty = true
}

// @TODO (!important) write tests for this
func (buffer *Buffer) RemoveCurrentLine() {
	for buffer.GapEnd != len(buffer.Data)-1 && buffer.nextCharacter() != '\n' {
		buffer.RemoveAfter()
	}
	buffer.RemoveAfter() // Remove new line

	for buffer.Cursor.Column > 0 {
		buffer.RemoveBefore()
	}

	// If we are removing the last line, remove it completely and jump to the next llne
	if buffer.GapEnd == len(buffer.Data)-1 {
		buffer.RemoveBefore()
	}

	buffer.Dirty = true
}

func (buffer *Buffer) ChangeCurrentLine() {
	for buffer.GapEnd != len(buffer.Data)-1 && buffer.nextCharacter() != '\n' {
		buffer.RemoveAfter()
	}

	for buffer.Cursor.Column > 0 {
		buffer.RemoveBefore()
	}

	buffer.Dirty = true
}

func (buffer *Buffer) MoveLeft() {
	if buffer.GapStart == 0 || buffer.prevCharacter() == '\n' {
		return
	}

	buffer.moveLeftInternal()
}

func (buffer *Buffer) MoveRight() {
	if buffer.GapEnd == len(buffer.Data)-1 || buffer.nextCharacter() == '\n' {
		return
	}

	buffer.moveRightInternal()
}

func (buffer *Buffer) MoveUp() {
	endColumn := int32(Max(int(buffer.Cursor.Column), int(buffer.Cursor.LastColumn)))

	if buffer.Cursor.Line == 0 {
		return
	}

	for buffer.Cursor.Column > 0 {
		buffer.moveLeftInternal()
	}

	if buffer.Cursor.Line > 0 {
		buffer.moveLeftInternal() // Move over new line symbol
		// @TODO (!important) this is probably not needed, get rid of this
		char := buffer.prevCharacter()
		if char != 0 && char != '\n' {
			buffer.moveLeftInternal() // Move into the previous line to get its size correctly, unless the line is empty
		}

		// @TODO (!important) do something better here
		buffer.Cursor.Column = int32(Max(int(buffer.currentLineSize()-1), 0))
		buffer.Cursor.Line -= 1

		for buffer.Cursor.Column > endColumn {
			buffer.moveLeftInternal()
		}

		buffer.Cursor.LastColumn = endColumn

		buffer.maybeScrollUp()
	}
}

func (buffer *Buffer) MoveDown() {
	if buffer.Cursor.Line == int32(buffer.TotalLines)-1 {
		return
	}

	endColumn := Max(int(buffer.Cursor.Column), int(buffer.Cursor.LastColumn))
	lineSize := buffer.currentLineSize() // @TODO (!important) this might not be needed, we can just check if the next symbol will be newline

	for buffer.Cursor.Column < lineSize {
		buffer.moveRightInternal()
	}

	if buffer.GapEnd != len(buffer.Data)-1 {
		buffer.moveRightInternal() // Move over the new line symbol

		buffer.Cursor.Column = 0
		buffer.Cursor.Line += 1

		// @TODO (!important) when the next line is shorter than the current column, this will unnecessarily try moving right
		for i := 0; i < int(endColumn); i += 1 {
			buffer.MoveRight()
		}

		buffer.Cursor.LastColumn = int32(endColumn)

		buffer.maybeScrollDown()
	}
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveToBookmark() {
	if buffer.BookmarkLine < buffer.Cursor.Line {
		for buffer.Cursor.Line != buffer.BookmarkLine {
			buffer.MoveUp()
		}

		return
	}

	for buffer.Cursor.Line != buffer.BookmarkLine {
		buffer.MoveDown()
	}
}

func (buffer *Buffer) MarkCurrentPosition() {
	buffer.BookmarkLine = buffer.Cursor.Line
}

func (buffer *Buffer) GetText() (lines []string, selection []Selection) {
	// @TODO (!important) it is possible to cache the text lines if the text did not change between frames
	var sb strings.Builder

	for i := 0; i < buffer.GapStart; i += 1 {
		sb.WriteByte(buffer.Data[i])
	}

	for i := buffer.GapEnd + 1; i < len(buffer.Data); i += 1 {
		sb.WriteByte(buffer.Data[i])
	}

	lines = strings.Split(sb.String(), "\n")

	if buffer.SelectionStartPoint.Column > -1 {
		if buffer.Cursor.Line > buffer.SelectionStartPoint.Line {
			selection = append(selection, Selection{Line: buffer.SelectionStartPoint.Line, Start: buffer.SelectionStartPoint.Column, End: int32(len(lines[buffer.SelectionStartPoint.Line]))})
			for i := buffer.SelectionStartPoint.Line + 1; i < buffer.Cursor.Line; i += 1 {
				selection = append(selection, Selection{Line: i, Start: 0, End: int32(len(lines[i]))})
			}
			selection = append(selection, Selection{Line: buffer.Cursor.Line, Start: 0, End: buffer.Cursor.Column})
		} else if buffer.Cursor.Line == buffer.SelectionStartPoint.Line {
			selection = append(selection, Selection{Line: buffer.Cursor.Line, Start: buffer.SelectionStartPoint.Column, End: buffer.Cursor.Column})
		} else {
			selection = append(selection, Selection{Line: buffer.Cursor.Line, Start: buffer.Cursor.Column, End: int32(len(lines[buffer.Cursor.Line]))})
			for i := buffer.Cursor.Line + 1; i < buffer.SelectionStartPoint.Line; i += 1 {
				selection = append(selection, Selection{Line: i, Start: 0, End: int32(len(lines[i]))})
			}
			selection = append(selection, Selection{Line: buffer.SelectionStartPoint.Line, Start: 0, End: buffer.SelectionStartPoint.Column + 1})
		}
	}

	return
}

func (buffer *Buffer) Render(renderer *sdl.Renderer, mode Mode, theme *Theme) {
	gutterRect := sdl.Rect{
		X: 0,
		Y: 0,
		W: 48,
		H: buffer.Rect.H,
	}
	DrawRect(renderer, &gutterRect, theme.Gutter.BackgroundColor)

	text, selection := buffer.GetText()

	buffer.Cursor.Render(renderer, mode, gutterRect.W, buffer.Rect.W, buffer.ScrollY)
	buffer.renderSelection(renderer, gutterRect.W+5, selection, theme.Buffer.SelectionColor)

	for index, line := range text {
		y := int32(index)*buffer.Cursor.Height + (buffer.Cursor.Height-int32(buffer.Font.Size))/2 + buffer.ScrollY

		if y > buffer.Rect.Y+buffer.Rect.H || y+int32(buffer.Font.Size) < buffer.Rect.Y {
			continue
		}

		buffer.renderLineNumber(renderer, &gutterRect, index, theme)

		if len(line) == 0 {
			continue
		}

		lineWidth := buffer.Font.GetStringWidth(line)
		x := gutterRect.W + 5

		if buffer.HighlighterFunc != nil {
			buffer.renderLine(renderer, line, x, y, &theme.Syntax)
		} else {
			rect := sdl.Rect{ // @TODO (!important) rect could be reused between iterations to decrease garbage produced by the loop
				X: x,
				Y: y,
				W: lineWidth,
				H: int32(buffer.Font.Size),
			}

			DrawText(renderer, buffer.Font, line, &rect, theme.Buffer.TextColor)
		}
	}
}

// =============================================================
// PRIVATE
// =============================================================

func (buffer *Buffer) renderLineNumber(renderer *sdl.Renderer, gutterRect *sdl.Rect, index int, theme *Theme) {
	lineNumber := Abs(int(buffer.Cursor.Line) - index)
	lineNumberColor := theme.Gutter.LineNumberInactiveColor
	lineNumberOffset := 0
	if lineNumber == 0 {
		lineNumber = index + 1
		lineNumberOffset = 10
		if theme.Gutter.LineNumberMatchModeColor {
			lineNumberColor = buffer.Cursor.Color
		} else {
			lineNumberColor = theme.Gutter.LineNumberActiveColor
		}

		numberHighlightRect := sdl.Rect{
			X: gutterRect.X,
			Y: buffer.Cursor.Line*buffer.Cursor.Height + buffer.ScrollY,
			W: gutterRect.W,
			H: buffer.Cursor.Height,
		}
		DrawRect(renderer, &numberHighlightRect, theme.Gutter.LineHighlightColor)
	}

	lineNumberStr := strconv.Itoa(lineNumber)
	width := buffer.Font.GetStringWidth(lineNumberStr)
	// @TODO (!important) rect could be reused between iterations to decrease garbage produced by the loop
	lineNumberRect := sdl.Rect{
		X: gutterRect.X + gutterRect.W - 10 - width - int32(lineNumberOffset),
		Y: int32(index)*buffer.Cursor.Height + (buffer.Cursor.Height-int32(buffer.Font.Size))/2 + buffer.ScrollY,
		W: width,
		H: int32(buffer.Font.Size),
	}
	DrawText(renderer, buffer.Font, lineNumberStr, &lineNumberRect, lineNumberColor)
}

func (buffer *Buffer) renderLine(renderer *sdl.Renderer, line string, leftStart int32, y int32, theme *SyntaxTheme) {
	tokens := buffer.HighlighterFunc([]byte(line), theme)

	left := leftStart

	for _, token := range tokens {
		width := buffer.Font.GetStringWidth(token.Value)
		rect := sdl.Rect{
			X: left,
			Y: y,
			W: width,
			H: int32(buffer.Font.Size),
		}
		DrawText(renderer, buffer.Font, token.Value, &rect, token.Color)

		left += width
	}
}

func (buffer *Buffer) renderSelection(renderer *sdl.Renderer, left int32, selection []Selection, color sdl.Color) {
	for _, sel := range selection {
		rect := sdl.Rect{
			X: left + sel.Start*int32(buffer.Font.CharacterWidth),
			Y: int32(sel.Line)*buffer.Cursor.Height + buffer.ScrollY,
			W: (sel.End - sel.Start) * int32(buffer.Font.CharacterWidth),
			H: buffer.Cursor.Height,
		}
		DrawRect(renderer, &rect, color)
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
	char := buffer.prevCharacter()
	buffer.Data[buffer.GapStart-1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.Data[buffer.GapEnd] = char

	buffer.GapStart -= 1
	buffer.GapEnd -= 1

	buffer.Cursor.Column -= 1
	buffer.Cursor.LastColumn = 0
}

func (buffer *Buffer) moveRightInternal() {
	char := buffer.nextCharacter()
	buffer.Data[buffer.GapEnd+1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.Data[buffer.GapStart] = char

	buffer.GapStart += 1
	buffer.GapEnd += 1

	buffer.Cursor.Column += 1
	buffer.Cursor.LastColumn = 0
}

func (buffer *Buffer) currentLineSize() (result int32) {
	preIndex := buffer.GapStart - 1
	for preIndex >= 0 && buffer.Data[preIndex] != '\n' {
		result += 1
		preIndex -= 1
	}

	postIndex := buffer.GapEnd + 1
	for postIndex != len(buffer.Data) && buffer.Data[postIndex] != '\n' {
		result += 1
		postIndex += 1
	}

	return result
}

func cleanText(data []byte) (result []byte) {
	for _, b := range data {
		if b == 13 { // Remove \r symbol
			continue
		}

		// @TODO (!important) temporary, should correctly handle tabs
		if b == 9 { // Turn tabs into spaces
			for i := 0; i < 4; i += 1 {
				result = append(result, 32)
			}

			continue
		}

		result = append(result, b)
	}

	return
}

func (buffer *Buffer) prevCharacter() byte {
	if buffer.GapStart == 0 {
		return 0
	}

	return buffer.Data[buffer.GapStart-1]
}

func (buffer *Buffer) nextCharacter() byte {
	if buffer.GapEnd == len(buffer.Data)-1 {
		return 0
	}

	return buffer.Data[buffer.GapEnd+1]
}

func (buffer *Buffer) maybeScrollDown() {
	cursorBottom := buffer.Cursor.GetBottom() + buffer.ScrollY
	diff := cursorBottom - (buffer.Rect.Y + buffer.Rect.H - buffer.ScrollOffset*buffer.Cursor.Height)
	if diff > 0 {
		buffer.ScrollY -= diff
	}
}

func (buffer *Buffer) maybeScrollUp() {
	cursorTop := buffer.Cursor.GetTop() + buffer.ScrollY
	diff := cursorTop - (buffer.Rect.Y + buffer.ScrollOffset*buffer.Cursor.Height)
	if diff < 0 {
		buffer.ScrollY = int32(Min(int(buffer.ScrollY-diff), 0))
	}
}
