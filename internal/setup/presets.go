package setup

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/huffmanks/stash/internal/assets"
)

type Preset struct {
	Label string
	Files []string
}

type PresetReport struct {
	Downloaded []string
	Skipped    []string
	Failed     []string
}

var Presets = map[string]Preset{
	"eslint": {
		Label: "ESLint",
		Files: []string{"eslint.config.mjs"},
	},
	"prettier": {
		Label: "Prettier",
		Files: []string{".prettierrc", ".prettierignore"},
	},
	"biome": {
		Label: "Biome",
		Files: []string{"biome.json", ".biomeignore"},
	},
	"docker": {
		Label: "Docker",
		Files: []string{".dockerignore"},
	},
	"pnpm": {
		Label: "pnpm",
		Files: []string{".pnpmfile.cjs"},
	},
	"vscode": {
		Label: "VS Code",
		Files: []string{
			".vscode/global.settings.json",
			".vscode/workspace.settings.json",
		},
	},
}

func DownloadPreset(name string, dryRun bool) (PresetReport, error) {
	var report PresetReport

	preset, ok := Presets[name]
	if !ok {
		return report, fmt.Errorf("unknown preset: %s", name)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return report, fmt.Errorf("getting working directory: %w", err)
	}

	for _, file := range preset.Files {
		cleanPath := filepath.Clean(file)

		if filepath.IsAbs(cleanPath) || cleanPath == "." ||
			cleanPath == ".." || strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
			report.Failed = append(report.Failed, file)
			continue
		}

		destination := filepath.Join(cwd, cleanPath)

		sourcePath := filepath.ToSlash(filepath.Join(
			".dotfiles/presets",
			cleanPath,
		))

		data, err := assets.Files.ReadFile(sourcePath)
		if err != nil {
			report.Failed = append(report.Failed, file)
			continue
		}

		if _, err := os.Lstat(destination); err == nil {
			report.Skipped = append(report.Skipped, file)
			continue
		} else if !os.IsNotExist(err) {
			report.Failed = append(report.Failed, file)
			continue
		}

		if dryRun {
			report.Downloaded = append(report.Downloaded, file)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			report.Failed = append(report.Failed, file)
			continue
		}

		err = writePresetFile(destination, data)
		if err != nil {
			report.Failed = append(report.Failed, file)
			continue
		}

		report.Downloaded = append(report.Downloaded, file)
	}

	if len(report.Failed) > 0 {
		return report, fmt.Errorf("failed to download %d file(s) for %s", len(report.Failed), preset.Label)
	}

	return report, nil
}

func writePresetFile(path string, data []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, fs.FileMode(0644))
	if err != nil {
		return err
	}

	_, writeErr := file.Write(data)
	closeErr := file.Close()

	if writeErr != nil {
		os.Remove(path)
		return writeErr
	}

	return closeErr
}
