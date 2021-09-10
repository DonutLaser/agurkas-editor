package main

import "strings"

func HighlightLineTheme(line []byte, theme *SyntaxTheme) (result []TokenInfo) {
	var sb strings.Builder

	index := 0
	for index < len(line) {
		symbol := line[index]

		if symbol == ' ' {
			result = append(result, TokenInfo{Value: " "})
			index += 1
		} else if symbol == '#' {
			for index < len(line) {
				sb.WriteByte(symbol)
				index += 1

				if index < len(line) {
					symbol = line[index]
				}
			}

			value := sb.String()
			sb.Reset()

			color := theme.BaseColor
			if len(value) == 7 {
				color = hexStringToColor(value)
			}

			result = append(result, TokenInfo{Value: value, Color: color})
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
			if value == "true" || value == "false" {
				color = theme.KeywordColor
			}

			result = append(result, TokenInfo{Value: value, Color: color})
		}
	}

	return
}
