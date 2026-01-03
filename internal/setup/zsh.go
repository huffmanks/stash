package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
)

func buildZshConfigs(c *config.Config, goos, arch string, dryRun bool) {
	home, _ := os.UserHomeDir()
	osFolder := map[string]string{"darwin": "macos"}[goos]
	if osFolder == "" { osFolder = goos }

	archFolder := "intel"
	if arch == "arm64" { archFolder = "arm" }

	var configFiles, exportFiles, promptFiles, aliasFiles, pluginFiles []string

	categorize := func(path string) {
		files, _ := filepath.Glob(filepath.Join(path, "*.zsh"))
		slices.Sort(files)
		for _, f := range files {
			base := filepath.Base(f)
			switch {
			case strings.Contains(base, "config"):
				configFiles = append(configFiles, f)
			case strings.Contains(base, "prompt"):
				promptFiles = append(promptFiles, f)
			case strings.Contains(base, "aliases"):
				aliasFiles = append(aliasFiles, f)
			case strings.Contains(base, "plugins"):
				pluginFiles = append(pluginFiles, f)
			}
		}
	}

	categorize(".dotfiles/.zsh/common")
	categorize(filepath.Join(".dotfiles/.zsh", osFolder))
	categorize(filepath.Join(".dotfiles/.zsh", osFolder, archFolder))

	exportSearchDirs := []string{
		".dotfiles/.zsh/common/exports",
		filepath.Join(".dotfiles/.zsh", osFolder, "exports"),
	}
	for _, dir := range exportSearchDirs {
		for _, pkg := range c.SelectedPkgs {
			pkgName := strings.TrimSuffix(pkg, "lang")
			path := filepath.Join(dir, pkgName+".zsh")
			if _, err := os.Stat(path); err == nil {
				exportFiles = append(exportFiles, path)
			}
		}
	}

	var manifest []string
	manifest = append(manifest, configFiles...)
	manifest = append(manifest, exportFiles...)
	manifest = append(manifest, promptFiles...)
	manifest = append(manifest, aliasFiles...)
	manifest = append(manifest, pluginFiles...)

	if dryRun {
		fmt.Printf("\n--- ZSH Build Manifest (%s/%s) ---\n", osFolder, arch)
	}

	var finalContent []byte
	exportsHeaderAdded := false

	for _, f := range manifest {
		data, err := os.ReadFile(f)
		if err != nil { continue }
		if dryRun { fmt.Printf("✅ INCLUDE: %s\n", f) }

		if !exportsHeaderAdded && slices.Contains(exportFiles, f) {
            header := "# =====================================\n" +
                      "# Exports\n" +
                      "# =====================================\n\n"
            finalContent = append(finalContent, []byte(header)...)
            exportsHeaderAdded = true
        }

		finalContent = append(finalContent, data...)
	}

	if slices.Contains(c.BuildFiles, ".zshrc") {
		utils.WriteFiles(filepath.Join(home, ".zshrc"), finalContent, dryRun)
	}

	if slices.Contains(c.BuildFiles, ".zprofile") {
		searchPaths := []string{
			filepath.Join(".dotfiles", ".zsh", osFolder, archFolder, ".zprofile"),
			filepath.Join(".dotfiles", ".zsh", osFolder, ".zprofile"),
		}

		for _, path := range searchPaths {
			if data, err := os.ReadFile(path); err == nil {
				if dryRun { fmt.Printf("🔍 Found .zprofile at: %s\n", path) }
				utils.WriteFiles(filepath.Join(home, ".zprofile"), data, dryRun)
				break
			}
		}
	}

	if dryRun { fmt.Println("--- End ZSH Manifest ---") }
}