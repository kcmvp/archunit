package archunit

import (
	"log"
	"os/exec"
	"strings"
)

func Module() string {
	cmd := exec.Command("go", "list", "-m")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error executing go list command:", err)
	}
	return strings.TrimSpace(string(output))
}
func RootDir() string {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}")
	output, err := cmd.Output()
	if err != nil {
		log.Fatal("Error executing go list command:", err)
	}
	return strings.TrimSpace(string(output))
}
