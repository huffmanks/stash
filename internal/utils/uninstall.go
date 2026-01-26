package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/yarlson/tap"
)

func HandleUninstall(banner string) {
	ctx := context.Background()

	tap.Intro(banner)

	confirmed := tap.Confirm(ctx, tap.ConfirmOptions{
		Message:      "Are you sure you want to uninstall?",
		InitialValue: false,
	})

	if !confirmed {
		tap.Outro(Style("🛑 [ABORTED]: stash remains installed.", "orange"))
		os.Exit(0)
	}

	if !HasSudoPrivilege() {
		tap.Message("Root privileges are required for updating stash.")

		cmd := exec.Command("sh", "-c", "sudo", "-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			msg := fmt.Sprintf("❌ %s\n   %s\n      %s\n      %s", Style("[ERROR]: Sudo authentication failed.", "red"), Style("To finish the cleanup, you can manually remove:", "dim"), Style("• /usr/local/bin/stash", "cyan"), Style("• ~/.config/stash", "cyan"))
			tap.Outro(msg)
			os.Exit(1)
		}
	}

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})

	spinner.Start("Uninstalling stash...")
	time.Sleep(time.Second * 2)

	err := os.Remove("/usr/local/bin/stash")
	if err != nil {
		msg := fmt.Sprintf("❌ %s\n   %s\n      %s\n      %s", Style("[ERROR]: Failed to remove the binary.", "red"), Style("To finish the cleanup, you can manually remove:", "dim"), Style("• /usr/local/bin/stash", "cyan"), Style("• ~/.config/stash", "cyan"))
		spinner.Stop(msg, 2)

		os.Exit(1)
	}

	spinner.Stop("✅ [UNINSTALLED]: stash has been removed successfully.", 0)
	os.Exit(0)
}
