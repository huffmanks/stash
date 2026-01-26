package utils

import (
	"context"
	"fmt"
	"os"
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

	binaryPath := "/usr/local/bin/stash"

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		tap.Outro("🛑 [ABORTED]: stash is not found in /usr/local/bin.")
		os.Exit(0)
	}

	errorMsg := fmt.Sprintf("❌ %s\n   %s\n      %s\n      %s", Style("[ERROR]: Failed to remove the binary.", "red"), Style("To finish the cleanup, you can manually remove:", "dim"), Style("• /usr/local/bin/stash", "cyan"), Style("• ~/.config/stash", "cyan"))
	command := fmt.Sprintf("rm %s", binaryPath)
	PromptForSudo(errorMsg, command)

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})
	spinner.Start("Uninstalling stash...")
	time.Sleep(time.Millisecond * 1300)
	spinner.Stop("Uninstalling stash...", 0)

	time.Sleep(time.Millisecond * 200)
	tap.Outro("✅ [UNINSTALLED]: stash has been removed successfully.")
	os.Exit(0)
}
