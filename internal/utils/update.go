package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/yarlson/tap"
)

func HandleUpdate(banner string, force bool, latest string) {
	ctx := context.Background()

	tap.Intro(banner)

	if !HasSudoPrivilege() {
		tap.Message("Root privileges are required for updating stash.")

		cmd := exec.Command("sh", "-c", "sudo", "-v")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin

		if err := cmd.Run(); err != nil {
			tap.Outro("❌ [ERROR]: Sudo authentication failed. Exiting.")
			os.Exit(1)
		}
	}

	if !force {
		msg := fmt.Sprintf("Update to version: [%s]?", Style(latest, "bold", "cyan"))
		confirmed := tap.Confirm(ctx, tap.ConfirmOptions{
			Message:      msg,
			InitialValue: false,
		})

		if !confirmed {
			tap.Outro(Style("🛑 [ABORTED]: stash remains installed.", "orange"))
			os.Exit(0)
		}
	}

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})
	spinner.Start(fmt.Sprintf("Updating to version: [%s]...", latest))

	scriptURL := "https://raw.githubusercontent.com/huffmanks/stash/main/install.sh"
	shellCmd := fmt.Sprintf("curl -sSL %s | bash -s --", scriptURL)

	if force {
		shellCmd += " --force"
	}

	cmd := exec.Command("sh", "-c", shellCmd)
	err := cmd.Run()

	if err != nil {
		spinner.Stop("❌ [FAILED]: updating stash.", 2)
		os.Exit(1)
	}

	time.Sleep(time.Second * 1)
	spinner.Stop(fmt.Sprintf("✅ [UPDATED]: successfully to version [%s]", latest), 0)

	os.Exit(0)
}
