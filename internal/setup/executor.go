package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
)


func ExecuteSetup(c *config.Config, dryRun bool) error {
	if !c.InstallPackages && len(c.BuildFiles) == 0 {
		fmt.Println("ℹ️  No packages selected and no files to build. Exiting.")
		return nil
	}

	pkgCount := 0
	fileCount := 0

	if c.InstallPackages && len(c.SelectedPkgs) > 0 {
		if runtime.GOOS == "darwin" {
			ensureMacOSPrereqs(c.PackageManager, dryRun)
		}

		if err := installSystemPkgs(c, dryRun); err != nil {
			return err
		}
		pkgCount = len(c.SelectedPkgs)
	}

	if len(c.BuildFiles) > 0 {
		if slices.Contains(c.BuildFiles, ".gitconfig") {
			createGitConfig(c, dryRun)
			fileCount++
		}

		zshProcessed := false
		if slices.Contains(c.BuildFiles, ".zshrc") {
			fileCount++
			zshProcessed = true
		}
		if slices.Contains(c.BuildFiles, ".zprofile") {
			fileCount++
			zshProcessed = true
		}

		if zshProcessed {
			buildZshConfigs(c, runtime.GOOS, runtime.GOARCH, dryRun)
		}

		if slices.Contains(c.BuildFiles, ".gitignore") {
			copyGitIgnore(dryRun)
			fileCount++
		}
	}

	printSummary(pkgCount, fileCount, dryRun)

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
    destPath := filepath.Join(home, ".gitignore")

    data, err := os.ReadFile(sourcePath)
    if err != nil {
        fmt.Printf("⚠️  Warning: Could not find %s to copy\n", sourcePath)
        return
    }

    utils.WriteFiles(destPath, data, dryRun)
}

func printSummary(pkgs, files int, dryRun bool) {
	modeText := "Installed/Built"
	if dryRun {
		modeText = "Would install/build"
	}

	fmt.Printf("\n✨ Setup Complete!\n")
	fmt.Printf("📦 Packages: %d %s\n", pkgs, modeText)
	fmt.Printf("📄 Files:    %d %s\n", files, modeText)

	if dryRun {
		fmt.Println("\n⚠️  Reminder: This was a dry run. .test.files were generated in your home directory.")
	}
}