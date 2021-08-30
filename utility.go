package main

import (
	"log"

	"github.com/veandco/go-sdl2/sdl"
)

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func expandRect(rect sdl.Rect, amount int32) sdl.Rect {
	return sdl.Rect{
		X: rect.X - amount,
		Y: rect.Y - amount,
		W: rect.W + amount*2,
		H: rect.H + amount*2,
	}
}
