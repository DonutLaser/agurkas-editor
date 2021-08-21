package main

import (
	"fmt"
	"testing"
)

func TestCreateBuffer(t *testing.T) {
	result := CreateBuffer(16)

	FailIfFalse(len(result.Data) == 16, "Incorrect buffer size", t)
	FailIfFalse(result.GapStart == 0, "Incorrect gap start", t)
	FailIfFalse(result.GapEnd == 15, "Incorrect gap end", t)
	FailIfFalse(result.Cursor.Column == 0, "Incorrect cursor column", t)
	FailIfFalse(result.Cursor.Line == 0, "Incorrect cursor line", t)
	FailIfFalse(result.Cursor.Height == 16, "Incorrect cursor height", t)
}

func TestGetText(t *testing.T) {
	buffer := CreateBuffer(16)

	// Setting up buffer data to be abcd\ne_________o which should procude two lines: 'abcd' and 'ei'
	bytes := []byte("abcd\nefghijklmno")
	buffer.Data = bytes
	buffer.GapStart = 6
	buffer.GapEnd = 14

	result := buffer.GetText()

	expected := []string{"abcd", "eo"}

	FailNowIfFalse(len(result) == len(expected), "Incorrect line count received", t)

	for index, line := range result {
		FailIfFalse(line == expected[index], fmt.Sprintf("Expected line %d to be %s, got %s", index, expected[index], line), t)
	}
}

func TestInsert(t *testing.T) {
	t.Run("Insert at the end of the buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('g')
		buffer.Insert('u')
		buffer.Insert('r')
		buffer.Insert('k')
		buffer.Insert('a')
		buffer.Insert('s')

		FailIfFalse(buffer.Cursor.Column == 7, "Cursor column is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Expected to get only 1 line of text", t)
		FailIfFalse(result[0] == "agurkas", "Incorrect resulting text", t)
	})

	t.Run("Insert new line at the end of the buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.Insert('\n')
		buffer.Insert('e')
		buffer.Insert('f')
		buffer.Insert('g')
		buffer.Insert('h')

		FailIfFalse(buffer.Cursor.Line == 1, "Cursor line is not where it should be", t)

		result := buffer.GetText()
		expected := []string{"abcd", "efgh"}
		FailNowIfFalse(len(result) == len(expected), "Incorrect line count received", t)

		for index, line := range result {
			FailIfFalse(line == expected[index], fmt.Sprintf("Expected line %d to be %s, got %s", index, expected[index], line), t)
		}
	})

	t.Run("Insert in the middle of the text", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.Insert('x')

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abxcd", "Incorrect resulting text", t)
	})

	t.Run("Insert new line in the middle of the text", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.Insert('\n')

		result := buffer.GetText()
		expected := []string{"ab", "cd"}
		FailNowIfFalse(len(result) == len(expected), "Incorrect line count received", t)

		for index, line := range result {
			FailIfFalse(line == expected[index], fmt.Sprintf("Expected line %d to be %s, got %s", index, expected[index], line), t)
		}
	})

	t.Run("Insert new line at the beginning of the text", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.Insert('\n')

		result := buffer.GetText()
		expected := []string{"", "abcd"}
		FailNowIfFalse(len(result) == len(expected), "Incorrect line count received", t)

		for index, line := range result {
			FailIfFalse(line == expected[index], fmt.Sprintf("Expected line %d to be %s, got %s", index, expected[index], line), t)
		}
	})

	t.Run("Expand when inserting characters", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.Insert('e')
		buffer.Insert('f')
		buffer.Insert('g')
		buffer.Insert('h')
		buffer.Insert('i')
		buffer.Insert('j')
		buffer.Insert('k')
		buffer.Insert('l')
		buffer.Insert('m')
		buffer.Insert('n')
		buffer.Insert('o')
		buffer.Insert('p')
		buffer.Insert('q')
		buffer.Insert('r')
		buffer.Insert('s')
		buffer.Insert('t')

		FailIfFalse(len(buffer.Data) == 32, "Buffer is incorrectly expanded", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abcdefghijklmnopqrst", "Incorrect resulting text", t)
	})
}

func TestRemoveBefore(t *testing.T) {
	t.Run("Remove from the end of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.RemoveBefore()
		buffer.RemoveBefore()

		FailIfFalse(buffer.Cursor.Column == 2, "Cursor column is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "ab", "Incorrect resulting text", t)
	})

	t.Run("Remove symbol from empty buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.RemoveBefore()

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "", "Incorrect resulting text", t)
	})

	t.Run("Remove new line symbol", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.Insert('\n')
		buffer.Insert('e')
		buffer.RemoveBefore()
		buffer.RemoveBefore()

		FailIfFalse(buffer.Cursor.Column == 4, "Cursor column is not where it should be", t)
		FailIfFalse(buffer.Cursor.Line == 0, "Cursor line is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abcd", "Incorrect resulting text", t)
	})
}

func TestRemoveAfter(t *testing.T) {
	t.Run("Remove from the end of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.RemoveAfter()

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abc", "Incorrect resulting text", t)
	})

	t.Run("Remove from the middle of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.RemoveAfter()
		buffer.RemoveAfter()

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "ad", "Incorrect resulting text", t)
	})
}

func TestMoveLeft(t *testing.T) {
	t.Run("Move in the middle of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.Insert('e')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()

		FailIfFalse(buffer.Cursor.Column == 1, "Cursor column is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abcde", "Incorrect resulting text", t)
	})

	t.Run("Move at the beginning of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()

		FailIfFalse(buffer.Cursor.Column == 0, "Cursor column is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "ab", "Incorrect resulting text", t)
	})

	t.Run("Move at the beginning of line", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('\n')
		buffer.Insert('c')
		buffer.Insert('d')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()

		FailIfFalse(buffer.Cursor.Column == 0, "Cursor column is not where it should be", t)
		FailIfFalse(buffer.Cursor.Line == 1, "Cursor line is not where it should be", t)

		result := buffer.GetText()
		expected := []string{"ab", "cd"}
		FailNowIfFalse(len(result) == len(expected), "Incorrect line count received", t)

		for index, line := range result {
			FailIfFalse(line == expected[index], fmt.Sprintf("Expected line %d to be %s, got %s", index, expected[index], line), t)
		}
	})
}

func TestMoveRight(t *testing.T) {
	t.Run("Move in the middle of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveLeft()
		buffer.MoveRight()
		buffer.MoveRight()

		FailIfFalse(buffer.Cursor.Column == 2, "Cursor column is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abc", "Incorrect resulting text", t)
	})

	t.Run("Move at the end of buffer", func(t *testing.T) {
		buffer := CreateBuffer(16)
		buffer.Insert('a')
		buffer.Insert('b')
		buffer.Insert('c')
		buffer.MoveRight()
		buffer.MoveRight()
		buffer.MoveRight()
		buffer.MoveRight()

		FailIfFalse(buffer.Cursor.Column == 3, "Cursor column is not where it should be", t)

		result := buffer.GetText()
		FailNowIfFalse(len(result) == 1, "Incorrect line count received", t)
		FailIfFalse(result[0] == "abc", "Incorrect resulting text", t)
	})

	// @TODO (!important) write test for moving right at the end of the line
}
