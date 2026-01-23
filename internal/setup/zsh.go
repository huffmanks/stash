package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
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

	if slices.Contains(c.BuildFiles, ".zshrc") {
		zshrcSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})

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

		zshrcBuildMsg := fmt.Sprintf("--- ZSH Build Manifest (%s:%s) ---", displayOS, arch)
		zshrcSpinner.Start(zshrcBuildMsg)
		time.Sleep(time.Second * 2)

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
				includeMsg := fmt.Sprintf("✅ [INCLUDE]: %s", f)
				zshrcSpinner.Message(includeMsg)

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

		zshrcSpinner.Message("--- End ZSH Manifest ---")

		utils.WriteFiles(path.Join(home, ".zshrc"), finalContent, dryRun, zshrcSpinner)

		zshrcSpinner.Stop("✅ [CREATED]: .zshrc", 0)
	}

	if slices.Contains(c.BuildFiles, ".zprofile") {
		zprofileSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})

		searchPaths := []string{
			path.Join(".dotfiles", ".zsh", osFolder, archFolder, ".zprofile"),
			path.Join(".dotfiles", ".zsh", osFolder, ".zprofile"),
		}

		for _, p := range searchPaths {
			if data, err := assets.Files.ReadFile(p); err == nil {
				zprofileMsg := fmt.Sprintf("📍 [FOUND]: .zprofile at: %s", p)
				zprofileSpinner.Start(zprofileMsg)
				time.Sleep(time.Second * 2)

				utils.WriteFiles(path.Join(home, ".zprofile"), data, dryRun, zprofileSpinner)
				zprofileSpinner.Stop("✅ [CREATED]: .zprofile", 0)
				break
			}
		}
	}

}
