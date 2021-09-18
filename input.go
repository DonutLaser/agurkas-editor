package main

type Input struct {
	TypedCharacter byte
	Backspace      bool
	Escape         bool
	Ctrl           bool
	Alt            bool
	Shift          bool
	CapsLock       bool // Used only to know when to visually show that caps lock is on, you already get uppercase typed character if caps lock is on
}

func (input *Input) Clear() {
	input.TypedCharacter = 0
	input.Backspace = false
	input.Escape = false
}
