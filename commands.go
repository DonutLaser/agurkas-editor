package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

func RunCommand(name string, cwd string, args ...string) {
	var stderr bytes.Buffer

	var cmd = exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = &stderr

	if cwd != "" {
		cmd.Dir = cwd
	}

	err := cmd.Run()
	if err != nil {
		fmt.Println("Error:", stderr.String())
	}
}
