package main

import (
	"fmt"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

var golangKeywords = []string{
	"package",
	"import",
	"func",
	"switch",
	"case",
	"default",
	"return",
	"defer",
	"var",
	"for",
	"if",
	"fallthrough",
	"type",
	"const",
	"else",
	"range",
	"append",
	"make",
}

var golangTypes = []string{
	"bool",
	"string",
	"int",
	"int8",
	"int16",
	"int32",
	"int64",
	"uint",
	"uint8",
	"uint16",
	"uint32",
	"uint64",
	"uintptr",
	"byte",
	"rune",
	"float32",
	"float64",
	"complex64",
	"complex128",
}

// var operators = []string{
// 	"+",
// 	"-",
// 	"*",
// 	"/",
// 	"%",
// 	"&",
// 	"|",
// 	"^",
// 	"=",
// 	";",
// 	",",
// 	"<<",
// 	">>",
// 	"==",
// 	"!=",
// 	"<",
// 	"<=",
// 	">",
// 	">=",
// 	"&&",
// 	"||",
// 	"!",
// }

type TokenInfo struct {
	Value string
	Color sdl.Color
}

func HighlightLineGolang(line []byte, theme *SyntaxTheme) (result []TokenInfo) {
	var sb strings.Builder

	index := 0
	for index < len(line) {
		symbol := line[index]

		if symbol == ' ' {
			result = append(result, TokenInfo{Value: " "})
			index += 1
		} else if symbol == '/' {
			for index < len(line) {
				sb.WriteByte(symbol)
				index += 1

				if index < len(line) {
					symbol = line[index]
				}
			}

			result = append(result, TokenInfo{Value: sb.String(), Color: theme.CommentColor})
			sb.Reset()
		} else if symbol == '"' {
			index += 1
			symbol = line[index]

			for symbol != '"' && index < len(line) {
				sb.WriteByte(symbol)
				index += 1

				if index < len(line) {
					symbol = line[index]
				}
			}

			index += 1

			result = append(result, TokenInfo{Value: fmt.Sprintf("\"%s\"", sb.String()), Color: theme.StringColor})
			sb.Reset()
		} else if symbol == '\'' {
			index += 1
			symbol = line[index]

			for symbol != '\'' && index < len(line) {
				sb.WriteByte(symbol)
				index += 1

				if index < len(line) {
					symbol = line[index]
				}
			}

			index += 1

			result = append(result, TokenInfo{Value: fmt.Sprintf("\"%s\"", sb.String()), Color: theme.StringColor})
			sb.Reset()
		} else if isAlphaNumeric(symbol) {
			for isAlphaNumeric(symbol) && index < len(line) {
				sb.WriteByte(symbol)
				index += 1

				if index < len(line) {
					symbol = line[index]
				}
			}

			value := sb.String()
			sb.Reset()

			color := theme.BaseColor
			if isStringInArray(golangKeywords, value) {
				color = theme.KeywordColor
			} else if isStringInArray(golangTypes, value) {
				color = theme.TypeColor
			}

			result = append(result, TokenInfo{Value: value, Color: color})
		} else {
			result = append(result, TokenInfo{Value: string(symbol), Color: theme.OperatorColor})
			index += 1
		}
	}

	return
}

func isAlphaNumeric(symbol byte) bool {
	return symbol >= 'a' && symbol <= 'z' || symbol >= 'A' && symbol <= 'Z' || symbol >= '0' && symbol <= '9' || symbol == '_' || symbol == '.'
}

func isStringInArray(array []string, value string) bool {
	for _, item := range array {
		if item == value {
			return true
		}
	}

	return false
}
