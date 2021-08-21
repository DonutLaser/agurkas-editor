package main

const (
	Modifier_Ctrl  uint8 = 1
	Modifier_Shift uint8 = 2
	Modifier_Alt   uint8 = 4
	Modifier_None  uint8 = 0
)

type Input struct {
	Modifier       uint8
	TypedCharacter string
	Backspace      bool
	Escape         bool
}

func (input *Input) Clear() {
	input.Modifier = Modifier_None
	input.TypedCharacter = ""
	input.Backspace = false
	input.Escape = false
}
