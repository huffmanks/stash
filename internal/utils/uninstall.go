package utils

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
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
	var homeDir string
	if err == nil {
		homeDir = usr.HomeDir
	} else {
		homeDir = os.Getenv("HOME")
	}

	localBinPath := filepath.Join(homeDir, ".local", "bin", "stash")
	systemBinPath := "/usr/local/bin/stash"

	var binaryPath string
	if _, err := os.Stat(localBinPath); err == nil {
		binaryPath = localBinPath
	} else if _, err := os.Stat(systemBinPath); err == nil {
		binaryPath = systemBinPath
	}

	if binaryPath == "" {
		tap.Outro(Style("🛑 [ABORTED]: stash is not found in ~/.local/bin or /usr/local/bin.", "orange"))
		os.Exit(0)
	}

	needsSudo := binaryPath == systemBinPath

	errorMsg := fmt.Sprintf("❌ %s\n   %s\n      %s\n      %s",
		Style("[ERROR]: Failed to remove the binary.", "red"),
		Style("To finish the cleanup, you can manually remove:", "dim"),
		Style(fmt.Sprintf("• %s", binaryPath), "cyan"),
		Style(fmt.Sprintf("• %s/.config/stash", homeDir), "cyan"),
	)

	command := fmt.Sprintf("rm %s", binaryPath)

	if needsSudo {
		PromptForSudo(errorMsg, false, command)
	} else {
		if err := os.Remove(binaryPath); err != nil {
			fmt.Println(errorMsg)
			os.Exit(1)
		}
	}

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})

	spinner.Start("Uninstalling stash...")
	time.Sleep(time.Millisecond * 1000)

	spinner.Stop("Uninstalling stash...", 0)
	time.Sleep(time.Millisecond * 100)

	tap.Outro("✅ [UNINSTALLED]: stash has been removed successfully.")
	time.Sleep(time.Millisecond * 100)
	os.Exit(0)
}
