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
	Font        *Font

	FileEntries      []FileSearchEntry
	FoundEntries     []int // Array of indexes into file entries array
	MaxSearchResults int32

	CloseCallback func(string)

	altWasPressed bool
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

func CreateFileSearch(lineHeight int32, font *Font) (result FileSearch) {
	result.Cursor.Width = 2
	result.Cursor.Height = lineHeight
	result.Cursor.Advance = int32(font.CharacterWidth)
	result.Width = 500
	result.LineHeight = lineHeight
	result.LineSpacing = (lineHeight - int32(font.Size)) / 2
	result.Font = font

	result.MaxSearchResults = 5

	return
}

func (fs *FileSearch) Open(availableFiles []FileSearchEntry, onClose func(string)) {
	fs.SelectionIndex = -1
	fs.Cursor.Column = 0
	fs.SearchQuery.Reset()

	fs.FileEntries = availableFiles
	fs.FoundEntries = make([]int, 0)
	fs.CloseCallback = onClose
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
		fs.altWasPressed = false
		if len(fs.FoundEntries) > 0 {
			fs.Submit()
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

func (fs *FileSearch) Render(renderer *sdl.Renderer, parentRect *sdl.Rect) {
	// @TODO (!important) border around the whole popup
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

	DrawRect(renderer, &borderRect, sdl.Color{R: 48, G: 48, B: 48, A: 255})
	DrawRect(renderer, &inputRect, sdl.Color{R: 13, G: 14, B: 16, A: 255})
	fs.Cursor.Render(renderer, inputRect)

	query := fs.SearchQuery.String()
	if len(query) > 0 {
		queryWidth, queryHeight := fs.Font.GetStringSize(query)
		queryRect := sdl.Rect{
			X: inputRect.X + 5,
			Y: inputRect.Y + (inputRect.H-queryHeight)/2,
			W: queryWidth,
			H: queryHeight,
		}
		DrawText(renderer, fs.Font, query, &queryRect, sdl.Color{R: 255, G: 255, B: 255, A: 255})
	}

	defaultBgColor := sdl.Color{R: 9, G: 9, B: 1, A: 255}
	defaultTextColor := sdl.Color{R: 137, G: 145, B: 162, A: 255}

	activeBgColor := sdl.Color{R: 34, G: 35, B: 38, A: 255}
	activeTextColor := sdl.Color{R: 255, G: 255, B: 255, A: 255}

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

		txtWidth, txtHeight := fs.Font.GetStringSize(fs.FileEntries[res].Name)
		txtRect := sdl.Rect{
			X: entryRect.X + 5,
			Y: entryRect.Y + (entryRect.H-txtHeight)/2,
			W: txtWidth,
			H: txtHeight,
		}
		DrawText(renderer, fs.Font, fs.FileEntries[res].Name, &txtRect, textColor)
	}
}
