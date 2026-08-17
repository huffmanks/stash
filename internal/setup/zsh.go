package setup

import (
	"bytes"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

var dynamicPkgs = map[string]config.DynamicPackage{
	"fzf": {
		CandidateGroups: [][]string{
			{
				"~/.zsh/fzf/completion.zsh",
				"/opt/homebrew/opt/fzf/shell/completion.zsh",
				"/usr/local/opt/fzf/shell/completion.zsh",
				"/home/linuxbrew/.linuxbrew/opt/fzf/shell/completion.zsh",
				"/opt/local/share/fzf/shell/completion.zsh",
				"/usr/share/fzf/completion.zsh",
				"/usr/share/fzf/shell/completion.zsh",
				"/usr/share/doc/fzf/examples/completion.zsh",
			},
			{
				"~/.zsh/fzf/key-bindings.zsh",
				"/opt/homebrew/opt/fzf/shell/key-bindings.zsh",
				"/usr/local/opt/fzf/shell/key-bindings.zsh",
				"/home/linuxbrew/.linuxbrew/opt/fzf/shell/key-bindings.zsh",
				"/opt/local/share/fzf/shell/key-bindings.zsh",
				"/usr/share/fzf/key-bindings.zsh",
				"/usr/share/fzf/shell/key-bindings.zsh",
				"/usr/share/doc/fzf/examples/key-bindings.zsh",
			},
		},
		GetPrefix: func(goos, pm string, hasResolved bool) []string {
			if !hasResolved {
				return []string{"source <(fzf --zsh)"}
			}
			return nil
		},
	},
	"zsh-autosuggestions": {
		CandidateGroups: [][]string{
			{
				"~/.zsh/zsh-autosuggestions/zsh-autosuggestions.zsh",
				"/opt/homebrew/share/zsh-autosuggestions/zsh-autosuggestions.zsh",
				"/usr/local/share/zsh-autosuggestions/zsh-autosuggestions.zsh",
				"/usr/share/zsh-autosuggestions/zsh-autosuggestions.zsh",
				"/usr/share/zsh/plugins/zsh-autosuggestions/zsh-autosuggestions.zsh",
			},
		},
	},
	"zsh-syntax-highlighting": {
		CandidateGroups: [][]string{
			{
				"~/.zsh/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh",
				"/opt/homebrew/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh",
				"/usr/local/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh",
				"/usr/share/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh",
				"/usr/share/zsh/plugins/zsh-syntax-highlighting/zsh-syntax-highlighting.zsh",
			},
		},
	},
}

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

	pm := utils.DetectPackageManager()

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

				fullPath := filepath.Join(dirPath, entry.Name())
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
					filePath := filepath.Join(level, subDir, pkg+".zsh")
					if _, err := fs.Stat(assets.Files, filePath); err == nil {
						collected = append(collected, filePath)
					}
				}
			}
			return collected
		}

		exportFiles = collectFiles("exports")
		pluginFiles = collectFiles("plugins")

		staticPlugins := make(map[string]string, len(pluginFiles))
		for _, f := range pluginFiles {
			pkg := strings.TrimSuffix(path.Base(f), ".zsh")
			staticPlugins[pkg] = f
		}

		zshrcSpinner.Start("Builidng .zshrc...")
		time.Sleep(time.Millisecond * 100)

		var finalBuffer bytes.Buffer
		writtenHeaders := make(map[string]bool)

		writeHeader := func(title string) {
			if title != "" && !writtenHeaders[title] {
				fmt.Fprintf(&finalBuffer, "# =====================================\n# %s\n# =====================================\n\n", title)
				writtenHeaders[title] = true
			}
		}

		appendSection := func(title string, files []string) {
			if len(files) == 0 {
				return
			}

			for _, f := range files {
				data, err := assets.Files.ReadFile(f)
				if err != nil {
					continue
				}
				zshrcSpinner.Message(fmt.Sprintf("✅ [INCLUDE]: %s", f))
				time.Sleep(time.Millisecond * 100)

				writeHeader(title)

				finalBuffer.Write(data)
				finalBuffer.WriteByte('\n')
			}
		}

		appendSection("", configFiles)
		appendSection(fmt.Sprintf("Exports (%s:%s)", displayOS, arch), exportFiles)
		appendSection("", promptFiles)
		appendSection("", aliasFiles)

		pluginSectionTitle := fmt.Sprintf("Plugins (%s:%s)", displayOS, arch)

		for _, pkg := range c.SelectedPkgs {
			hasDynamic := false
			if dynPkg, exists := dynamicPkgs[pkg]; exists {
				if snippet := utils.GenerateSourceSnippet(pkg, dynPkg, goos, pm); snippet != "" {
					writeHeader(pluginSectionTitle)
					finalBuffer.WriteString(snippet)
					hasDynamic = true
				}
			}

			hasStatic := false
			if filePath, exists := staticPlugins[pkg]; exists {
				if data, err := assets.Files.ReadFile(filePath); err == nil {
					writeHeader(pluginSectionTitle)
					zshrcSpinner.Message(fmt.Sprintf("✅ [INCLUDE]: %s", filePath))
					time.Sleep(time.Millisecond * 100)

					if hasDynamic && !bytes.HasPrefix(data, []byte("\n")) {
						finalBuffer.WriteByte('\n')
					}

					finalBuffer.Write(data)
					if !bytes.HasSuffix(data, []byte("\n")) {
						finalBuffer.WriteByte('\n')
					}
					hasStatic = true
				}
			}

			if hasDynamic || hasStatic {
				finalBuffer.WriteByte('\n')
			}
		}

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
