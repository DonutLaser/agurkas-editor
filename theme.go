package main

import (
	"log"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type StatusBarTheme struct {
	BackgroundColor sdl.Color

	NormalColor         sdl.Color
	NormalTextColor     sdl.Color
	InsertColor         sdl.Color
	InsertTextColor     sdl.Color
	VisualColor         sdl.Color
	VisualTextColor     sdl.Color
	VisualLineColor     sdl.Color
	VisualLineTextColor sdl.Color

	TextColor  sdl.Color
	DirtyColor sdl.Color
}

type BufferTheme struct {
	BackgroundColor    sdl.Color
	LineHighlightColor sdl.Color
	TextColor          sdl.Color

	CursorColor               sdl.Color
	CursorColorMatchModeColor bool
}

type GutterTheme struct {
	BackgroundColor    sdl.Color
	LineHighlightColor sdl.Color

	LineNumberInactiveColor  sdl.Color
	LineNumberActiveColor    sdl.Color
	LineNumberMatchModeColor bool
}

type FileSearchTheme struct {
	InputBackgroundColor sdl.Color
	BorderColor          sdl.Color
	CursorColor          sdl.Color
	InputTextColor       sdl.Color

	ResultBackgroundColor sdl.Color
	ResultActiveColor     sdl.Color
	ResultNameColor       sdl.Color
	ResultNameActiveColor sdl.Color
	ResultPathColor       sdl.Color
	ResultPathActiveColor sdl.Color
}

type Theme struct {
	StatusBar  StatusBarTheme
	Buffer     BufferTheme
	Gutter     GutterTheme
	FileSearch FileSearchTheme
}

func ParseTheme(path string) (result Theme) {
	data, _, success := OpenFile(path)
	if !success {
		return
	}

	split := strings.Split(string(data), "\n")

	for _, line := range split {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}

		key, value := getKeyValue(l, " ")
		if strings.HasPrefix(key, "statusbar") {
			parseStatusBar(key, value, &result.StatusBar)
		} else if strings.HasPrefix(key, "buffer") {
			parseBuffer(key, value, &result.Buffer)
		} else if strings.HasPrefix(key, "gutter") {
			parseGutter(key, value, &result.Gutter)
		} else if strings.HasPrefix(key, "fs") {
			parseFileSearch(key, value, &result.FileSearch)
		}
	}

	return
}

func (theme *Theme) GetColorForMode(mode Mode) sdl.Color {
	switch mode {
	case Mode_Normal:
		return theme.StatusBar.NormalColor
	case Mode_Insert:
		return theme.StatusBar.InsertColor
	case Mode_Visual:
		return theme.StatusBar.VisualColor
	case Mode_VisualLine:
		return theme.StatusBar.VisualLineColor
	}

	return sdl.Color{}
}

func (theme *StatusBarTheme) GetColorForMode(mode Mode) sdl.Color {
	switch mode {
	case Mode_Normal:
		return theme.NormalColor
	case Mode_Insert:
		return theme.InsertColor
	case Mode_Visual:
		return theme.VisualColor
	case Mode_VisualLine:
		return theme.VisualLineColor
	}

	return sdl.Color{}
}

func (theme *StatusBarTheme) GetTextColorForMode(mode Mode) sdl.Color {
	switch mode {
	case Mode_Normal:
		return theme.NormalTextColor
	case Mode_Insert:
		return theme.InsertTextColor
	case Mode_Visual:
		return theme.VisualTextColor
	case Mode_VisualLine:
		return theme.VisualLineTextColor
	}

	return sdl.Color{}
}

func parseStatusBar(key string, value string, theme *StatusBarTheme) {
	switch key {
	case "statusbar_bg_color":
		theme.BackgroundColor = hexStringToColor(value)
	case "statusbar_normal_color":
		theme.NormalColor = hexStringToColor(value)
	case "statusbar_normal_txt_color":
		theme.NormalTextColor = hexStringToColor(value)
	case "statusbar_insert_color":
		theme.InsertColor = hexStringToColor(value)
	case "statusbar_insert_txt_color":
		theme.InsertTextColor = hexStringToColor(value)
	case "statusbar_visual_color":
		theme.VisualColor = hexStringToColor(value)
	case "statusbar_visual_txt_color":
		theme.VisualTextColor = hexStringToColor(value)
	case "statusbar_vline_color":
		theme.VisualLineColor = hexStringToColor(value)
	case "statusbar_vline_txt_color":
		theme.VisualLineTextColor = hexStringToColor(value)
	case "statusbar_txt_color":
		theme.TextColor = hexStringToColor(value)
	case "statusbar_dirty_color":
		theme.DirtyColor = hexStringToColor(value)
	default:
		log.Printf("Unsupported property for status bar theme: %s = %s", key, value)
	}
}

func parseBuffer(key string, value string, theme *BufferTheme) {
	switch key {
	case "buffer_bg_color":
		theme.BackgroundColor = hexStringToColor(value)
	case "buffer_line_highlight_color":
		theme.LineHighlightColor = hexStringToColor(value)
	case "buffer_txt_color":
		theme.TextColor = hexStringToColor(value)
	case "buffer_cursor_color_match_mode":
		theme.CursorColorMatchModeColor = stringToBool(value)
	case "buffer_cursor_color":
		theme.CursorColor = hexStringToColor(value)
	default:
		log.Printf("Unsupported property for status bar theme: %s = %s", key, value)
	}
}

func parseGutter(key string, value string, theme *GutterTheme) {
	switch key {
	case "gutter_bg_color":
		theme.BackgroundColor = hexStringToColor(value)
	case "gutter_line_highlight_color":
		theme.LineHighlightColor = hexStringToColor(value)
	case "gutter_line_number_inactive_color":
		theme.LineNumberInactiveColor = hexStringToColor(value)
	case "gutter_line_number_color_match_mode":
		theme.LineNumberMatchModeColor = stringToBool(value)
	case "gutter_line_number_active_color":
		theme.LineNumberActiveColor = hexStringToColor(value)
	default:
		log.Printf("Unsupported property for status bar theme: %s = %s", key, value)
	}
}

func parseFileSearch(key string, value string, theme *FileSearchTheme) {
	switch key {
	case "fs_input_bg_color":
		theme.InputBackgroundColor = hexStringToColor(value)
	case "fs_border_color":
		theme.BorderColor = hexStringToColor(value)
	case "fs_input_txt_color":
		theme.InputTextColor = hexStringToColor(value)
	case "fs_cursor_color":
		theme.CursorColor = hexStringToColor(value)
	case "fs_result_bg_color":
		theme.ResultBackgroundColor = hexStringToColor(value)
	case "fs_result_active_bg_color":
		theme.ResultActiveColor = hexStringToColor(value)
	case "fs_result_name_color":
		theme.ResultNameColor = hexStringToColor(value)
	case "fs_result_name_active_color":
		theme.ResultNameActiveColor = hexStringToColor(value)
	case "fs_result_path_color":
		theme.ResultPathColor = hexStringToColor(value)
	case "fs_result_path_active_color":
		theme.ResultPathActiveColor = hexStringToColor(value)
	default:
		log.Printf("Unsupported property for status bar theme: %s = %s", key, value)
	}
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
