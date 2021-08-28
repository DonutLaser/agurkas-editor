package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
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
	Submode_Replace Submode = "replace"
	Submode_Delete  Submode = "delete"
	Submode_Goto    Submode = "goto"
	Submode_Change  Submode = "change"
	Submode_None    Submode = "none"
)

type PlatformApi struct {
	SetWindowTitle func(string)
}

type App struct {
	PlatformApi PlatformApi

	RegularFont14 Font
	BoldFont14    Font

	ColorMap          map[Mode]sdl.Color
	StatusBarTriangle Image
	WindowRect        sdl.Rect

	LineHeight int32

	Buffer Buffer

	Mode    Mode
	Submode Submode
}

// @TODO is there a way to avoid passing a renderer here?
func Init(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) (result App) {
	result = App{}

	result.RegularFont14 = LoadFont("./assets/fonts/consola.ttf", 14)
	result.BoldFont14 = LoadFont("./assets/fonts/consolab.ttf", 14)

	result.StatusBarTriangle = LoadImage("./assets/images/status_bar_triangle.png", renderer)
	result.ColorMap = make(map[Mode]sdl.Color)
	result.ColorMap[Mode_Normal] = sdl.Color{R: 90, G: 169, B: 230, A: 255}
	result.ColorMap[Mode_Insert] = sdl.Color{R: 213, G: 41, B: 65, A: 255}
	result.ColorMap[Mode_Visual] = sdl.Color{R: 245, G: 213, B: 71, A: 255}
	result.ColorMap[Mode_VisualLine] = sdl.Color{R: 73, G: 132, B: 103, A: 255}

	result.WindowRect = sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}

	result.LineHeight = 18

	result.Mode = Mode_Normal
	result.Submode = Submode_None
	result.Buffer = CreateBuffer(result.LineHeight, &result.RegularFont14, sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight - 22})

	return
}

func (app *App) Resized(windowWidth int32, windowHeight int32) {
	app.WindowRect.W = windowWidth
	app.WindowRect.H = windowHeight
	app.Buffer.Rect.W = windowWidth
	app.Buffer.Rect.H = windowHeight - 22
}

func (app *App) handleInputNormal(input Input) {
	// @TODO (!important) e and E (move end word)
	// @TODO (!important) y (yank)
	// @TODO (!important) u (undo)
	// @TODO (!important) p and P (paste)
	// @TODO (!important) H and L (move to viewport top and down)
	// @TODO (!important) f and F (find forward and backward)
	// @TODO (!important) v and V (visual mode and visual line mode)
	// @TODO (!important) gd and ga and gv and gh (goto)
	// @TODO (!important) ck and cj (change)

	if app.Submode == Submode_Replace {
		app.handleInputSubmodeReplace(input)
		return
	}

	if app.Submode == Submode_Delete {
		app.handleInputSubmodeDelete(input)
		return
	}

	if app.Submode == Submode_Goto {
		app.handleInputSubmodeGoto(input)
		return
	}

	if app.Submode == Submode_Change {
		app.handleInputSubmodeChange(input)
		return
	}

	if input.Ctrl {
		if input.TypedCharacter == 's' && app.Buffer.Dirty {
			filepath, success := SaveFile(app.Buffer.Filepath, app.Buffer.GetText())
			if success {
				app.Buffer.Filepath = filepath
				app.Buffer.Dirty = false
				app.PlatformApi.SetWindowTitle(filepath)
			}

			return
		}

		if input.TypedCharacter == 'o' {
			data, filepath, success := OpenFile("")
			if success {
				app.Buffer.SetData(data, filepath)
				app.PlatformApi.SetWindowTitle(app.Buffer.Filepath)
			}

			return
		}
	}

	switch input.TypedCharacter {
	case 'j':
		app.Buffer.MoveDown()
	case 'k':
		app.Buffer.MoveUp()
	case 'h':
		app.Buffer.MoveLeft()
	case 'l':
		app.Buffer.MoveRight()
	case 'i':
		app.Mode = Mode_Insert
	case 'I':
		// @TODO (!important) move to the first non white space character in the line
		app.Mode = Mode_Insert
		app.Buffer.MoveToStartOfLine()
	case 'a':
		app.Mode = Mode_Insert
		app.Buffer.MoveRight()
	case 'A':
		app.Mode = Mode_Insert
		app.Buffer.MoveToEndOfLine()
	case 'o':
		app.Mode = Mode_Insert
		app.Buffer.InsertNewLineBelow()
	case 'O':
		app.Mode = Mode_Insert
		app.Buffer.InsertNewLineAbove()
	case 'x':
		app.Buffer.RemoveAfter()
	case 'D':
		app.Buffer.RemoveRemainingLine()
	case '0':
		app.Buffer.MoveToStartOfLine()
	case '$':
		app.Buffer.MoveToEndOfLine()
	case 'w':
		app.Buffer.MoveRightToWordStart(false) // Do not ignore punctuation
	case 'W':
		app.Buffer.MoveRightToWordStart(true) // Ignore puctuation
	case 'b':
		app.Buffer.MoveLeftToWordStart(false) // Do not ignore punctuation
	case 'B':
		app.Buffer.MoveLeftToWordStart(true) // Ignore punctuation
	case 'g':
		app.Submode = Submode_Goto
	case 'G':
		app.Buffer.MoveToBufferEnd()
	case 'J':
		app.Buffer.MergeLineBelow()
	case 'r':
		app.Submode = Submode_Replace
	case 'd':
		app.Submode = Submode_Delete
	case 'c':
		app.Submode = Submode_Change
	case 'C':
		app.Buffer.ChangeRemainingLine()
		app.Mode = Mode_Insert
	case 'm':
		app.Buffer.MarkCurrentPosition()
	case '`':
		app.Buffer.MoveToBookmark()
	}
}

func (app *App) handleInputInsert(input Input) {
	if input.Ctrl || input.Alt {
		return
	}

	if input.TypedCharacter != 0 {
		app.Buffer.Insert(input.TypedCharacter)
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

func (app *App) handleInputSubmodeChange(input Input) {
	if input.Ctrl || input.Alt {
		return
	}

	if input.Escape {
		app.Submode = Submode_None
	}

	if input.TypedCharacter == 'c' {
		app.Buffer.ChangeCurrentLine()
		app.Submode = Submode_None
		app.Mode = Mode_Insert
		return
	}
}

func (app *App) handleInputSubmodeGoto(input Input) {
	if input.Ctrl || input.Alt {
		return
	}

	if input.Escape {
		app.Submode = Submode_None
		return
	}

	if input.TypedCharacter == 'g' {
		app.Buffer.MoveToBufferStart()
		app.Submode = Submode_None
		return
	}
}

func (app *App) handleInputSubmodeReplace(input Input) {
	if input.Ctrl || input.Alt {
		return
	}

	if input.Escape {
		app.Submode = Submode_None
		return
	}

	if input.TypedCharacter != 0 {
		app.Buffer.ReplaceCurrentCharacter(input.TypedCharacter)
		app.Submode = Submode_None
		return
	}
}

func (app *App) handleInputSubmodeDelete(input Input) {
	if input.Ctrl || input.Alt {
		return
	}

	if input.Escape {
		app.Submode = Submode_None
		return
	}

	switch input.TypedCharacter {
	case 'd':
		app.Buffer.RemoveCurrentLine()
		app.Submode = Submode_None
	case 'j':
		app.Submode = Submode_None
		app.Buffer.RemoveLines(Direction_Down, 1)
	case 'k':
		app.Submode = Submode_None
		app.Buffer.RemoveLines(Direction_Up, 1)
	}
}

func (app *App) Render(renderer *sdl.Renderer) {
	app.Buffer.Render(renderer, app.Mode, app.ColorMap[app.Mode])

	statusBarRect := sdl.Rect{
		X: 0,
		Y: app.WindowRect.H - 22,
		W: app.WindowRect.W,
		H: 22,
	}
	DrawRect(renderer, &statusBarRect, sdl.Color{R: 48, G: 48, B: 48, A: 255})

	statusWidth, statusHeight := app.BoldFont14.GetStringSize(string(app.Mode))
	statusBgRect := sdl.Rect{
		X: 0,
		Y: statusBarRect.Y,
		W: statusWidth + 16,
		H: statusBarRect.H,
	}
	DrawRect(renderer, &statusBgRect, app.ColorMap[app.Mode])
	statusRect := sdl.Rect{
		X: 8,
		Y: statusBarRect.Y + (statusBarRect.H-statusHeight)/2,
		W: statusWidth,
		H: statusHeight,
	}
	DrawText(renderer, &app.BoldFont14, string(app.Mode), &statusRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	app.StatusBarTriangle.Render(renderer, sdl.Point{X: statusBgRect.X + statusBgRect.W, Y: statusBgRect.Y}, app.ColorMap[app.Mode])

	if app.Submode != Submode_None {
		submodeWidth, submodeHeight := app.RegularFont14.GetStringSize(string(app.Submode))
		submodeRect := sdl.Rect{
			X: statusBgRect.X + statusBgRect.W + app.StatusBarTriangle.Width + 5,
			Y: statusBarRect.Y + (statusBarRect.H-submodeHeight)/2,
			W: submodeWidth,
			H: submodeHeight,
		}
		DrawText(renderer, &app.RegularFont14, string(app.Submode), &submodeRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}

	linesStatus := fmt.Sprintf("Lines: %d", app.Buffer.TotalLines)
	linesWidth, linesHeight := app.RegularFont14.GetStringSize(linesStatus)
	linesRect := sdl.Rect{
		X: app.WindowRect.W - 8 - linesWidth,
		Y: statusBarRect.Y + (statusBarRect.H-linesHeight)/2,
		W: linesWidth,
		H: linesHeight,
	}
	DrawText(renderer, &app.RegularFont14, linesStatus, &linesRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
}

func (app *App) Tick(input Input) {
	// @TODO (!important) ctrl + shift + o command
	switch app.Mode {
	case Mode_Normal:
		app.handleInputNormal(input)
	case Mode_Insert:
		app.handleInputInsert(input)
	}
}

func (app *App) Close() {
	app.RegularFont14.Unload()
	app.BoldFont14.Unload()
}
