package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/yarlson/tap"
)

func HandleUpdate(banner string, force bool, latest string) error {
	ctx := context.Background()

	tap.Intro(banner)

	if !force {
		msg := fmt.Sprintf("Update to version: [%s]?", Style(latest, "bold", "cyan"))
		confirmed := tap.Confirm(ctx, tap.ConfirmOptions{
			Message:      msg,
			InitialValue: false,
		})

		if !confirmed {
			tap.Outro(Style("Aborted. stash remains installed.", "orange"))
			return nil
		}
	} else {
		msg := fmt.Sprintf("Updating to version: [%s]...", latest)
		tap.Message(msg)
	}

	scriptURL := "https://raw.githubusercontent.com/huffmanks/stash/main/install.sh"

	shellCmd := fmt.Sprintf("curl -sSL %s | bash -s --", scriptURL)

	if force {
		shellCmd += " --force"
	}

	cmd := exec.Command("sh", "-c", shellCmd)

	cmd.Stdin = os.Stdin

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	stream := tap.NewStream(tap.StreamOptions{ShowTimer: true})

	stream.Start(outputStr)

	if err != nil {
		stream.Stop("Update failed.", 2)
		return fmt.Errorf("Update failed: %w", err)
	}

	stream.Stop("Update complete!", 0)

	return nil
}
