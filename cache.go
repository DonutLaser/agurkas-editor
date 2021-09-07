package main

import (
	"fmt"
	"strings"
)

type Cache struct {
	Path     string
	Projects []string
}

func ParseCache(dir string) (result Cache) {
	data, path, success := OpenFile(fmt.Sprintf("%s/cache.acache", dir))
	if !success {
		CreateDirectory(dir)
		SaveFile(path, make([]string, 0))
		result.Path = path
		return
	}

	result.Path = path

	split := strings.Split(string(data), "\n")

	for _, line := range split {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}

		key, value := getKeyValue(l, " ")
		if key == "project" {
			result.Projects = append(result.Projects, value)
		}
	}

	return
}

func (cache *Cache) Write(key string, value string) {
	data, _, success := OpenFile(cache.Path)
	if !success {
		return
	}

	lines := strings.Split(string(data), "\n")
	lines = append(lines, fmt.Sprintf("%s %s", key, value))

	SaveFile(cache.Path, lines)

	if key == "project" {
		cache.Projects = append(cache.Projects, value)
	}
}
