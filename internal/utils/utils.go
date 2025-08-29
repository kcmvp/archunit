package utils

import (
	"log"
	"os/exec"
	"strings"
	"sync"
)

var (
	rootDir string
	module  string
	once    sync.Once
)

func ProjectInfo() (string, string) {
	once.Do(func() {
		cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}:{{.Path}}")
		output, err := cmd.Output()
		if err != nil {
			log.Fatal("Error executing go list command:", err)
		}
		item := strings.Split(strings.TrimSpace(string(output)), ":")
		rootDir = item[0]
		module = item[1]
	})
	return rootDir, module
}
