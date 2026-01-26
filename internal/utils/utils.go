package utils

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

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
		msg := fmt.Sprintf("___ [DRY-RUN]: Would execute: %s ___", shellCmd)
		progress.Message(msg)
		return nil
	}

	if strings.HasPrefix(shellCmd, "sudo") {
		if !HasSudoPrivilege() {
			progress.Message("⚠️ [SKIPPED]: No sudo session active.")
			return fmt.Errorf("skipping: sudo required")
		}
	}

	executingMsg := fmt.Sprintf("🪓 [EXECUTING]: %s", shellCmd)
	progress.Message(executingMsg)
	cmd := exec.Command("sh", "-c", shellCmd)

	_, err := cmd.CombinedOutput()

	if err != nil {
		progress.Message(fmt.Sprintf("❌ [ERROR]: %s", shellCmd))
		return err
	}

	return nil
}

func WriteFiles(fileName string, content []byte, dryRun bool, spinner *tap.Spinner) {

	home, _ := os.UserHomeDir()
	finalPath := filepath.Join(home, fileName)

	if dryRun {
		finalPath = filepath.Join(home, "test_"+fileName)
		msg := fmt.Sprintf("___ [DRY-RUN]: Writing test file to: %s  ___", finalPath)
		spinner.Message(msg)
	} else {
		if _, err := os.Stat(finalPath); err == nil {
			now := time.Now()
			timestamp := now.Format("20060102_150405")

			bakDir := filepath.Join(home, ".config", "stash")
			if err := os.MkdirAll(bakDir, 0755); err != nil {
				spinner.Message(fmt.Sprintf("❌ [ERROR]: Could not create backup dir: %v", err))
			}

			bakFileName := fmt.Sprintf("bak_%s_%s", timestamp, fileName)
			bakPath := filepath.Join(bakDir, bakFileName)

			if err := os.Rename(finalPath, bakPath); err == nil {
				os.Chtimes(bakPath, now, now)
				msg := fmt.Sprintf("🚚 [MOVED]: Existing file moved to %s", bakPath)
				spinner.Message(msg)
			} else {
				msg := fmt.Sprintf("⚠️ %s %v", Style("[WARNING]: Could not backup existing file:", "orange"), err)
				spinner.Message(msg)
			}
		}

		msg := fmt.Sprintf("📝 [WRITING]: file to %s", finalPath)
		spinner.Message(msg)
	}

	err := os.WriteFile(finalPath, content, 0644)
	if err != nil {
		errMsg := fmt.Sprintf("❌ [ERROR]: writing %s - %v", finalPath, err)
		spinner.Message(errMsg)
	}
}

var pkgOverrides = map[string]map[string]string{
	"fd": {
		"apt": "fd-find",
		"dnf": "fd-find",
	},
	"java-android-studio": {
		"homebrew": "--cask zulu@17",
		"macports": "openjdk17-zulu",
	},
}

func ResolvePkgName(pm, pkg string) string {
	if overrides, found := pkgOverrides[pkg]; found {
		if specificName, ok := overrides[pm]; ok {
			return specificName
		}
	}
	return pkg
}

func CommandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func HasSudoPrivilege() bool {
	err := exec.Command("sudo", "-n", "true").Run()
	return err == nil
}

func GetLatestVersion(version string) string {
	cmd := `curl -s "https://api.github.com/repos/huffmanks/stash/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'`
	out, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		return version
	}
	return string(bytes.TrimSpace(out))
}

var styles = map[string]string{
	"reset":  "\033[0m",
	"bold":   "\033[1m",
	"dim":    "\033[2m",
	"unbold": "\033[22m",
	"red":    "\033[31m",
	"green":  "\033[32m",
	"orange": "\033[33m",
	"cyan":   "\033[36m",
}

func Style(s string, keys ...string) string {
	var builder strings.Builder

	for _, key := range keys {
		if code, ok := styles[key]; ok {
			builder.WriteString(code)
		}
	}

	builder.WriteString(s)
	builder.WriteString(styles["reset"])

	return builder.String()
}
