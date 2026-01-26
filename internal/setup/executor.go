package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func ExecuteSetup(c *config.Config, dryRun bool) error {

	if c.Operation == "install" && len(c.SelectedPkgs) == 0 {
		tap.Outro("💡 [INFO]: No packages selected to install. Exiting.")
		return nil
	}

	if c.Operation == "configure" && (len(c.BuildFiles) == 0 || (len(c.SelectedPkgs) == 0 && !slices.ContainsFunc(c.BuildFiles, func(f string) bool { return f != ".zshrc" }))) {
		tap.Outro("💡 [INFO]:  No shell files or packages selected to configure. Exiting.")
		return nil
	}

	pkgCount := 0
	extraPkgs := 0
	needsSystemTools := false

	if c.Operation == "install" && len(c.SelectedPkgs) > 0 {

		if !dryRun && !utils.HasSudoPrivilege() {
			tap.Message("Root privileges are required for installing packages.")

			cmd := exec.Command("sh", "-c", "sudo", "-v")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				tap.Outro("❌ [ERROR]: Sudo authentication failed. Exiting.")
				os.Exit(1)
			}
		}

		if runtime.GOOS == "darwin" {
			if !utils.CommandExists("xcode-select") {
				extraPkgs++
				needsSystemTools = true
			}

			if c.PackageManager == "brew" && !utils.CommandExists("brew") {
				extraPkgs++
				needsSystemTools = true
			} else if c.PackageManager == "macports" && !utils.CommandExists("port") {
				extraPkgs++
				needsSystemTools = true
			}
		}

		if runtime.GOOS == "linux" {
			hasPlugins := slices.Contains(c.SelectedPkgs, "zsh-syntax-highlighting") ||
				slices.Contains(c.SelectedPkgs, "zsh-autosuggestions")
			if hasPlugins && !slices.Contains(c.SelectedPkgs, "zsh") {
				c.SelectedPkgs = append(c.SelectedPkgs, "zsh")
				extraPkgs++
			}
		}

		pkgCount = len(c.SelectedPkgs) + extraPkgs

		progress := tap.NewProgress(tap.ProgressOptions{
			Max:   pkgCount,
			Style: "heavy",
			Size:  40,
		})

		progress.Start("Installing packages...")

		var failedPkgs []string

		if needsSystemTools {
			ensureMacOSPrereqs(c.PackageManager, dryRun, progress, &failedPkgs)
		}

		if err := installSystemPkgs(c, dryRun, progress, &failedPkgs); err != nil {
			return err
		}

		time.Sleep(time.Second * 1)
		progress.Stop("🏁 [FINISHED]", 0)
		time.Sleep(time.Second * 1)

		successfulPkgs := c.SelectedPkgs

		if len(failedPkgs) > 0 {
			failedMap := make(map[string]bool)
			for _, p := range failedPkgs {
				failedMap[p] = true
			}

			successfulPkgs = []string{}
			for _, p := range c.SelectedPkgs {
				if !failedMap[p] {
					successfulPkgs = append(successfulPkgs, p)
				}
			}
		}

		installedPkgsMsg := fmt.Sprintf("📦 [INSTALLED]: %d packages\n\n   %s",
			len(successfulPkgs),
			strings.Join(successfulPkgs, ", "))

		if len(failedPkgs) > 0 {

			if len(successfulPkgs) > 0 {
				tap.Message(installedPkgsMsg)
			}

			failedPkgsMsg := fmt.Sprintf("❌ [FAILED]: %d packages\n\n   %s",
				len(failedPkgs),
				strings.Join(failedPkgs, ", "))

			tap.Outro(failedPkgsMsg)
		} else {
			tap.Outro(installedPkgsMsg)
		}
	}

	if c.Operation == "configure" && len(c.BuildFiles) > 0 {

		gitignoreSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})
		gitconfigSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})

		zshProcessed := false

		if slices.Contains(c.BuildFiles, ".zshrc") {
			zshProcessed = true
		}

		if slices.Contains(c.BuildFiles, ".zprofile") {
			zshProcessed = true
		}

		if zshProcessed {
			buildZshConfigs(c, runtime.GOOS, runtime.GOARCH, dryRun)
		}

		if slices.Contains(c.BuildFiles, ".gitignore") {
			gitignoreSpinner.Start("Creating .gitignore...")
			time.Sleep(time.Second * 2)

			copyGitIgnore(dryRun, gitignoreSpinner)

			gitignoreSpinner.Stop("✅ [CREATED]: .gitignore", 0)
		}

		if slices.Contains(c.BuildFiles, ".gitconfig") {
			gitconfigSpinner.Start("Creating .gitconfig...")
			time.Sleep(time.Second * 2)

			createGitConfig(c, dryRun, gitconfigSpinner)

			gitconfigSpinner.Stop("✅ [CREATED]: .gitconfig", 0)
		}

		confMsg := fmt.Sprintf("⚙️  [CONFIGURED]: %d packages\n   🗂️  [FILES]: %d created",
			len(c.SelectedPkgs),
			len(c.BuildFiles),
		)
		tap.Message(confMsg)

		var displayNames []string
		prefix := ""
		if dryRun {
			prefix = "test"
		}

		for _, f := range c.BuildFiles {
			displayNames = append(displayNames, prefix+f)
		}
		outroMsg := fmt.Sprintf("The following files were created in your home directory:\n   %s",
			utils.Style(strings.Join(displayNames, ", "), "cyan"))
		tap.Outro(outroMsg)

	}

	return nil
}

func createGitConfig(c *config.Config, dryRun bool, spinner *tap.Spinner) {
	home, _ := os.UserHomeDir()

	content := fmt.Sprintf(`[init]
	defaultBranch = %s
[user]
	name = %s
	email = %s
[core]
	excludesfile = ~/.gitignore

[http]
	postBuffer = 10485760
`, c.GitBranch, c.GitName, c.GitEmail)

	utils.WriteFiles(home+"/.gitconfig", []byte(content), dryRun, spinner)
}

func copyGitIgnore(dryRun bool, spinner *tap.Spinner) {
	home, _ := os.UserHomeDir()
	sourcePath := ".dotfiles/.gitignore"
	destPath := path.Join(home, ".gitignore")

	data, err := assets.Files.ReadFile(sourcePath)
	if err != nil {
		msg := fmt.Sprintf("⚠️ [WARNING]: Could not find %s to copy", sourcePath)
		spinner.Message(msg)
		return
	}

	utils.WriteFiles(destPath, data, dryRun, spinner)
}
