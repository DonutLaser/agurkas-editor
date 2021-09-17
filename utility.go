package main

import (
	"log"
	"strings"

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

func cleanString(text string) string {
	var sb strings.Builder
	for _, symbol := range text {
		if isAlpha(byte(symbol)) {
			sb.WriteByte(byte(symbol))
		}
	}

	return strings.ToLower(sb.String())
}

func isAlpha(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z'
}

func isAlphaNumeric(char byte) bool {
	return char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' || char == '_' || char == '.'
}

func isPunctuation(char byte) bool {
	return (char >= 33 && char < 47) || (char >= 58 && char <= 64) || (char >= 91 && char <= 94) || char == '`' || (char >= 123 && char <= 126) || isWhitespace(char)
}

func isWhitespace(char byte) bool {
	return char == '\n' || char == '\t' || char == ' '
}

func getSymbolPair(char byte) byte {
	switch char {
	case '{':
		return '}'
	case '[':
		return ']'
	case '(':
		return ')'
	case '\'':
		return '\''
	case '"':
		return '"'
	}

	return 0
}

func isStringInArray(array []string, value string) bool {
	for _, item := range array {
		if item == value {
			return true
		}
	}

	return false
}

func hexStringToColor(color string) (result sdl.Color) {
	result.A = 255
	result.R = hexToByte(color[1])<<4 + hexToByte(color[2])
	result.G = hexToByte(color[3])<<4 + hexToByte(color[4])
	result.B = hexToByte(color[5])<<4 + hexToByte(color[6])

	return
}

func stringToBool(str string) bool {
	return str == "true"
}

func hexToByte(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	}

	return 0
}
