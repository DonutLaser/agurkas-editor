package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type Mode string

const (
	Mode_Normal     Mode = "Normal"
	Mode_Insert     Mode = "Insert"
	Mode_Visual     Mode = "Visual"
	Mode_VisualLine Mode = "Visual Line"
)

type App struct {
	Font           *ttf.Font
	CharacterWidth int32
	LineHeight     int32

	Buffer Buffer

	Mode Mode
}

func Init() (result App) {
	result = App{}

	font, err := ttf.OpenFont("consola.ttf", 16)
	checkError(err)
	result.Font = font
	result.CharacterWidth = GetCharacterWidth(font)
	result.LineHeight = 16

	result.Mode = Mode_Normal
	result.Buffer = CreateBuffer(result.LineHeight)

	return
}

func (app *App) handleInputNormal(input Input) {
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
	case "a":
		app.Mode = Mode_Insert
		app.Buffer.MoveRight()
	case "x":
		app.Buffer.RemoveAfter()
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

func (app *App) Render(renderer *sdl.Renderer, windowWidth int32, windowHeight int32) {
	app.Buffer.Render(renderer, app.Font, app.Mode, app.CharacterWidth)

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
