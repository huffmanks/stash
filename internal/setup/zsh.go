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

	displayOS := "Linux"
	if goos == "darwin" {
		displayOS = "macOS"
	}

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
			}
		}
	}

	categorize(".dotfiles/.zsh/common")
	categorize(filepath.Join(".dotfiles/.zsh", osFolder))
	categorize(filepath.Join(".dotfiles/.zsh", osFolder, archFolder))

	exportSearchDirs := []string{
		".dotfiles/.zsh/common/exports",
		filepath.Join(".dotfiles/.zsh", osFolder, "exports"),
		filepath.Join(".dotfiles/.zsh", osFolder, archFolder, "exports"),
	}

	for _, dir := range exportSearchDirs {
        for _, pkg := range c.SelectedPkgs {
            path := filepath.Join(dir, pkg+".zsh")
            if _, err := os.Stat(path); err == nil {
                if !slices.Contains(exportFiles, path) {
                    exportFiles = append(exportFiles, path)
                }
            }
        }
    }

	slices.Sort(exportFiles)

	pluginSearchDirs := []string{
		".dotfiles/.zsh/common/plugins",
		filepath.Join(".dotfiles/.zsh", osFolder, "plugins"),
		filepath.Join(".dotfiles/.zsh", osFolder, archFolder, "plugins"),
	}

	for _, dir := range pluginSearchDirs {
		for _, pkg := range c.SelectedPkgs {
			path := filepath.Join(dir, pkg+".zsh")
			if _, err := os.Stat(path); err == nil {
				if !slices.Contains(pluginFiles, path) {
					pluginFiles = append(pluginFiles, path)
				}
			}
		}
	}

	slices.Sort(pluginFiles)

	if dryRun {
		fmt.Printf("\n--- ZSH Build Manifest (%s:%s) ---\n", displayOS, arch)
	}

	var finalContent []byte
	exportsHeaderAdded := false
	pluginsHeaderAdded := false

	appendSection := func(files []string, isExport bool, isPlugin bool) {
        if len(files) == 0 { return }

        for i, f := range files {
            data, err := os.ReadFile(f)
            if err != nil { continue }
            if dryRun { fmt.Printf("✅ [INCLUDE]: %s\n", f) }

            if isExport && !exportsHeaderAdded {
                header := "# =====================================\n" +
                          "# Exports\n" +
                          "# =====================================\n\n"
                finalContent = append(finalContent, []byte(header)...)
                exportsHeaderAdded = true
            }

            if isPlugin && !pluginsHeaderAdded {
                header := fmt.Sprintf("# =====================================\n"+
                                     "# Plugins (%s:%s)\n"+
                                     "# =====================================\n\n", displayOS, arch)
                finalContent = append(finalContent, []byte(header)...)
                pluginsHeaderAdded = true
            }

            finalContent = append(finalContent, data...)
			if !isPlugin || i < len(files)-1 {
				finalContent = append(finalContent, '\n')
			}
        }
    }

	appendSection(configFiles, false, false)
    appendSection(exportFiles, true, false)
    appendSection(promptFiles, false, false)
    appendSection(aliasFiles, false, false)
    appendSection(pluginFiles, false, true)

	if dryRun {
		fmt.Print("\n--- End ZSH Manifest ---\n")
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
				if dryRun { fmt.Printf("📍 [FOUND]: .zprofile at: %s\n", path) }
				utils.WriteFiles(filepath.Join(home, ".zprofile"), data, dryRun)
				break
			}
		}
	}
}
