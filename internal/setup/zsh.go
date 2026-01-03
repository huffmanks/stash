package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/huffmanks/config-stash/internal/config"
	"github.com/huffmanks/config-stash/internal/utils"
)

func buildZshConfigs(c *config.Config, goos, arch string, dryRun bool) {
	home, _ := os.UserHomeDir()
	osFolder := map[string]string{"darwin": "macos"}[goos]
	if osFolder == "" { osFolder = goos }

	manifest := []string{}

	addFromDir := func(path string) {
		files, _ := filepath.Glob(filepath.Join(path, "*.zsh"))
		slices.Sort(files)
		manifest = append(manifest, files...)
	}

	addFromDir(".dotfiles/.zsh/common")

	exportDirs := []string{
		".dotfiles/.zsh/common/exports",
		fmt.Sprintf(".dotfiles/.zsh/%s/exports", osFolder),
	}
	for _, dir := range exportDirs {
		for _, pkg := range c.SelectedPkgs {
			pkgName := strings.TrimSuffix(pkg, "lang")
			exportFile := filepath.Join(dir, pkgName+".zsh")
			if _, err := os.Stat(exportFile); err == nil {
				manifest = append(manifest, exportFile)
			}
		}
	}

	addFromDir(filepath.Join(".dotfiles/.zsh", osFolder))

	archPath := filepath.Join(".dotfiles/.zsh", osFolder, "intel")
	if arch == "arm64" {
		archPath = filepath.Join(".dotfiles/.zsh", osFolder, "arm")
	}
	addFromDir(archPath)

	if dryRun {
		fmt.Printf("\n--- ZSH Build Manifest (%s/%s) ---\n", osFolder, arch)
	}

	var finalContent []byte

	if c.PackageManager == "macports" {
		pathGuard := []byte(`# MacPorts Path Setup
		export PATH="/opt/local/bin:/opt/local/sbin:$PATH"
	`)
		finalContent = append(finalContent, pathGuard...)
	}

	for _, f := range manifest {
		data, err := os.ReadFile(f)
		if err != nil { continue }

		if dryRun { fmt.Printf("✅ INCLUDE: %s\n", f) }

		header := fmt.Sprintf("\n# --- Source: %s ---\n", f)
		finalContent = append(finalContent, []byte(header)...)
		finalContent = append(finalContent, data...)
		finalContent = append(finalContent, '\n')
	}

	if slices.Contains(c.BuildFiles, ".zshrc") {
		utils.WriteFilesDry(filepath.Join(home, ".zshrc"), finalContent, dryRun)
	}

	if slices.Contains(c.BuildFiles, ".zprofile") {
		zprof := []byte("# Source zshrc\n[[ -f ~/.zshrc ]] && . ~/.zshrc\n")
		utils.WriteFilesDry(filepath.Join(home, ".zprofile"), zprof, dryRun)
	}

	if dryRun { fmt.Println("--- End ZSH Manifest ---") }
}