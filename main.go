package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

func main() {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	checkError(err)
	defer sdl.Quit()

	err = ttf.Init()
	checkError(err)
	defer ttf.Quit()

	window, err := sdl.CreateWindow("Agurkas", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, 800, 600, sdl.WINDOW_RESIZABLE)
	checkError(err)
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	checkError(err)
	defer renderer.Destroy()

	font, err := ttf.OpenFont("consola.ttf", 16)
	checkError(err)
	defer font.Close()

	app := Init()
	input := Input{}

	running := true
	for running {
		input.Clear()

		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch t := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				keycode := t.Keysym.Sym

				switch keycode {
				case sdl.K_LCTRL:
					fallthrough
				case sdl.K_RCTRL:
					input.Modifier |= Modifier_Ctrl
				case sdl.K_LALT:
					fallthrough
				case sdl.K_RALT:
					input.Modifier |= Modifier_Alt
				case sdl.K_LSHIFT:
					fallthrough
				case sdl.K_RSHIFT:
					input.Modifier |= Modifier_Shift
				case sdl.K_RETURN:
					if t.State != sdl.RELEASED {
						input.TypedCharacter = "\n"
					}
				case sdl.K_BACKSPACE:
					if t.State != sdl.RELEASED {
						input.Backspace = true
					}
				case sdl.K_ESCAPE:
					if t.State != sdl.RELEASED {
						input.Escape = true
					}
				}
			case *sdl.TextInputEvent:
				input.TypedCharacter = t.GetText()
			}
		}

		app.Tick(input)

		renderer.SetDrawColor(0, 0, 0, 255)
		renderer.Clear()

		windowWidth, windowHeight := window.GetSize()

		app.Render(renderer, windowWidth, windowHeight)

		renderer.Present()
	}
}
