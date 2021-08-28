package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/sqweek/dialog"
)

func SaveFile(path string, data []string) (string, bool) {
	if path == "" {
		newPath, err := dialog.File().Filter("All files (*.*)", "*").Filter("Text file (*.txt)", "txt").Filter("Go file (*.go)", "go").Save()
		if err != dialog.ErrCancelled {
			checkError(err)
		} else {
			return "", false
		}

		path = newPath
	}

	file, err := os.Create(path)
	if err != nil {
		return "", false
	}
	defer file.Close()

	file.WriteString(strings.Join(data, "\n"))
	file.Sync()

	return path, true
}

func OpenFile(path string) ([]byte, string, bool) {
	if path == "" {
		newPath, err := dialog.File().Filter("All files (*.*)", "*").Filter("Text file (*.txt)", "txt").Filter("Go file (*.go)", "go").Load()
		if err != dialog.ErrCancelled {
			checkError(err)
		} else {
			return make([]byte, 0), "", false
		}

		path = newPath
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return make([]byte, 0), "", false
	}

	return data, path, true
}

func SelectDirectory() (string, bool) {
	path, err := dialog.Directory().Title("Select folder...").Browse()
	if err != dialog.ErrCancelled {
		checkError((err))
	} else {
		return "", false
	}

	return path, true
}
