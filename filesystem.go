package main

import (
	"os"
	"strings"

	"github.com/sqweek/dialog"
)

func SaveFile(path string, data []string) bool {
	if path == "" {
		newPath, err := dialog.File().Filter("Text file", "txt").Save()
		if err != dialog.ErrCancelled {
			checkError(err)
		} else {
			return false
		}

		path = newPath
	}

	file, err := os.Create(path)
	if err != nil {
		return false
	}
	defer file.Close()

	file.WriteString(strings.Join(data, "\n"))
	file.Sync()

	return true
}
