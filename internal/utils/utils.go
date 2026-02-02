package utils

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/config"
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
		msg := fmt.Sprintf(Style("___ [DRY_RUN]: Would execute: %s ___", "orange"), shellCmd)
		progress.Message(msg)
		return nil
	}

	if strings.HasPrefix(shellCmd, "sudo") {
		PromptForSudo("❌ [ERROR]: sudo authentication failed.", "true", true)
	}

	executingMsg := fmt.Sprintf("🪓 [EXECUTING]: %s", shellCmd)
	progress.Message(executingMsg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", shellCmd)
	cmd.Stdin = nil

	output, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		msg := fmt.Sprintf("❌ [ERROR]: %s took to long and timed out after 30s", shellCmd)
		progress.Message(msg)
		return fmt.Errorf("%s", msg)
	}

	if err != nil {
		cleanOut := strings.TrimSpace(string(output))
		lines := strings.Split(cleanOut, "\n")
		shortErr := lines[len(lines)-1]

		if len(shortErr) > 100 {
			shortErr = shortErr[:97] + "..."
		}

		progress.Message(fmt.Sprintf("❌ [ERROR]: %s\n%s", shellCmd, shortErr))
		return err
	}

	return nil
}

func WriteFiles(fileName string, content []byte, dryRun bool, spinner *tap.Spinner) {

	home, _ := os.UserHomeDir()
	finalPath := filepath.Join(home, fileName)

	if dryRun {
		finalPath = filepath.Join(home, "test_"+fileName)
		msg := fmt.Sprintf(Style("___ [DRY_RUN]: Writing test file to: %s  ___", "orange"), finalPath)
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

func DeleteFiles(dryRun bool, spinner *tap.Spinner) config.DeleteResult {
	home, _ := os.UserHomeDir()
	stashDir := filepath.Join(home, ".config", "stash")
	pattern := filepath.Join(stashDir, "bak*")

	res := config.DeleteResult{}

	files, err := filepath.Glob(pattern)
	if err != nil {
		spinner.Message(fmt.Sprintf("❌ [ERROR]: Glob pattern failed: %v", err))
		return res
	}

	if len(files) == 0 {
		spinner.Message("‼️ [EMPTY]: No backup files found to delete.")
		return res
	}

	for _, f := range files {
		base := filepath.Base(f)
		time.Sleep(time.Millisecond * 100)

		if dryRun {
			msg := fmt.Sprintf(Style("___ [DRY_RUN]: Would delete: %s ___", "orange"), base)
			spinner.Message(msg)
			res.Deleted = append(res.Deleted, base)
			continue
		}

		err := os.Remove(f)
		if err != nil {
			spinner.Message(fmt.Sprintf("❌ [ERROR]: %s", base))
			res.Failed = append(res.Failed, fmt.Sprintf("%s (%v)", base, err))
		} else {
			spinner.Message(fmt.Sprintf("🗑️  [DELETED]: %s", base))
			res.Deleted = append(res.Deleted, base)
		}
	}

	return res
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

func hasSudoPrivilege() bool {
	err := exec.Command("sudo", "-n", "true").Run()
	return err == nil
}

func PromptForSudo(errorMsg string, command string, useSkipCmd ...bool) {
	ctx := context.Background()

	skipCmd := false
	if len(useSkipCmd) > 0 {
		skipCmd = useSkipCmd[0]
	}

	time.Sleep(100 * time.Millisecond)

	if hasSudoPrivilege() {
		if !skipCmd {
			_ = exec.Command("sudo", "-S", "sh", "-c", command).Run()
		}
		return
	}

	tap.Message("Authenticate to continue...")

	maxRetries := 3
	for i := range maxRetries {
		password := tap.Password(ctx, tap.PasswordOptions{
			Message: "Enter sudo password:",
		})

		sudoCmd := exec.Command("sudo", "-S", "sh", "-c", command)

		stdin, err := sudoCmd.StdinPipe()
		if err != nil {
			if hasSudoPrivilege() {
				return
			}
			continue
		}

		go func() {
			defer stdin.Close()
			fmt.Fprintln(stdin, password)
		}()

		if err := sudoCmd.Run(); err != nil {

			time.Sleep(100 * time.Millisecond)
			if hasSudoPrivilege() {
				return
			}
			if i < maxRetries-1 {
				tap.Message(Style("⚠️  Invalid password, try again.", "orange"))
				continue
			}

			tap.Outro(errorMsg)
			os.Exit(1)
		}

		time.Sleep(100 * time.Millisecond)
		return
	}
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

func IsAndroid() bool {
	if version, err := os.ReadFile("/proc/version"); err == nil {
		if strings.Contains(strings.ToLower(string(version)), "android") {
			return true
		}
	}

	return false
}
