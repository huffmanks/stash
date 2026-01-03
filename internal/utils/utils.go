package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func DetectPackageManager() string {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("brew"); err == nil {
			return "homebrew"
		}
		if _, err := exec.LookPath("port"); err == nil {
			return "macports"
		}

	case "linux":
		managers := []struct {
			name string
			bin  string
		}{
			{"apt", "apt-get"},
			{"pacman", "pacman"},
			{"dnf", "dnf"},
			{"yum", "yum"},
		}

		for _, pm := range managers {
			if _, err := exec.LookPath(pm.bin); err == nil {
				return pm.name
			}
		}
	}

	return "unknown"
}

func RunCmd(shellCmd string, dryRun bool) {
	if dryRun {
		fmt.Printf("[DRY-RUN] Would execute: %s\n", shellCmd)
		return
	}

	fmt.Printf("Executing: %s\n", shellCmd)
	cmd := exec.Command("sh", "-c", shellCmd)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing command: %v\n", err)
	}
}

func WriteFiles(path string, content []byte, dryRun bool) {
    finalPath := path

    if dryRun {
        dir := filepath.Dir(path)
        name := filepath.Base(path)
        finalPath = filepath.Join(dir, ".test"+name)
        fmt.Printf("[DRY-RUN] Writing test file to: %s\n", finalPath)
    } else {
        fmt.Printf("💾 Writing file to: %s\n", finalPath)
    }

    err := os.WriteFile(finalPath, content, 0644)
    if err != nil {
        fmt.Printf("❌ Error writing %s: %v\n", finalPath, err)
    }
}