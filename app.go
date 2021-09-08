package main

// @TODO (!important) 2 buffers support
// @TODO (!important) language server
// @TODO (!important) project selection
// @TODO (!important) snippets
// @TODO (!important) task
// @TODO (!important) syntax highlighting
// @TODO (!important) intellisense
// @TODO (!important) surround with
// @TODO (!important) search file/project
// @TODO (!important) search and replace file/project
// @TODO (!important) app.go is lagging already

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

// ==============================================================
// TYPES
// ==============================================================

type Mode string
type Submode string

const (
	Mode_Normal     Mode = "NORMAL"
	Mode_Insert     Mode = "INSERT"
	Mode_Visual     Mode = "VISUAL"
	Mode_VisualLine Mode = "VISUAL LINE"
)

const (
	Submode_Replace  Submode = "replace"
	Submode_Delete   Submode = "delete"
	Submode_Goto     Submode = "goto"
	Submode_Change   Submode = "change"
	Submode_FindNext Submode = "find next"
	Submode_FindPrev Submode = "find prev"
	Submode_None     Submode = "none"
)

type App struct {
	RegularFont14 Font
	RegularFont12 Font
	BoldFont14    Font

	WindowRect sdl.Rect
	Theme      Theme
	LineHeight int32

	StatusBar  StatusBar
	Buffer     Buffer
	FileSearch FileSearch

	Mode           Mode
	Submode        Submode
	Project        Project
	Cache          Cache
	FileSearchOpen bool
}

// ==============================================================
// PUBLIC FUNCTIONS
// ==============================================================

// @TODO is there a way to avoid passing a renderer here?
func Init(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) (result App) {
	result.RegularFont14 = LoadFont("./assets/fonts/consola.ttf", 14)
	result.RegularFont12 = LoadFont("./assets/fonts/consola.ttf", 12)
	result.BoldFont14 = LoadFont("./assets/fonts/consolab.ttf", 14)

	result.WindowRect = sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight}
	result.Theme = ParseTheme("./default_theme.atheme")
	result.LineHeight = 18

	result.StatusBar = CreateStatusBar(renderer, &result.WindowRect)
	result.Buffer = CreateBuffer(result.LineHeight, &result.RegularFont14, sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight - result.StatusBar.Rect.H})
	result.FileSearch = CreateFileSearch(result.LineHeight, &result.RegularFont14, &result.RegularFont12)
	cacheDir, _ := os.UserCacheDir()
	result.Cache = ParseCache(fmt.Sprintf("%s/agurkas", cacheDir))

	result.Mode = Mode_Normal
	result.Submode = Submode_None

	return
}

func (app *App) Close() {
	app.RegularFont14.Unload()
	app.BoldFont14.Unload()
}

func (app *App) Resized(windowWidth int32, windowHeight int32) {
	app.WindowRect.W = windowWidth
	app.WindowRect.H = windowHeight
	app.Buffer.Rect.W = windowWidth
	app.Buffer.Rect.H = windowHeight - app.StatusBar.Rect.H
	app.StatusBar.Update(&app.WindowRect)
}

func (app *App) Tick(input Input) {
	if app.FileSearchOpen {
		app.FileSearch.Tick(input)
		return
	}

	if input.Ctrl {
		if input.Alt {
			// Save workspace
			if input.TypedCharacter == 's' {
				app.saveProject()
			} else if input.TypedCharacter == 'o' {
				app.openProject("")
			} else if input.TypedCharacter == 'p' {
				app.openProjectSearch()
			}

			return
		}

		if input.TypedCharacter == 'p' && !app.FileSearchOpen {
			app.openFileSearch()
		} else if input.TypedCharacter == 's' && app.Buffer.Dirty {
			app.saveSourceFile()
		} else if input.TypedCharacter == 'o' {
			app.openSourceFile("")
		}

		return
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

func (app *App) Render(renderer *sdl.Renderer) {
	cc := app.Theme.Buffer.BackgroundColor
	renderer.SetDrawColor(cc.R, cc.G, cc.B, cc.A)
	renderer.Clear()

	app.Buffer.Render(renderer, app.Mode, &app.Theme)

	app.StatusBar.Begin(renderer, &app.Theme.StatusBar)
	app.StatusBar.RenderMode(renderer, app.Mode, &app.BoldFont14, &app.Theme.StatusBar)
	app.StatusBar.RenderSubmode(renderer, app.Submode, &app.RegularFont14, &app.Theme.StatusBar)
	app.StatusBar.RenderProject(renderer, app.Project.Name, GetFileNameFromPath(app.Buffer.Filepath), app.Buffer.Dirty, &app.RegularFont14, &app.Theme.StatusBar)
	app.StatusBar.RenderLineCount(renderer, fmt.Sprintf("Lines: %d", app.Buffer.TotalLines), &app.RegularFont14, &app.Theme.StatusBar)

	if app.FileSearchOpen {
		app.FileSearch.Render(renderer, &app.Buffer.Rect, &app.Theme.FileSearch)
	}

	renderer.Present()
}

// ==============================================================
// PRIVATE FUNCTIONS
// ==============================================================

func (app *App) handleInputNormal(input Input) {
	// @TODO (!important) e and E (move end word)
	// @TODO (!important) y (yank)
	// @TODO (!important) u (undo)
	// @TODO (!important) p and P (paste)
	// @TODO (!important) H and L (move to viewport top and down)
	// @TODO (!important) v and V (visual mode and visual line mode)
	// @TODO (!important) gd and ga and gv and gh (goto)
	// @TODO (!important) ck and cj (change)
	// @TODO (!important) ci and vi and di

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

	if app.Submode == Submode_FindNext || app.Submode == Submode_FindPrev {
		app.handleInputSubmodeFind(input)
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
	case 'f':
		app.Submode = Submode_FindNext
	case 'F':
		// @TODO (!important) perhaps this is not very useful
		app.Submode = Submode_FindPrev
	case ';':
		app.Buffer.MoveToNextLineQuerySymbol()
	case ',':
		app.Buffer.MoveToPrevLineQuerySymbol()
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

func (app *App) handleInputSubmodeFind(input Input) {
	if input.Ctrl || input.Alt {
		return
	}

	if input.Escape {
		app.Submode = Submode_None
		return
	}

	if input.TypedCharacter != 0 {
		forwards := app.Submode == Submode_FindNext
		app.Buffer.FindInLine(input.TypedCharacter, forwards)
		app.Submode = Submode_None
	}
}

func (app *App) createProject(dirPath string) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "root: %s", dirPath)

	dirName := filepath.Base(dirPath)

	finalPath, saveSuccess := SaveFile(filepath.Join(dirPath, fmt.Sprintf("%s.aproject", dirName)), []string{sb.String()})
	if !saveSuccess {
		return
	}

	app.Cache.Write("project", finalPath)
	app.openSourceFile(finalPath)
}

func (app *App) saveProject() {
	path, success := SelectDirectory()
	if success {
		app.createProject(path)
	}
}

func (app *App) openProject(path string) {
	data, _, success := OpenFile(path)
	if success {
		app.Project = ParseProject(string(data))
	}
}

func (app *App) openFileSearch() {
	app.FileSearchOpen = true
	app.FileSearch.Open(PathsToFileSearchEntries(app.Project.Files), func(path string) {
		app.FileSearchOpen = false
		if path == "" {
			return
		}

		app.openSourceFile(path)
	})
}

func (app *App) openProjectSearch() {
	app.FileSearchOpen = true
	app.FileSearch.Open(PathsToFileSearchEntries(app.Cache.Projects), func(path string) {
		app.FileSearchOpen = false
		if path == "" {
			return
		}

		app.openProject(path)
	})
}

func (app *App) saveSourceFile() {
	filepath, success := SaveFile(app.Buffer.Filepath, app.Buffer.GetText())
	if success {
		app.Buffer.Filepath = filepath
		app.Buffer.Dirty = false
	}
}

func (app *App) openSourceFile(path string) {
	data, filepath, success := OpenFile(path)
	if success {
		app.Mode = Mode_Normal
		app.Buffer.SetData(data, filepath)
	}
}
