package main

import (
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type FileSearchEntry struct {
	Name     string
	FullPath string
}

type FileSearch struct {
	SelectionIndex int32
	SearchQuery    strings.Builder

	Cursor      InputCursor
	Width       int32
	LineHeight  int32
	LineSpacing int32
	Font14      *Font
	Font12      *Font

	FileEntries  []FileSearchEntry
	FoundEntries []int // Array of indexes into file entries array

	CloseCallback func(string)

	altWasPressed bool
	firstTime     bool
}

func PathsToFileSearchEntries(paths []string) (result []FileSearchEntry) {
	for _, path := range paths {
		result = append(result, FileSearchEntry{
			Name:     GetFileNameFromPath(path),
			FullPath: path,
		})
	}

	return
}

func CreateFileSearch(lineHeight int32, font14 *Font, font12 *Font) (result FileSearch) {
	result.Cursor.Width = 2
	result.Cursor.Height = lineHeight
	result.Cursor.Advance = int32(font14.CharacterWidth)
	result.Width = 500
	result.LineHeight = lineHeight
	result.LineSpacing = (lineHeight - int32(font14.Size)) / 2
	result.Font14 = font14
	result.Font12 = font12

	return
}

func (fs *FileSearch) Open(availableFiles []FileSearchEntry, onClose func(string)) {
	fs.SelectionIndex = 0
	fs.Cursor.Column = 0
	fs.SearchQuery.Reset()

	fs.FileEntries = availableFiles

	size := Min(len(fs.FileEntries), 5)
	fs.FoundEntries = make([]int, size)
	for i := 0; i < size; i += 1 {
		fs.FoundEntries[i] = i
	}

	fs.CloseCallback = onClose

	fs.firstTime = true
}

func (fs *FileSearch) Close() {
	fs.CloseCallback("")
}

func (fs *FileSearch) Submit() {
	fs.CloseCallback(fs.FileEntries[fs.FoundEntries[fs.SelectionIndex]].FullPath)
}

func (fs *FileSearch) updateSearchResults() {
	fs.FoundEntries = make([]int, 0)
	if fs.SearchQuery.Len() == 0 {
		return
	}

	query := fs.SearchQuery.String()
	for index, entry := range fs.FileEntries {
		if strings.Contains(entry.Name, query) || strings.Contains(entry.FullPath, query) {
			fs.FoundEntries = append(fs.FoundEntries, index)
		}
	}

	if len(fs.FoundEntries) > 0 {
		fs.SelectionIndex = 0
	}
}

func (fs *FileSearch) Tick(input Input) {
	if input.Alt {
		if input.TypedCharacter == 'j' {
			fs.SelectionIndex = int32(Min(int(fs.SelectionIndex)+1, len(fs.FoundEntries)-1))
		} else if input.TypedCharacter == 'k' {
			fs.SelectionIndex = int32(Max(int(fs.SelectionIndex)-1, 0))
		}

		fs.altWasPressed = true

		return
	} else if fs.altWasPressed {
		if !fs.firstTime {
			fs.altWasPressed = false
			if len(fs.FoundEntries) > 0 {
				fs.Submit()
			}
		} else {
			fs.firstTime = false
			fs.altWasPressed = false
		}
	}

	if input.Escape {
		fs.Close()
		return
	}

	if input.Ctrl && input.Backspace {
		fs.Cursor.Column = 0
		fs.SearchQuery.Reset()
		fs.updateSearchResults()
		return
	}

	if input.TypedCharacter != 0 {
		if input.TypedCharacter == '\t' || input.TypedCharacter == '\n' {
			if len(fs.FoundEntries) > 0 {
				fs.Submit()
			}
		} else {
			fs.SearchQuery.WriteByte(input.TypedCharacter)
			fs.Cursor.Column += 1
			fs.updateSearchResults()
		}
		return
	}
}

func (fs *FileSearch) Render(renderer *sdl.Renderer, parentRect *sdl.Rect, theme *FileSearchTheme) {
	inputRect := sdl.Rect{
		X: parentRect.W/2 - fs.Width/2,
		Y: parentRect.Y + int32(float32(parentRect.H)*0.15),
		W: fs.Width,
		H: fs.LineHeight + 10,
	}

	borderRect := expandRect(sdl.Rect{
		X: inputRect.X,
		Y: inputRect.Y,
		W: fs.Width,
		H: inputRect.H + int32(len(fs.FoundEntries))*(fs.LineHeight+fs.LineSpacing*2),
	}, 1)

	DrawRect(renderer, &borderRect, theme.BorderColor)
	DrawRect(renderer, &inputRect, theme.InputBackgroundColor)
	fs.Cursor.Render(renderer, inputRect, theme.CursorColor)

	query := fs.SearchQuery.String()
	if len(query) > 0 {
		queryWidth := fs.Font14.GetStringWidth(query)
		queryRect := sdl.Rect{
			X: inputRect.X + 5,
			Y: inputRect.Y + (inputRect.H-int32(fs.Font14.Size))/2,
			W: queryWidth,
			H: int32(fs.Font14.Size),
		}
		DrawText(renderer, fs.Font14, query, &queryRect, theme.InputTextColor)
	}

	defaultBgColor := theme.ResultBackgroundColor
	defaultTextColor := theme.ResultNameColor

	activeBgColor := theme.ResultActiveColor
	activeTextColor := theme.ResultNameActiveColor

	for index, res := range fs.FoundEntries {
		entryColor := defaultBgColor
		textColor := defaultTextColor
		if index == int(fs.SelectionIndex) {
			entryColor = activeBgColor
			textColor = activeTextColor
		}

		entryRect := sdl.Rect{
			X: parentRect.W/2 - fs.Width/2,
			Y: inputRect.Y + inputRect.H + int32(index)*(fs.LineHeight+fs.LineSpacing*2),
			W: inputRect.W,
			H: fs.LineHeight + fs.LineSpacing*2,
		}
		DrawRect(renderer, &entryRect, entryColor)

		txtWidth := fs.Font14.GetStringWidth(fs.FileEntries[res].Name)
		txtRect := sdl.Rect{
			X: entryRect.X + 5,
			Y: entryRect.Y + (entryRect.H-int32(fs.Font14.Size))/2,
			W: txtWidth,
			H: int32(fs.Font14.Size),
		}
		DrawText(renderer, fs.Font14, fs.FileEntries[res].Name, &txtRect, textColor)

		pathWidth := fs.Font12.GetStringWidth(fs.FileEntries[res].FullPath)
		pathRect := sdl.Rect{
			X: entryRect.X + entryRect.W - 5 - pathWidth,
			Y: entryRect.Y + (entryRect.H-int32(fs.Font12.Size))/2,
			W: pathWidth,
			H: int32(fs.Font12.Size),
		}
		DrawText(renderer, fs.Font12, fs.FileEntries[res].FullPath, &pathRect, theme.ResultPathColor)
	}
}
