package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
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
		return make([]byte, 0), path, false
	}

	return data, path, true
}

func CreateDirectory(path string) {
	err := os.Mkdir(path, 0755)
	checkError(err)
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

func ReadDirectory(dirPath string, exclude []string) (result []string) {
	files, err := ioutil.ReadDir(dirPath)
	checkError(err)

	shouldInclude := func(excludes []string, path string) bool {
		for _, ex := range excludes {
			if strings.Contains(path, ex) {
				return false
			}
		}

		return true
	}

	for _, file := range files {
		fullPath := filepath.Join(dirPath, file.Name())
		if !shouldInclude(exclude, fullPath) {
			continue
		}

		if !file.IsDir() {
			result = append(result, fullPath)
		} else {
			result = append(result, ReadDirectory(fullPath, exclude)...)
		}
	}

	return
}

func GetFileNameFromPath(path string) string {
	name := filepath.Base(path)
	if name == "." {
		return ""
	}

	return name
}
