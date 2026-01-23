package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
)

func buildZshConfigs(c *config.Config, goos, arch string, dryRun bool) {
	home, _ := os.UserHomeDir()
	osFolder := map[string]string{"darwin": "macos"}[goos]
	if osFolder == "" {
		osFolder = goos
	}

	archFolder := "intel"
	if arch == "arm64" {
		archFolder = "arm"
	}

	displayOS := "Linux"
	if goos == "darwin" {
		displayOS = "macOS"
	}

	var configFiles, exportFiles, promptFiles, aliasFiles, pluginFiles []string

	categorize := func(dirPath string) {
		entries, err := assets.Files.ReadDir(dirPath)
		if err != nil {
			return
		}

		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".zsh") {
				continue
			}

			fullPath := path.Join(dirPath, entry.Name())
			base := entry.Name()

			switch {
			case strings.Contains(base, "config"):
				configFiles = append(configFiles, fullPath)
			case strings.Contains(base, "prompt"):
				promptFiles = append(promptFiles, fullPath)
			case strings.Contains(base, "aliases"):
				aliasFiles = append(aliasFiles, fullPath)
			}
		}
	}

	categorize(".dotfiles/.zsh/common")
	categorize(path.Join(".dotfiles/.zsh", osFolder))
	categorize(path.Join(".dotfiles/.zsh", osFolder, archFolder))

	slices.Sort(c.SelectedPkgs)

	searchLevels := []string{
		path.Join(".dotfiles/.zsh", osFolder, archFolder),
		path.Join(".dotfiles/.zsh", osFolder),
		".dotfiles/.zsh/common",
	}

	collectFiles := func(subDir string) []string {
		var collected []string
		for _, pkg := range c.SelectedPkgs {
			for _, level := range searchLevels {
				filePath := path.Join(level, subDir, pkg+".zsh")
				if _, err := fs.Stat(assets.Files, filePath); err == nil {
					collected = append(collected, filePath)
				}
			}
		}
		return collected
	}

	exportFiles = collectFiles("exports")
	pluginFiles = collectFiles("plugins")

	if dryRun {
		fmt.Printf("\n--- ZSH Build Manifest (%s:%s) ---\n", displayOS, arch)
	}

	var finalContent []byte
	exportsHeaderAdded := false
	pluginsHeaderAdded := false

	appendSection := func(files []string, isExport bool, isPlugin bool) {
		if len(files) == 0 {
			return
		}

		for i, f := range files {
			data, err := assets.Files.ReadFile(f)
			if err != nil {
				continue
			}
			if dryRun {
				fmt.Printf("✅ [INCLUDE]: %s\n", f)
			}

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
		utils.WriteFiles(path.Join(home, ".zshrc"), finalContent, dryRun)
	}

	if slices.Contains(c.BuildFiles, ".zprofile") {
		searchPaths := []string{
			path.Join(".dotfiles", ".zsh", osFolder, archFolder, ".zprofile"),
			path.Join(".dotfiles", ".zsh", osFolder, ".zprofile"),
		}

		for _, p := range searchPaths {
			if data, err := assets.Files.ReadFile(p); err == nil {
				if dryRun {
					fmt.Printf("📍 [FOUND]: .zprofile at: %s\n", p)
				}
				utils.WriteFiles(path.Join(home, ".zprofile"), data, dryRun)
				break
			}
		}
	}
}
