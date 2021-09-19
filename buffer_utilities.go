package main

import "fmt"

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
	char := buffer.nextCharacter()
	for char != '\n' && buffer.GapEnd != len(buffer.Data)-1 {
		buffer.RemoveAfter()
		char = buffer.nextCharacter()
	}

	buffer.Dirty = true
}

func (buffer *Buffer) RemoveSelection() {
	start, end := buffer.sortSelectionEnds(buffer.SelectionStartPoint, buffer.cursorToCursorPoint())

	if start.Column == buffer.Cursor.Column && start.Line == buffer.Cursor.Line {
		for buffer.GapEnd != int(end.OffsetRight) {
			buffer.RemoveAfter()
		}
	} else {
		for buffer.GapStart != int(start.OffsetLeft) {
			buffer.RemoveBefore()
		}
	}

	buffer.RemoveAfter() // Remove symbol under the cursor
}

func (buffer *Buffer) ChangeRemainingLine() {
	buffer.RemoveRemainingLine()
	buffer.Dirty = true
}

func (buffer *Buffer) IndentSelection() {
	start, end := buffer.sortSelectionEnds(buffer.SelectionStartPoint, buffer.cursorToCursorPoint())

	if buffer.Cursor.Line == start.Line {
		for buffer.Cursor.Line != end.Line {
			buffer.Indent()
			buffer.MoveDown()
		}
	} else {
		for buffer.Cursor.Line != start.Line {
			buffer.Indent()
			buffer.MoveUp()
		}
	}

	buffer.Indent() // Indent the last line
}

func (buffer *Buffer) OutdentSelection() {
	start, end := buffer.sortSelectionEnds(buffer.SelectionStartPoint, buffer.cursorToCursorPoint())

	if buffer.Cursor.Line == start.Line {
		for buffer.Cursor.Line != end.Line {
			buffer.Outdent()
			buffer.MoveDown()
		}
	} else {
		for buffer.Cursor.Line != start.Line {
			buffer.Outdent()
			buffer.MoveUp()
		}
	}

	buffer.Outdent() // Outdent the last line
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

func (buffer *Buffer) MoveToLine(line int32) {
	// Line - 1 because line starts at 1, but cursor line starts at 0
	if buffer.Cursor.Line > line-1 {
		for buffer.Cursor.Line > line-1 {
			buffer.MoveUp()
		}
	} else {
		if line-1 > int32(buffer.TotalLines) {
			line = int32(buffer.TotalLines)
		}

		fmt.Println(line)

		for buffer.Cursor.Line < line-1 {
			buffer.MoveDown()
		}
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
	for buffer.GapEnd != len(buffer.Data)-1 && buffer.nextCharacter() != '\n' {
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
func (buffer *Buffer) MoveRightToWordStart(ignorePunctuation bool) {
	if !ignorePunctuation {
		for !isPunctuation(buffer.nextCharacter()) && buffer.GapEnd != len(buffer.Data)-1 {
			buffer.MoveRight()
		}
	} else {
		for !isWhitespace(buffer.nextCharacter()) && buffer.GapEnd != len(buffer.Data)-1 {
			buffer.MoveRight()
		}
	}

	buffer.MoveRight()
}

// @TODO (!important) write tests for this
func (buffer *Buffer) MoveLeftToWordStart(ignorePunctuation bool) {
	buffer.MoveLeft()

	if !ignorePunctuation {
		for !isPunctuation(buffer.prevCharacter()) && buffer.Cursor.Column > 0 {
			buffer.MoveLeft()
		}
	} else {
		for !isWhitespace(buffer.prevCharacter()) && buffer.Cursor.Column > 0 {
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

	nextChar := buffer.nextCharacter()
	for nextChar != '\n' && nextChar != buffer.LineFindQuery && buffer.GapEnd != len(buffer.Data)-1 {
		buffer.MoveRight()
		nextChar = buffer.nextCharacter()
	}
}

func (buffer *Buffer) MoveToPrevLineQuerySymbol() {
	buffer.MoveLeft() // Ignore the query symbol that we are currently standing on

	nextChar := buffer.nextCharacter()
	for nextChar != buffer.LineFindQuery && buffer.Cursor.Column != 0 {
		buffer.MoveLeft()
		nextChar = buffer.nextCharacter()
	}
}
