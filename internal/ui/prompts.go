package ui

import (
	"context"
	"slices"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func RunPrompts() (*config.Config, error) {
    ctx := context.Background()
	conf := &config.Config{}

    conf.InstallPackages = tap.Confirm(ctx, tap.ConfirmOptions{
        Message:      "Do you want to install packages?",
        InitialValue: true,
    })

    if conf.InstallPackages {
        detectedPM := utils.DetectPackageManager()
        if detectedPM == "unknown" {
            conf.PackageManager = tap.Select(ctx, tap.SelectOptions[string]{
                Message: "Select your package manager:",
                Options: []tap.SelectOption[string]{
                    {Value: "apt", Label: "apt"},
                    {Value: "homebrew", Label: "homebrew"},
                    {Value: "macports", Label: "macports"},
                },
            })
        } else {
            conf.PackageManager = detectedPM
        }

        categoryOrder := []string{"Essentials", "Tools", "ZSH shell"}
        categories := map[string][]string{
            "Essentials": {"bat", "fastfetch", "fd", "ffmpeg", "fzf", "gh", "git", "jq", "tree"},
            "Tools":      {"bun", "go", "nvm", "pipx", "pnpm"},
            "ZSH shell":  {"zsh-syntax-highlighting", "zsh-autosuggestions"},
        }

        for _, cat := range categoryOrder {
            pkgs := categories[cat]
            slices.Sort(pkgs)

            opts := []tap.SelectOption[string]{}
            for _, p := range pkgs {
                opts = append(opts, tap.SelectOption[string]{Value: p, Label: p})
            }

            var initial []string

            switch cat {
            case "Essentials":
                for _, p := range pkgs {
                    if p != "ffmpeg" {
                        initial = append(initial, p)
                    }
                }
            case "ZSH shell":
                initial = append(initial, pkgs...)
            }

            selected := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
                Message:       "Select " + cat,
                Options:       opts,
                InitialValues: initial,
            })

            conf.SelectedPkgs = append(conf.SelectedPkgs, selected...)
        }
    }

    conf.BuildFiles = tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
        Message: "What do you want built?",
        Options: []tap.SelectOption[string]{
            {Value: ".zshrc", Label: ".zshrc"},
            {Value: ".zprofile", Label: ".zprofile"},
            {Value: ".gitconfig", Label: ".gitconfig"},
            {Value: ".gitignore", Label: ".gitignore"},
        },
    })

	if slices.Contains(conf.BuildFiles, ".gitconfig") {
		conf.GitName = tap.Text(ctx, tap.TextOptions{Message: "Git Name:", DefaultValue: "John Doe"})
		conf.GitEmail = tap.Text(ctx, tap.TextOptions{Message: "Git Email:", DefaultValue: "email@example.com"})
		conf.GitBranch = tap.Text(ctx, tap.TextOptions{Message: "Default Branch:", DefaultValue: "main"})
	}

    return conf, nil
}