package main

// @TODO (!important) 2 buffers support
// @TODO (!important) language server
// @TODO (!important) snippets
// @TODO (!important) task
// @TODO (!important) surround with
// @TODO (!important) search project
// @TODO (!important) search and replace file/project
// @TODO (!important) builtin todos
// @TODO (!important) language switching
// @TODO (!important) auto formatting

// @NEXT ctrl + backspace in insert mode
// @NEXT notifications

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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
	Icon       *sdl.Surface

	StatusBar      StatusBar
	Buffer         Buffer
	FileSearch     FileSearch
	CommandPalette CommandPalette
	Search         Search
	Commands       map[string]func(app *App)

	Mode               Mode
	Submode            Submode
	AmountModifier     strings.Builder
	Project            Project
	Cache              Cache
	FileSearchOpen     bool
	CommandPaletteOpen bool
	SearchOpen         bool
	CapsOn             bool
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
	result.Icon = LoadIcon("./assets/images/icon.png")

	result.StatusBar = CreateStatusBar(renderer, &result.WindowRect)
	result.Buffer = CreateBuffer(result.LineHeight, &result.RegularFont14, sdl.Rect{X: 0, Y: 0, W: windowWidth, H: windowHeight - result.StatusBar.Rect.H})
	result.FileSearch = CreateFileSearch(result.LineHeight, &result.RegularFont14, &result.RegularFont12)
	result.CommandPalette = CreateCommandPalette(result.LineHeight, &result.RegularFont14)
	result.Search = CreateSearch(result.LineHeight, &result.RegularFont14)
	result.Commands = map[string]func(app *App){}

	cacheDir, _ := os.UserCacheDir()
	result.Cache = ParseCache(fmt.Sprintf("%s/agurkas", cacheDir))

	result.startNormalMode()
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
	app.CapsOn = input.CapsLock

	if app.FileSearchOpen {
		app.FileSearch.Tick(input)
		return
	}

	if app.CommandPaletteOpen {
		app.CommandPalette.Tick(input)
		return
	}

	if app.SearchOpen {
		app.Search.Tick(input)
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
				app.CommandPaletteOpen = false
			}

			return
		}

		if input.TypedCharacter == 'p' && !app.FileSearchOpen {
			app.openFileSearch()
			app.CommandPaletteOpen = false
		} else if input.TypedCharacter == 's' && app.Buffer.Dirty {
			app.saveSourceFile()
		} else if input.TypedCharacter == 'o' {
			app.openSourceFile("")
		} else if input.TypedCharacter == 'O' {
			app.showFileInExplorer()
		}

		return
	}

	switch app.Mode {
	case Mode_Visual:
		fallthrough
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
	if app.CapsOn {
		app.StatusBar.RenderCaps(renderer, "CAPS ON", &app.RegularFont14, &app.Theme.StatusBar)
	}

	if app.FileSearchOpen {
		app.FileSearch.Render(renderer, &app.Buffer.Rect, &app.Theme.FileSearch)
	} else if app.CommandPaletteOpen {
		app.CommandPalette.Render(renderer, &app.Buffer.Rect, &app.Theme.FileSearch)
	} else if app.SearchOpen {
		app.Search.Render(renderer, &app.Buffer.Rect, &app.Theme.FileSearch)
	}

	renderer.Present()
}

// ==============================================================
// PRIVATE FUNCTIONS
// ==============================================================

func (app *App) handleInputNormal(input Input) {
	// @TODO (!important) e and E (move end word)
	// @TODO (!important) u (undo)
	// @TODO (!important) p and P (paste)
	// @TODO (!important) V (visual mode and visual line mode)
	// @TODO (!important) gd and ga and gv and gh (goto)
	// @TODO (!important) ck and cj (change)
	// @TODO (!important) ci and vi and di
	// @TODO (!important) daw
	// @TODO (!important) vaf and vaw

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

	if input.Escape {
		app.AmountModifier.Reset()
		app.startNormalMode()
		return
	}

	if input.TypedCharacter >= '1' && input.TypedCharacter <= '9' || input.TypedCharacter == '0' && app.AmountModifier.Len() > 0 {
		app.AmountModifier.WriteByte(input.TypedCharacter)
		return
	}

	switch input.TypedCharacter {
	case 'j':
		if app.AmountModifier.Len() > 0 {
			amount, _ := strconv.Atoi(app.AmountModifier.String())
			app.AmountModifier.Reset()

			app.Buffer.MoveDownByLines(amount)
		} else {
			app.Buffer.MoveDown()
		}
	case 'k':
		if app.AmountModifier.Len() > 0 {
			amount, _ := strconv.Atoi(app.AmountModifier.String())
			app.AmountModifier.Reset()

			app.Buffer.MoveUpByLines(amount)
		} else {
			app.Buffer.MoveUp()
		}
	case 'h':
		app.Buffer.MoveLeft()
	case 'H':
		app.Buffer.MoveUpByLines(32)
	case 'l':
		app.Buffer.MoveRight()
	case 'L':
		app.Buffer.MoveDownByLines(32)
	case 'i':
		if app.Mode == Mode_Normal {
			app.startInsertMode()
		}
	case 'I':
		if app.Mode == Mode_Normal {
			// @TODO (!important) move to the first non white space character in the line
			app.startInsertMode()
			app.Buffer.MoveToStartOfLine()
		}
	case 'a':
		if app.Mode == Mode_Normal {
			app.startInsertMode()
			app.Buffer.MoveRight()
		}
	case 'A':
		if app.Mode == Mode_Normal {
			app.startInsertMode()
			app.Buffer.MoveToEndOfLine()
		}
	case 'o':
		if app.Mode == Mode_Normal {
			app.startInsertMode()
			app.Buffer.InsertNewLineBelow()
		}
	case 'O':
		if app.Mode == Mode_Normal {
			app.startInsertMode()
			app.Buffer.InsertNewLineAbove()
		}
	case 'x':
		if app.Mode == Mode_Normal {
			app.Buffer.RemoveAfter()
		} else if app.Mode == Mode_Visual {
			app.Buffer.RemoveSelection()
			app.startNormalMode()
		}
	case 'D':
		if app.Mode == Mode_Normal {
			app.Buffer.RemoveRemainingLine()
		}
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
		if app.AmountModifier.Len() > 0 {
			amount, _ := strconv.Atoi(app.AmountModifier.String())
			app.AmountModifier.Reset()
			app.Buffer.MoveToLine(int32(amount))
		} else {
			app.Buffer.MoveToBufferEnd()
		}
	case 'J':
		if app.Mode == Mode_Normal {
			app.Buffer.MergeLineBelow()
		}
	case 'r':
		app.Submode = Submode_Replace
	case 'd':
		if app.Mode == Mode_Normal {
			app.Submode = Submode_Delete
		} else if app.Mode == Mode_Visual {
			app.Buffer.RemoveSelection()
		}
	case 'c':
		if app.Mode == Mode_Normal {
			app.Submode = Submode_Change
		} else if app.Mode == Mode_Visual {
			app.Buffer.RemoveSelection()
			app.startInsertMode()
		}
	case 'C':
		if app.Mode == Mode_Normal {

			app.startInsertMode()
			app.Buffer.ChangeRemainingLine()
		}
	case 'm':
		if app.Mode == Mode_Normal {
			app.Buffer.MarkCurrentPosition()
		}
	case '`':
		if app.Mode == Mode_Normal {
			app.Buffer.MoveToBookmark()
		}
	case 'f':
		if app.Mode == Mode_Normal {
			app.Submode = Submode_FindNext
		}
	case 'F':
		if app.Mode == Mode_Normal {
			// @TODO (!important) perhaps this is not very useful
			app.Submode = Submode_FindPrev
		}
	case ';':
		if app.Mode == Mode_Normal {
			app.Buffer.MoveToNextLineQuerySymbol()
		}
	case ',':
		if app.Mode == Mode_Normal {
			app.Buffer.MoveToPrevLineQuerySymbol()
		}
	case '>':
		if app.Mode == Mode_Normal {
			app.Buffer.Indent()
		} else if app.Mode == Mode_Visual {
			app.Buffer.IndentSelection()
			app.startNormalMode()
		}
	case '<':
		if app.Mode == Mode_Normal {
			app.Buffer.Outdent()
		} else if app.Mode == Mode_Visual {
			app.Buffer.OutdentSelection()
			app.startNormalMode()
		}
	case 'v':
		if app.Mode == Mode_Normal {
			app.startVisualMode()
		}
	case 'y':
		text := app.Buffer.GetSelectionText()
		err := sdl.SetClipboardText(text)
		checkError(err)

		app.startNormalMode()
	case 'Y':
		text := app.Buffer.GetCurrentLineText()
		err := sdl.SetClipboardText(text)
		checkError(err)

		app.startNormalMode()
	case ':':
		app.openCommandPalette()
	case '/':
		app.openSearch()
	case 'n':
		app.Buffer.MoveToNextFindResult()
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
		app.startNormalMode()
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

	if input.TypedCharacter >= '1' && input.TypedCharacter <= '9' || input.TypedCharacter == '0' && app.AmountModifier.Len() > 0 {
		app.AmountModifier.WriteByte(input.TypedCharacter)
		return
	}

	switch input.TypedCharacter {
	case 'd':
		app.Buffer.RemoveCurrentLine()
		app.Submode = Submode_None
	case 'j':
		amount := 1
		if app.AmountModifier.Len() > 0 {
			number, _ := strconv.Atoi(app.AmountModifier.String())
			amount = number
			app.AmountModifier.Reset()
		}

		app.Buffer.RemoveLines(Direction_Down, amount)
		app.Submode = Submode_None
	case 'k':
		amount := 1
		if app.AmountModifier.Len() > 0 {
			number, _ := strconv.Atoi(app.AmountModifier.String())
			amount = number
			app.AmountModifier.Reset()
		}

		app.Submode = Submode_None
		app.Buffer.RemoveLines(Direction_Up, amount)
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

func (app *App) openCommandPalette() {
	app.CommandPaletteOpen = true
	app.CommandPalette.Open(func(command string) {
		app.CommandPaletteOpen = false
		if command == "" {
			return
		}

		com := app.Commands[command]
		if com != nil {
			com(app)
		}
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

func (app *App) openSearch() {
	app.SearchOpen = true
	app.Search.Open(func(value string) {
		app.SearchOpen = false
		if value == "" {
			return
		}

		app.Buffer.Find(value)
	})
}

func (app *App) saveSourceFile() {
	text, _ := app.Buffer.GetText()
	filepath, success := SaveFile(app.Buffer.Filepath, text)
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

func (app *App) showFileInExplorer() {
	RunCommand("explorer", "", "/select,", app.Buffer.Filepath)
}

func (app *App) startInsertMode() {
	app.Mode = Mode_Insert
	if app.Theme.Buffer.CursorColorMatchModeColor {
		app.Buffer.Cursor.Color = app.Theme.StatusBar.InsertColor
	} else {
		app.Buffer.Cursor.Color = app.Theme.Buffer.CursorColor
	}
}

func (app *App) startNormalMode() {
	app.Mode = Mode_Normal
	if app.Theme.Buffer.CursorColorMatchModeColor {
		app.Buffer.Cursor.Color = app.Theme.StatusBar.NormalColor
	} else {
		app.Buffer.Cursor.Color = app.Theme.Buffer.CursorColor
	}

	app.Buffer.StopSelection()
}

func (app *App) startVisualMode() {
	app.Mode = Mode_Visual
	if app.Theme.Buffer.CursorColorMatchModeColor {
		app.Buffer.Cursor.Color = app.Theme.StatusBar.VisualColor
	} else {
		app.Buffer.Cursor.Color = app.Theme.Buffer.CursorColor
	}

	app.Buffer.StartSelection()
}
