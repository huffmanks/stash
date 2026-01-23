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

	tap.Message(banner)

	confirmed := tap.Confirm(ctx, tap.ConfirmOptions{
		Message:      "Are you sure you want to uninstall?",
		InitialValue: false,
	})

	if !confirmed {
		tap.Outro("Aborted. stash remains installed.")
		return
	}

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})

	spinner.Start("Uninstalling stash...")
	time.Sleep(time.Second * 2)

	err := os.Remove("/usr/local/bin/stash")
	if err != nil {
		msg := fmt.Sprintf("Error removing binary: %v\n   Note: You may need to run this as root: sudo stash --uninstall", err)
		spinner.Stop(msg, 2)

		return
	}

	spinner.Stop("stash has been removed successfully.", 0)
}
