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

type Buffer struct {
	Data       []byte
	GapStart   int
	GapEnd     int
	TotalLines int

	Font        *Font
	LineSpacing int32

	Cursor       BufferCursor
	Rect         sdl.Rect
	ScrollY      int32
	ScrollOffset int32

	BookmarkLine  int32
	LineFindQuery byte

	Filepath        string
	HighlighterFunc func(line []byte, theme *SyntaxTheme) []TokenInfo
	Dirty           bool
}

func CreateBuffer(lineHeight int32, font *Font, rect sdl.Rect) Buffer {
	result := Buffer{
		Data:        make([]byte, 16),
		GapStart:    0,
		GapEnd:      15,
		TotalLines:  1,
		Font:        font,
		LineSpacing: (lineHeight - int32(font.Size)) / 2,
		Cursor: BufferCursor{
			WidthNormal: 8,
			WidthInsert: 2,
			Height:      lineHeight,
			Advance:     int32(font.CharacterWidth),
		},
		Rect:         rect,
		ScrollOffset: 8, // Lines
		BookmarkLine: 0,
		Filepath:     "",
		Dirty:        false,
	}

	return result
}

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

	buffer.TotalLines = len(buffer.GetText())

	buffer.HighlighterFunc = nil
	if strings.HasSuffix(buffer.Filepath, ".go") {
		buffer.HighlighterFunc = HighlightLineGolang
	} else if strings.HasSuffix(buffer.Filepath, ".atheme") {
		buffer.HighlighterFunc = HighlightLineTheme
	}
}

func (buffer *Buffer) Insert(char byte) {
	prevChar := buffer.getPrevCharacter()
	nextChar := buffer.getNextCharacter()

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

// @TODO (!important) write tests for this
func (buffer *Buffer) InsertNewLineBelow() {
	buffer.MoveToEndOfLine()
	buffer.Insert('\n')

	buffer.Dirty = true
}

// @TODO (!important) write tests for this
func (buffer *Buffer) InsertNewLineAbove() {
	buffer.MoveToStartOfLine()
	buffer.Insert('\n')
	buffer.MoveUp()

	buffer.Dirty = true
}

func (buffer *Buffer) RemoveBefore() {
	if buffer.GapStart == 0 {
		return
	}

	char := buffer.getPrevCharacter()
	buffer.Data[buffer.GapStart-1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.GapStart -= 1

	if char == '\n' {
		buffer.Cursor.Line -= 1
		buffer.Cursor.Column = buffer.getCurrentLineSize()
		buffer.TotalLines -= 1
	} else {
		buffer.Cursor.Column -= 1
	}

	buffer.Dirty = true
}

func (buffer *Buffer) RemoveAfter() {
	if buffer.GapEnd == len(buffer.Data)-1 {
		return
	}

	if buffer.getNextCharacter() == '\n' {
		buffer.TotalLines -= 1
	}

	buffer.Data[buffer.GapEnd] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.GapEnd += 1

	buffer.Dirty = true
}

// @TODO (!important) write tests for this
func (buffer *Buffer) RemoveCurrentLine() {
	for buffer.GapEnd != len(buffer.Data)-1 && buffer.getNextCharacter() != '\n' {
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

// @TODO (!important) write tests for thi
func (buffer *Buffer) RemoveLines(direction Direction, count int) {
	if direction == Direction_Up {
		buffer.RemoveCurrentLine()

		for i := 0; i < count; i += 1 {
			if buffer.Cursor.Line == 0 {
				break
			}

			buffer.MoveUp()
			buffer.RemoveCurrentLine()
		}

		return
	}

	buffer.RemoveCurrentLine()
	for i := 0; i < count; i += 1 {
		buffer.RemoveCurrentLine()
	}

	buffer.Dirty = true
}

// @TODO (!important) write tests for this
func (buffer *Buffer) RemoveRemainingLine() {
	char := buffer.getNextCharacter()
	for char != '\n' && buffer.GapEnd != len(buffer.Data)-1 {
		buffer.RemoveAfter()
		char = buffer.getNextCharacter()
	}

	buffer.Dirty = true
}

func (buffer *Buffer) ChangeCurrentLine() {
	for buffer.GapEnd != len(buffer.Data)-1 && buffer.getNextCharacter() != '\n' {
		buffer.RemoveAfter()
	}

	for buffer.Cursor.Column > 0 {
		buffer.RemoveBefore()
	}

	buffer.Dirty = true
}

func (buffer *Buffer) ChangeRemainingLine() {
	buffer.RemoveRemainingLine()
	buffer.Dirty = true
}

func (buffer *Buffer) MoveLeft() {
	if buffer.GapStart == 0 || buffer.getPrevCharacter() == '\n' {
		return
	}

	buffer.moveLeftInternal()
}

func (buffer *Buffer) MoveRight() {
	if buffer.GapEnd == len(buffer.Data)-1 || buffer.getNextCharacter() == '\n' {
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
		char := buffer.getPrevCharacter()
		if char != 0 && char != '\n' {
			buffer.moveLeftInternal() // Move into the previous line to get its size correctly, unless the line is empty
		}

		// @TODO (!important) do something better here
		buffer.Cursor.Column = int32(Max(int(buffer.getCurrentLineSize()-1), 0))
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
	lineSize := buffer.getCurrentLineSize() // @TODO (!important) this might not be needed, we can just check if the next symbol will be newline

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

func (buffer *Buffer) MoveUpByLines(lines int) {
	for i := 0; i < lines; i += 1 {
		buffer.MoveUp()
	}
}

func (buffer *Buffer) MoveDownByLines(lines int) {
	for i := 0; i < lines; i += 1 {
		buffer.MoveDown()
	}
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveToStartOfLine() {
	for buffer.Cursor.Column > 0 {
		buffer.MoveLeft()
	}
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveToEndOfLine() {
	for buffer.GapEnd != len(buffer.Data)-1 && buffer.getNextCharacter() != '\n' {
		buffer.MoveRight()
	}
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveToBufferStart() {
	for buffer.Cursor.Line > 0 {
		buffer.MoveUp()
	}
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveToBufferEnd() {
	for buffer.Cursor.Line < int32(buffer.TotalLines)-1 {
		buffer.MoveDown()
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

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveRightToWordStart(ignorePunctuation bool) {
	if !ignorePunctuation {
		for !isPunctuation(buffer.getNextCharacter()) && buffer.GapEnd != len(buffer.Data)-1 {
			buffer.MoveRight()
		}
	} else {
		for !isWhitespace(buffer.getNextCharacter()) && buffer.GapEnd != len(buffer.Data)-1 {
			buffer.MoveRight()
		}
	}

	buffer.MoveRight()
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveLeftToWordStart(ignorePunctuation bool) {
	buffer.MoveLeft()

	if !ignorePunctuation {
		for !isPunctuation(buffer.getPrevCharacter()) && buffer.Cursor.Column > 0 {
			buffer.MoveLeft()
		}
	} else {
		for !isWhitespace(buffer.getPrevCharacter()) && buffer.Cursor.Column > 0 {
			buffer.MoveLeft()
		}
	}
}

// @TODO (!important) write tests for this
// @TODO (!important) improve: remove all white from the next line as well
func (buffer *Buffer) MergeLineBelow() {
	buffer.MoveToEndOfLine()

	if buffer.GapEnd == len(buffer.Data)-1 {
		return
	}

	// We are guaranteed to be removing a new line here
	buffer.RemoveAfter()
	buffer.Insert(' ')
	buffer.MoveLeft()

	buffer.Dirty = true
}

func (buffer *Buffer) MarkCurrentPosition() {
	buffer.BookmarkLine = buffer.Cursor.Line
}

func (buffer *Buffer) FindInLine(symbol byte, forwards bool) {
	buffer.LineFindQuery = symbol

	if forwards {
		buffer.MoveToNextLineQuerySymbol()
	} else {
		buffer.MoveToPrevLineQuerySymbol()
	}
}

func (buffer *Buffer) MoveToNextLineQuerySymbol() {
	buffer.MoveRight() // Ignore the query symbol that we are currently standing on

	nextChar := buffer.getNextCharacter()
	for nextChar != '\n' && nextChar != buffer.LineFindQuery && buffer.GapEnd != len(buffer.Data)-1 {
		buffer.MoveRight()
		nextChar = buffer.getNextCharacter()
	}
}

func (buffer *Buffer) MoveToPrevLineQuerySymbol() {
	buffer.MoveLeft() // Ignore the query symbol that we are currently standing on

	nextChar := buffer.getNextCharacter()
	for nextChar != buffer.LineFindQuery && buffer.Cursor.Column != 0 {
		buffer.MoveLeft()
		nextChar = buffer.getNextCharacter()
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

func (buffer *Buffer) Render(renderer *sdl.Renderer, mode Mode, theme *Theme) {
	gutterRect := sdl.Rect{
		X: 0,
		Y: 0,
		W: 48,
		H: buffer.Rect.H,
	}
	DrawRect(renderer, &gutterRect, theme.Gutter.BackgroundColor)

	var cursorColor sdl.Color
	if theme.Buffer.CursorColorMatchModeColor {
		cursorColor = theme.GetColorForMode(mode)
	} else {
		cursorColor = theme.Buffer.CursorColor
	}

	buffer.Cursor.Render(renderer, mode, cursorColor, gutterRect.W, buffer.Rect.W, buffer.ScrollY)

	text := buffer.GetText()
	for index, line := range text {
		y := int32(index)*buffer.LineSpacing + int32(index)*int32(buffer.Font.Size) + int32(index+1)*buffer.LineSpacing + buffer.ScrollY

		if y > buffer.Rect.Y+buffer.Rect.H || y+int32(buffer.Font.Size) < buffer.Rect.Y {
			continue
		}

		lineNumber := Abs(int(buffer.Cursor.Line) - index)
		lineNumberColor := theme.Gutter.LineNumberInactiveColor
		lineNumberOffset := 0
		if lineNumber == 0 {
			lineNumber = index + 1
			lineNumberOffset = 10
			if theme.Gutter.LineNumberMatchModeColor {
				lineNumberColor = cursorColor
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
			Y: int32(index)*buffer.LineSpacing + int32(index)*int32(buffer.Font.Size) + int32(index+1)*buffer.LineSpacing + buffer.ScrollY,
			W: width,
			H: int32(buffer.Font.Size),
		}
		DrawText(renderer, buffer.Font, lineNumberStr, &lineNumberRect, lineNumberColor)

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
	char := buffer.getPrevCharacter()
	buffer.Data[buffer.GapStart-1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.Data[buffer.GapEnd] = char

	buffer.GapStart -= 1
	buffer.GapEnd -= 1

	buffer.Cursor.Column -= 1
	buffer.Cursor.LastColumn = 0
}

func (buffer *Buffer) moveRightInternal() {
	char := buffer.getNextCharacter()
	buffer.Data[buffer.GapEnd+1] = '_' // @TODO (!important) only useful for debug, remove when buffer implementation is stable
	buffer.Data[buffer.GapStart] = char

	buffer.GapStart += 1
	buffer.GapEnd += 1

	buffer.Cursor.Column += 1
	buffer.Cursor.LastColumn = 0
}

func (buffer *Buffer) getCurrentLineSize() (result int32) {
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

func (buffer *Buffer) getPrevCharacter() byte {
	if buffer.GapStart == 0 {
		return 0
	}

	return buffer.Data[buffer.GapStart-1]
}

func (buffer *Buffer) getNextCharacter() byte {
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
