package main

import (
	"path/filepath"
	"strings"
)

type Project struct {
	Root  string
	Name  string
	Files []string // Paths
}

func ParseProject(data string) (result Project) {
	split := strings.Split(data, "\n")
	exclude := make([]string, 0)

	for _, line := range split {
		key, value := getKeyValue(line, ": ")

		if key == "root" {
			result.Root = value
			result.Name = filepath.Base(value)
		} else if key == "exclude" {
			exclude = append(exclude, strings.Split(value, ",")...)
		}
	}

	result.Files = ReadDirectory(result.Root, exclude)
	return
}

func getKeyValue(line string, separator string) (key string, value string) {
	split := strings.Split(line, separator)
	key = split[0]
	value = split[1]
	return
}
