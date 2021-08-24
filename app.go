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

type PlatformApi struct {
	SetWindowTitle func(string)
}

type App struct {
	PlatformApi PlatformApi

	RegularFont14 *ttf.Font
	BoldFont14    *ttf.Font

	ColorMap map[Mode]sdl.Color

	LineHeight int32

	Buffer Buffer

	Mode    Mode
	Submode Submode
}

func Init() (result App) {
	result = App{}

	result.RegularFont14 = LoadFont("consola.ttf", 14)
	result.BoldFont14 = LoadFont("consolab.ttf", 14)

	result.ColorMap = make(map[Mode]sdl.Color)
	result.ColorMap[Mode_Normal] = sdl.Color{R: 90, G: 169, B: 230, A: 255}
	result.ColorMap[Mode_Insert] = sdl.Color{R: 213, G: 41, B: 65, A: 255}
	result.ColorMap[Mode_Visual] = sdl.Color{R: 245, G: 213, B: 71, A: 255}
	result.ColorMap[Mode_VisualLine] = sdl.Color{R: 73, G: 132, B: 103, A: 255}

	result.LineHeight = 18

	result.Mode = Mode_Normal
	result.Submode = Submode_None
	result.Buffer = CreateBuffer(result.LineHeight, GetCharacterWidth(result.RegularFont14), 14)

	return
}

func (app *App) handleInputNormal(input Input) {
	// @TODO (!important) w and W (move word)
	// @TODO (!important) e and E (move end word)
	// @TODO (!important) y (yank)
	// @TODO (!important) u (undo)
	// @TODO (!important) p and P (paste)
	// @TODO (!important) H and L (move to viewport top and down)
	// @TODO (!important) f and F (find forward and backward)
	// @TODO (!important) v and V (visual mode and visual line mode)
	// @TODO (!important) gg and gd and ga and gv and gh (goto)
	// @TODO (!important) b and B (move word back)
	// @TODO (!important) cc and C and ck and cj (change)

	if app.Submode == Submode_Replace {
		app.handleInputSubmodeReplace(input)
		return
	}

	if app.Submode == Submode_Delete {
		app.handleInputSubmodeDelete(input)
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
	case 'G':
		app.Buffer.MoveToBufferEnd()
	case 'J':
		app.Buffer.MergeLineBelow()
	case 'r':
		app.Submode = Submode_Replace
	case 'd':
		app.Submode = Submode_Delete
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

func (app *App) Render(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) {
	app.Buffer.Render(renderer, app.RegularFont14, app.Mode, windowHeight, windowWidth)

	statusBarRect := sdl.Rect{
		X: 0,
		Y: windowHeight - 22,
		W: windowWidth,
		H: 22,
	}
	DrawRect(renderer, &statusBarRect, sdl.Color{R: 48, G: 48, B: 48, A: 255})

	statusWidth, statusHeight := GetStringSize(app.BoldFont14, string(app.Mode))
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
	DrawText(renderer, app.BoldFont14, string(app.Mode), &statusRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})

	if app.Submode != Submode_None {
		submodeWidth, submodeHeight := GetStringSize(app.RegularFont14, string(app.Submode))
		submodeRect := sdl.Rect{
			X: 10 + statusRect.W + 10,
			Y: statusBarRect.Y + (statusBarRect.H-submodeHeight)/2,
			W: submodeWidth,
			H: submodeHeight,
		}
		DrawText(renderer, app.RegularFont14, string(app.Submode), &submodeRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}
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
	app.RegularFont14.Close()
	app.BoldFont14.Close()
}
