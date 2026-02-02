package setup

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func buildZshConfigs(c *config.Config, goos, arch string, dryRun bool, created *[]string) {

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

	if utils.IsAndroid() {
		archFolder = "android"
		displayOS = "Android"
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

		zshrcSpinner.Start("Builidng .zshrc...")
		time.Sleep(time.Millisecond * 100)

		var finalBuffer bytes.Buffer
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
				zshrcSpinner.Message(fmt.Sprintf("✅ [INCLUDE]: %s", f))
				time.Sleep(time.Millisecond * 100)

				if isExport && !exportsHeaderAdded {
					fmt.Fprint(&finalBuffer, "# =====================================\n# Exports\n# =====================================\n\n")
					exportsHeaderAdded = true
				}

				if isPlugin && !pluginsHeaderAdded {
					fmt.Fprintf(&finalBuffer, "# =====================================\n# Plugins (%s:%s)\n# =====================================\n\n", displayOS, arch)
					pluginsHeaderAdded = true
				}

				finalBuffer.Write(data)
				if !isPlugin || i < len(files)-1 {
					finalBuffer.WriteByte('\n')
				}
			}
		}

		appendSection(configFiles, false, false)
		appendSection(exportFiles, true, false)
		appendSection(promptFiles, false, false)
		appendSection(aliasFiles, false, false)
		appendSection(pluginFiles, false, true)

		zshrcSpinner.Message("--- End ZSH Manifest ---")
		time.Sleep(time.Millisecond * 100)

		err := utils.WriteFiles(".zshrc", finalBuffer.Bytes(), dryRun, zshrcSpinner)
		if err != nil {
			zshrcSpinner.Stop("❌ [FAILED]: writing .zshrc", 1)
			time.Sleep(time.Millisecond * 100)
			return
		}

		*created = append(*created, ".zshrc")
		zshrcSpinner.Stop("✅ [CREATED]: .zshrc", 0)
		time.Sleep(time.Millisecond * 100)
	}

	if slices.Contains(c.BuildFiles, ".zprofile") {
		zprofileSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})

		searchPaths := []string{
			path.Join(".dotfiles", ".zsh", osFolder, archFolder, ".zprofile"),
			path.Join(".dotfiles", ".zsh", osFolder, ".zprofile"),
		}

		zprofileSpinner.Start("Searching for .zprofile...")
		time.Sleep(time.Millisecond * 100)

		var foundData []byte
		var foundPath string

		for _, p := range searchPaths {
			if data, err := assets.Files.ReadFile(p); err == nil {
				foundData = data
				foundPath = p
				break
			}
		}

		if foundData != nil {
			zprofileSpinner.Message(fmt.Sprintf("📍 [FOUND]: .zprofile at: %s", foundPath))
			time.Sleep(time.Millisecond * 100)

			err := utils.WriteFiles(".zprofile", foundData, dryRun, zprofileSpinner)
			if err != nil {
				zprofileSpinner.Stop("❌ [FAILED]: writing .zprofile", 1)
				time.Sleep(time.Millisecond * 100)
				return
			}

			*created = append(*created, ".zprofile")
			zprofileSpinner.Stop("✅ [CREATED]: .zprofile", 0)
			time.Sleep(time.Millisecond * 100)
		} else {
			zprofileSpinner.Stop("⚠️ [SKIPPED]: No .zprofile found in search paths", 1)
			time.Sleep(time.Millisecond * 100)
		}
	}

}
