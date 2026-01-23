package ui

import (
	"context"
	"runtime"
	"slices"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func RunPrompts(dryRun bool) (*config.Config, error) {
	ctx := context.Background()
	conf := &config.Config{}

	message := DisplayBanner(
		"Welcome to stash!",
		"This tool will help you install packages and configure your shell.",
	)

	tap.Intro(message)

	conf.Operation = tap.Select(ctx, tap.SelectOptions[string]{
		Message: "What would you like to do?",
		Options: []tap.SelectOption[string]{
			{Value: "configure", Label: "Configure shell", Hint: ".zshrc, .zprofile, .gitconfig, .gitignore"},
			{Value: "install", Label: "Install packages", Hint: "bun, docker, go, nvm, pipx, pnpm, etc."},
		},
	})

	if conf.Operation == "install" {
		detectedPM := utils.DetectPackageManager()
		if detectedPM == "unknown" {
			conf.PackageManager = tap.Select(ctx, tap.SelectOptions[string]{
				Message: "Select your package manager:",
				Options: []tap.SelectOption[string]{
					{Value: "apt", Label: "apt", Hint: "Debian, Ubuntu"},
					{Value: "dnf", Label: "dnf", Hint: "Fedora, RHEL, AlmaLinux"},
					{Value: "homebrew", Label: "homebrew", Hint: "macOS"},
					{Value: "macports", Label: "macports", Hint: "macOS"},
					{Value: "pacman", Label: "pacman", Hint: "Arch Linux"},
				},
			})
		} else {
			conf.PackageManager = detectedPM
		}
	}

	categories := map[string][]string{
		"CLI tools": {},
		"Exports":   {"bun", "docker", "go", "nvm", "pipx", "pnpm"},
		"Plugins":   {"fzf", "zsh-autosuggestions", "zsh-syntax-highlighting"},
	}

	if conf.Operation == "install" {
		categories["CLI tools"] = []string{"bat", "fastfetch", "fd", "ffmpeg", "gh", "git", "jq", "tree"}
	}

	if runtime.GOOS == "darwin" {
		categories["Exports"] = append(categories["Exports"], "java-android-studio")
	}

	categoryOrder := []string{"CLI tools", "Exports", "Plugins"}

	for _, cat := range categoryOrder {
		pkgs := categories[cat]
		if len(pkgs) == 0 {
			continue
		}

		slices.Sort(pkgs)

		opts := make([]tap.SelectOption[string], len(pkgs))
		for i, p := range pkgs {
			opts[i] = tap.SelectOption[string]{Value: p, Label: p}
		}

		var initial []string

		switch cat {
		case "CLI tools":
			skipDefaults := []string{"fastfetch", "ffmpeg"}

			for _, p := range pkgs {
				if !slices.Contains(skipDefaults, p) {
					initial = append(initial, p)
				}
			}
		case "Plugins":
			initial = pkgs
		}

		selected := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
			Message: "Select " + cat,
			Options: opts,
			// InitialValues: initial,
		})

		conf.SelectedPkgs = append(conf.SelectedPkgs, selected...)
	}

	if conf.Operation == "configure" {
		conf.BuildFiles = tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
			Message: "What do you want built?",
			Options: []tap.SelectOption[string]{
				{Value: ".zshrc", Label: ".zshrc"},
				{Value: ".zprofile", Label: ".zprofile"},
				{Value: ".gitconfig", Label: ".gitconfig", Hint: "Requires name and email"},
				{Value: ".gitignore", Label: ".gitignore"},
			},
		})
	}

	if slices.Contains(conf.BuildFiles, ".gitconfig") {
		conf.GitName = tap.Text(ctx, tap.TextOptions{Message: "Git Name:", DefaultValue: "John Doe", Placeholder: "John Doe"})
		conf.GitEmail = tap.Text(ctx, tap.TextOptions{Message: "Git Email:", DefaultValue: "email@example.com", Placeholder: "email@example.com"})
		conf.GitBranch = tap.Text(ctx, tap.TextOptions{Message: "Default Branch:", DefaultValue: "main", InitialValue: "main", Placeholder: "main"})
	}

	if dryRun {
		tap.Message("___ [DRY-RUN]: No changes will be written to disk.")
	}

	return conf, nil
}
