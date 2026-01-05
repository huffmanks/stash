package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
)

func ExecuteSetup(c *config.Config, dryRun bool) error {
    if !c.InstallPackages && len(c.BuildFiles) == 0 {
        fmt.Println("ℹ️  No packages selected and no files to build. Exiting.")
        return nil
    }

    pkgCount := 0
    var filesProcessed []string

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
            filesProcessed = append(filesProcessed, ".gitconfig")
        }

        zshProcessed := false
        if slices.Contains(c.BuildFiles, ".zshrc") {
            filesProcessed = append(filesProcessed, ".zshrc")
            zshProcessed = true
        }
        if slices.Contains(c.BuildFiles, ".zprofile") {
            filesProcessed = append(filesProcessed, ".zprofile")
            zshProcessed = true
        }

        if zshProcessed {
            buildZshConfigs(c, runtime.GOOS, runtime.GOARCH, dryRun)
        }

        if slices.Contains(c.BuildFiles, ".gitignore") {
            copyGitIgnore(dryRun)
            filesProcessed = append(filesProcessed, ".gitignore")
        }
    }

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
    destPath := filepath.Join(home, ".gitignore")

    data, err := os.ReadFile(sourcePath)
    if err != nil {
        fmt.Printf("⚠️  Warning: Could not find %s to copy\n", sourcePath)
        return
    }

    utils.WriteFiles(destPath, data, dryRun)
}

func printSummary(pkgsInstalled int, pkgsConfigured int, files []string, dryRun bool) {

    fmt.Printf("\n✨ Setup Complete!\n")
    if pkgsInstalled > 0 {
        fmt.Printf("📦 Installed:  %02d packages\n", pkgsInstalled)
    } else {
		fmt.Printf("⚙️  Configured: %02d packages\n", pkgsConfigured)
	}
    fmt.Printf("📄 Files:       %d %s\n", len(files), "generated")

    if dryRun {
        if len(files) > 0 {
            var displayNames []string
            for _, f := range files {
                displayNames = append(displayNames, "test"+f)
            }
            fmt.Printf("\n⚠️  DRY-RUN: The following files were generated in your home directory:\n   %s\n",
                strings.Join(displayNames, ", "))
        } else {
            fmt.Println("\n⚠️  DRY-RUN: No files were generated.")
        }
    }
}