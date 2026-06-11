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

	PromptForSudo("❌ [ERROR]: sudo authentication failed.", true)

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})
	spinner.Start("Updating...")
	time.Sleep(time.Millisecond * 100)

	scriptURL := "https://raw.githubusercontent.com/huffmanks/stash/main/install.sh"
	shellCmd := fmt.Sprintf("curl -sSL %s | bash -s --", scriptURL)

	if force {
		shellCmd += " --force"
	}

	cmd := exec.Command("sh", "-c", shellCmd)
	err := cmd.Run()

	if err != nil {
		spinner.Stop("❌ [FAILED]: updating stash.", 2)
		time.Sleep(time.Millisecond * 100)
		os.Exit(1)
	}

	time.Sleep(time.Millisecond * 1000)
	spinner.Stop("Updating...", 0)

	time.Sleep(time.Millisecond * 100)
	tap.Outro(fmt.Sprintf("✅ [UPDATED]: successfully to version [%s]", latest))
	time.Sleep(time.Millisecond * 100)

	os.Exit(0)
}
