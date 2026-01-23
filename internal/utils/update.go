package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/yarlson/tap"
)

func HandleUpdate(banner string, force bool) error {

	tap.Message(banner)
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
