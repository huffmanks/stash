package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/yarlson/tap"
)

func HandleUninstall(banner string) {
	ctx := context.Background()

	tap.Intro(banner)

	var initialValue *string
	var noValue = "no"

	initialValue = &noValue

	confirmed := tap.Select(ctx, tap.SelectOptions[string]{
		Message:      "Are you sure you want to uninstall?",
		InitialValue: initialValue,
		Options: []tap.SelectOption[string]{
			{Value: "yes", Label: "Yes", Hint: "Requires root privileges"},
			{Value: "no", Label: "No"},
		},
	})

	if confirmed == "no" {
		tap.Outro(Style("🛑 [ABORTED]: stash remains installed.", "orange"))
		os.Exit(0)
	}

	usr, err := user.Current()
	homeDir := os.Getenv("HOME")
	if err == nil {
		homeDir = usr.HomeDir
	}

	configDir := filepath.Join(homeDir, ".config", "stash")
	var candidatePaths []string

	out, err := exec.Command("sh", "-c", "type -a -p stash || which -a stash").Output()
	if err == nil {
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			p := strings.TrimSpace(line)
			if p != "" {
				candidatePaths = append(candidatePaths, p)
			}
		}
	}

	defaultLocalBin := filepath.Join(homeDir, ".local", "bin", "stash")
	if len(candidatePaths) == 0 {
		candidatePaths = append(candidatePaths, defaultLocalBin)
	}

	var targets []string
	var formattedTargets []string
	needsSudo := false
	seen := make(map[string]bool)

	for _, path := range candidatePaths {
		if seen[path] {
			continue
		}
		seen[path] = true

		if _, err := os.Stat(path); err == nil {
			targets = append(targets, path)
			formattedTargets = append(formattedTargets, Style(fmt.Sprintf("• %s", path), "cyan"))

			if !strings.HasPrefix(path, homeDir) {
				needsSudo = true
			}
		}
	}

	if len(targets) == 0 {
		tap.Outro(Style("🛑 [ABORTED]: stash binary not found.", "orange"))
		os.Exit(0)
	}

	targetsBulletList := strings.Join(formattedTargets, "\n      ")
	errorMsg := fmt.Sprintf("❌ %s\n   %s\n      %s\n      %s",
		Style("[ERROR]: Failed to remove the binary.", "red"),
		Style("To finish the cleanup, you can manually remove:", "dim"),
		targetsBulletList,
		Style(fmt.Sprintf("• %s", configDir), "cyan"),
	)

	command := fmt.Sprintf("rm -f %s", strings.Join(targets, " "))

	if needsSudo {
		PromptForSudo(errorMsg, false, command)
	} else {
		for _, path := range targets {
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				tap.Message(errorMsg)
				os.Exit(1)
			}
		}
	}

	var configRemoved bool
	if _, err := os.Stat(configDir); err == nil {
		if err := os.RemoveAll(configDir); err == nil {
			configRemoved = true
		}
	}

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})

	spinner.Start("Uninstalling stash...")
	time.Sleep(time.Millisecond * 1000)

	spinner.Stop("Uninstalling stash...", 0)
	time.Sleep(time.Millisecond * 100)

	summaryBinary := fmt.Sprintf("%s\n      %s", Style("Removed binary location(s):", "bold"), targetsBulletList)
	tap.Message(summaryBinary)

	if configRemoved {
		summaryConfig := fmt.Sprintf("%s\n      %s",
			Style("Removed configuration directory:", "bold"),
			Style(fmt.Sprintf("• %s", configDir), "cyan"),
		)
		tap.Message(summaryConfig)
	}

	tap.Outro("✅ [UNINSTALLED]: stash has been removed successfully.")
	time.Sleep(time.Millisecond * 100)
	os.Exit(0)
}
