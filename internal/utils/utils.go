package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/yarlson/tap"
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
		}

		for _, pm := range managers {
			if _, err := exec.LookPath(pm.bin); err == nil {
				return pm.name
			}
		}
	}

	return "unknown"
}

func RunCmd(shellCmd string, dryRun bool, progress *tap.Progress) error {
	if dryRun {
		msg := fmt.Sprintf("___ [DRY-RUN]: Would execute: %s", shellCmd)
		progress.Message(msg)
		return nil
	}

	if strings.HasPrefix(shellCmd, "sudo") {
		if !HasSudoPrivilege() {
			progress.Message("⏭️ [SKIPPING]: No sudo session active.")
			return fmt.Errorf("skipping: sudo required")
		}
	}

	executingMsg := fmt.Sprintf("🪓 [EXECUTING]: %s", shellCmd)
	progress.Message(executingMsg)
	cmd := exec.Command("sh", "-c", shellCmd)

	_, err := cmd.CombinedOutput()

	if err != nil {
		progress.Message(fmt.Sprintf("⛔️ [ERROR]: %s", shellCmd))
		return err
	}

	return nil
}

func WriteFiles(path string, content []byte, dryRun bool, spinner *tap.Spinner) {
	finalPath := path

	if dryRun {
		dir := filepath.Dir(path)
		name := filepath.Base(path)
		finalPath = filepath.Join(dir, ".test"+name)
		msg := fmt.Sprintf("___ [DRY-RUN]: Writing test file to: %s", finalPath)
		spinner.Message(msg)
	} else {
		msg := fmt.Sprintf("📝 [WRITING]: file to %s", finalPath)
		spinner.Message(msg)
	}

	err := os.WriteFile(finalPath, content, 0644)
	if err != nil {
		errMsg := fmt.Sprintf("⛔️ [ERROR]: writing %s - %v", finalPath, err)
		spinner.Message(errMsg)
	}
}

func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func HasSudoPrivilege() bool {
	err := exec.Command("sudo", "-n", "true").Run()
	return err == nil
}
