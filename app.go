package main

import (
	"fmt"
	"path/filepath"
	"strings"

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

type App struct {
	RegularFont14 Font
	BoldFont14    Font

	ColorMap   map[Mode]sdl.Color
	WindowRect sdl.Rect

	LineHeight int32

	StatusBar  StatusBar
	Buffer     Buffer
	FileSearch FileSearch

	Mode           Mode
	Submode        Submode
	Project        Project
	FileSearchOpen bool
}

// @TODO is there a way to avoid passing a renderer here?
func Init(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) (result App) {
	result.RegularFont14 = LoadFont("./assets/fonts/consola.ttf", 14)
	result.BoldFont14 = LoadFont("./assets/fonts/consolab.ttf", 14)

	result.ColorMap = make(map[Mode]sdl.Color)
	result.ColorMap[Mode_Normal] = sdl.Color{R: 90, G: 169, B: 230, A: 255}
	result.ColorMap[Mode_Insert] = sdl.Color{R: 213, G: 41, B: 65, A: 255}
	result.ColorMap[Mode_Visual] = sdl.Color{R: 245, G: 213, B: 71, A: 255}
	result.ColorMap[Mode_VisualLine] = sdl.Color{R: 73, G: 132, B: 103, A: 255}
	result.WindowRect = sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}

	result.LineHeight = 18

	result.StatusBar = CreateStatusBar(renderer, &result.WindowRect)
	result.Buffer = CreateBuffer(result.LineHeight, &result.RegularFont14, sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight - result.StatusBar.Rect.H})
	result.FileSearch = CreateFileSearch(result.LineHeight, &result.RegularFont14)

	result.Mode = Mode_Normal
	result.Submode = Submode_None

	return
}

func (app *App) Resized(windowWidth int32, windowHeight int32) {
	app.WindowRect.W = windowWidth
	app.WindowRect.H = windowHeight
	app.Buffer.Rect.W = windowWidth
	app.Buffer.Rect.H = windowHeight - app.StatusBar.Rect.H
	app.StatusBar.Update(&app.WindowRect)
}

func (app *App) createProject(dirPath string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "root: %s", dirPath)

	dirName := filepath.Base(dirPath)

	finalPath, saveSuccess := SaveFile(filepath.Join(dirPath, fmt.Sprintf("%s.aproject", dirName)), []string{sb.String()})
	if !saveSuccess {
		return
	}
	data, _, openSuccess := OpenFile(finalPath)
	if !openSuccess {
		return
	}

	app.Mode = Mode_Normal
	app.Buffer.SetData(data, finalPath)
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

	app.StatusBar.Begin(renderer)
	app.StatusBar.RenderMode(renderer, app.Mode, app.ColorMap[app.Mode], &app.BoldFont14)
	app.StatusBar.RenderSubmode(renderer, app.Submode, &app.RegularFont14)
	app.StatusBar.RenderProject(renderer, app.Project.Name, GetFileNameFromPath(app.Buffer.Filepath), app.Buffer.Dirty, &app.RegularFont14)
	app.StatusBar.RenderLineCount(renderer, fmt.Sprintf("Lines: %d", app.Buffer.TotalLines), &app.RegularFont14)

	if app.FileSearchOpen {
		app.FileSearch.Render(renderer, &app.Buffer.Rect)
	}
}

func (app *App) Tick(input Input) {
	// @TODO (!important) cleanup this function
	if app.FileSearchOpen {
		app.FileSearch.Tick(input)
		return
	}

	if input.Ctrl {
		if input.Alt {
			// Save workspace
			if input.TypedCharacter == 's' {
				path, success := SelectDirectory()
				if success {
					app.createProject(path)
				}

				return
			}

			// Open workspace
			if input.TypedCharacter == 'o' {
				data, _, success := OpenFile("")
				if success {
					app.Project = ParseProject(string(data))
				}

				return
			}

			if input.TypedCharacter == 'p' {
				return
			}
		}

		if input.TypedCharacter == 'p' && !app.FileSearchOpen {
			app.FileSearchOpen = true
			app.FileSearch.Open(PathsToFileSearchEntries(app.Project.Files), func(path string) {
				app.FileSearchOpen = false
				if path != "" {
					return
				}

				data, _, success := OpenFile(path)
				if success {
					app.Mode = Mode_Normal
					app.Buffer.SetData(data, path)
				}
			})
		}

		// Save a file
		if input.TypedCharacter == 's' && app.Buffer.Dirty {
			filepath, success := SaveFile(app.Buffer.Filepath, app.Buffer.GetText())
			if success {
				app.Buffer.Filepath = filepath
				app.Buffer.Dirty = false
			}

			return
		}

		// Open a file
		if input.TypedCharacter == 'o' {
			data, filepath, success := OpenFile("")
			if success {
				app.Mode = Mode_Normal
				app.Buffer.SetData(data, filepath)
			}

			return
		}

	}

	// @TODO (!important) ctrl + shift + o command
	// @TODO (!important) ctrl + alt + p
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
