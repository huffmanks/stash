package setup

import (
	"fmt"
	"os"
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
	if !c.InstallPackages && len(c.BuildFiles) == 0 {
		fmt.Println("[INFO]: No packages selected and no files to build. Exiting.")
		return nil
	}

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})
	spinner.Start("Initializing...")

	pkgCount := 0
	var filesProcessed []string

	if c.InstallPackages && len(c.SelectedPkgs) > 0 {
		if runtime.GOOS == "darwin" {
			spinner.Message("Checking prerequisities...")
			time.Sleep(time.Second * 2)
			ensureMacOSPrereqs(c.PackageManager, dryRun)
		}

		spinner.Message("Installing packages...")
		time.Sleep(time.Second * 2)

		if err := installSystemPkgs(c, dryRun); err != nil {
			return err
		}
		pkgCount = len(c.SelectedPkgs)
	}

	if len(c.BuildFiles) > 0 {

		zshProcessed := false

		spinner.Message("Building files...")
		time.Sleep(time.Second * 1)

		if slices.Contains(c.BuildFiles, ".zshrc") {
			filesProcessed = append(filesProcessed, ".zshrc")
			zshProcessed = true
		}

		if slices.Contains(c.BuildFiles, ".zprofile") {
			filesProcessed = append(filesProcessed, ".zprofile")
			zshProcessed = true
		}

		if zshProcessed {
			spinner.Message("Creating zsh files...")
			time.Sleep(time.Second * 2)
			buildZshConfigs(c, runtime.GOOS, runtime.GOARCH, dryRun)
		}

		if slices.Contains(c.BuildFiles, ".gitignore") {
			copyGitIgnore(dryRun)
			filesProcessed = append(filesProcessed, ".gitignore")
		}

		if slices.Contains(c.BuildFiles, ".gitconfig") {
			createGitConfig(c, dryRun)
			filesProcessed = append(filesProcessed, ".gitconfig")
		}

		spinner.Message("Finalizing files...")
		time.Sleep(time.Second * 1)
	}

	spinner.Stop("Setup Complete!", 0)
	printSummary(pkgCount, len(c.SelectedPkgs), filesProcessed, dryRun)

	return nil
}

func createGitConfig(c *config.Config, dryRun bool) {
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

	utils.WriteFiles(home+"/.gitconfig", []byte(content), dryRun)
}

func copyGitIgnore(dryRun bool) {
	home, _ := os.UserHomeDir()
	sourcePath := ".dotfiles/.gitignore"
	destPath := path.Join(home, ".gitignore")

	data, err := assets.Files.ReadFile(sourcePath)
	if err != nil {
		fmt.Printf("[WARNING]: Could not find %s to copy\n", sourcePath)
		return
	}

	utils.WriteFiles(destPath, data, dryRun)
}

func printSummary(pkgsInstalled int, pkgsConfigured int, files []string, dryRun bool) {

	if pkgsInstalled > 0 {
		fmt.Printf("📦 [INSTALLED]: %d packages\n", pkgsInstalled)
	} else {
		fmt.Printf("⚙️  [CONFIGURED]: %d packages\n", pkgsConfigured)
	}
	fmt.Printf("🗂️  [FILES]: %d %s\n", len(files), "generated")

	if dryRun {
		if len(files) > 0 {
			var displayNames []string
			for _, f := range files {
				displayNames = append(displayNames, "test"+f)
			}
			fmt.Printf("[DRY-RUN]: The following files were generated in your home directory:\n   %s\n",
				strings.Join(displayNames, ", "))
		}
	}
}
