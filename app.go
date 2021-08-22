package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Mode string
type Submode string

const (
	Mode_Normal     Mode = "NORMAL"
	Mode_Insert     Mode = "INSERT"
	Mode_Visual     Mode = "VISUAL"
	Mode_VisualLine Mode = "VISUAL LINE"
)

const (
	Submode_Replace Submode = "Replace"
	Submode_Delete  Submode = "Delete"
	Submode_Goto    Submode = "Goto"
	Submode_None    Submode = "None"
)

type App struct {
	Font       *ttf.Font
	LineHeight int32

	Buffer Buffer

	Mode    Mode
	Submode Submode
}

func Init() (result App) {
	result = App{}

	font, err := ttf.OpenFont("consola.ttf", 16)
	checkError(err)
	result.Font = font
	result.LineHeight = 16

	result.Mode = Mode_Normal
	result.Submode = Submode_None
	result.Buffer = CreateBuffer(result.LineHeight, GetCharacterWidth(result.Font))

	return
}

func (app *App) handleInputNormal(input Input) {
	// @TODO (!important) w and W (move word)
	// @TODO (!important) e and E (move end word)
	// @TODO (!important) y (yank)
	// @TODO (!important) u (undo)
	// @TODO (!important) p and P (paste)
	// @TODO (!important) dj and dk (delete)
	// @TODO (!important) H and L (move to viewport top and down)
	// @TODO (!important) f and F (find forward and backward)
	// @TODO (!important) v and V (visual mode and visual line mode)
	// @TODO (!important) m (mark)
	// @TODO (!important) G (goto file end)
	// @TODO (!important) gg (goto file start)
	// @TODO (!important) b and B (move word back)

	if app.Submode == Submode_Replace {
		app.handleInputSubmodeReplace(input)
		return
	}

	if app.Submode == Submode_Delete {
		app.handleInputSubmodeDelete(input)
		return
	}

	switch input.TypedCharacter {
	case "j":
		app.Buffer.MoveDown()
	case "k":
		app.Buffer.MoveUp()
	case "h":
		app.Buffer.MoveLeft()
	case "l":
		app.Buffer.MoveRight()
	case "i":
		app.Mode = Mode_Insert
	case "I":
		app.Mode = Mode_Insert
		app.Buffer.MoveToStartOfLine()
	case "a":
		app.Mode = Mode_Insert
		app.Buffer.MoveRight()
	case "A":
		app.Mode = Mode_Insert
		app.Buffer.MoveToEndOfLine()
	case "o":
		app.Mode = Mode_Insert
		app.Buffer.InsertNewLineBelow()
	case "O":
		app.Mode = Mode_Insert
		app.Buffer.InsertNewLineAbove()
	case "x":
		app.Buffer.RemoveAfter()
	case "D":
		app.Buffer.RemoveRemainingLine()
	case "0":
		app.Buffer.MoveToStartOfLine()
	case "$":
		app.Buffer.MoveToEndOfLine()
	case "J":
		app.Buffer.MergeLineBelow()
	case "r":
		app.Submode = Submode_Replace
	case "d":
		app.Submode = Submode_Delete
	}
}

func (app *App) handleInputInsert(input Input) {
	ctrl := input.Modifier & Modifier_Ctrl
	alt := input.Modifier & Modifier_Alt

	if ctrl != 0 || alt != 0 {
		return
	}

	if input.TypedCharacter != "" {
		app.Buffer.Insert(input.TypedCharacter[0])
		return
	}

	if input.Escape {
		app.Mode = Mode_Normal
		return
	}

	if input.Backspace {
		app.Buffer.RemoveBefore()
		return
	}
}

func (app *App) handleInputSubmodeReplace(input Input) {
	ctrl := input.Modifier & Modifier_Ctrl
	alt := input.Modifier & Modifier_Alt

	if ctrl != 0 || alt != 0 {
		return
	}

	if input.TypedCharacter != "" {
		app.Buffer.ReplaceCurrentCharacter(input.TypedCharacter[0])
		app.Submode = Submode_None
		return
	}
}

func (app *App) handleInputSubmodeDelete(input Input) {
	ctrl := input.Modifier & Modifier_Ctrl
	alt := input.Modifier & Modifier_Alt

	if ctrl != 0 || alt != 0 {
		return
	}

	switch input.TypedCharacter {
	case "d":
		app.Buffer.RemoveCurrentLine()
		app.Submode = Submode_None
	case "j":
		app.Submode = Submode_None
		app.Buffer.RemoveLines(Direction_Down, 1)
	case "k":
		app.Submode = Submode_None
		app.Buffer.RemoveLines(Direction_Up, 1)
	}
}

func (app *App) Render(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) {
	app.Buffer.Render(renderer, app.Font, app.Mode, windowHeight)

	statusBarRect := sdl.Rect{
		X: 0,
		Y: windowHeight - 32,
		W: windowWidth,
		H: 32,
	}
	DrawRect(renderer, &statusBarRect, sdl.Color{R: 20, G: 20, B: 20, A: 255})

	statusWidth, statusHeight := GetStringSize(app.Font, string(app.Mode))
	statusRect := sdl.Rect{
		X: 10,
		Y: statusBarRect.Y + (statusBarRect.H-statusHeight)/2,
		W: statusWidth,
		H: statusHeight,
	}
	DrawText(renderer, app.Font, string(app.Mode), &statusRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})

	if app.Submode != Submode_None {
		submodeWidth, submodeHeight := GetStringSize(app.Font, string(app.Submode))
		submodeRect := sdl.Rect{
			X: 10 + statusRect.W + 10,
			Y: statusBarRect.Y + (statusBarRect.H-submodeHeight)/2,
			W: submodeWidth,
			H: submodeHeight,
		}
		DrawText(renderer, app.Font, string(app.Submode), &submodeRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}
}

func (app *App) Tick(input Input) {
	switch app.Mode {
	case Mode_Normal:
		app.handleInputNormal(input)
	case Mode_Insert:
		app.handleInputInsert(input)
	}
}

func (app *App) Close() {
	app.Font.Close()
}
