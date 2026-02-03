package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func RunPrompts(dryRun bool, version string) (*config.Config, error) {
	ctx := context.Background()

	savedConf, _ := config.Load()
	conf := &config.Config{}

	title := fmt.Sprintf("Welcome to stash! [%s]", utils.Style(version, "green"))

	if dryRun {
		title += fmt.Sprintf(" [%s]", utils.Style("DRY_RUN", "cyan"))
	}
	message := DisplayBanner(title, utils.Style("This tool will help you install packages and configure your shell.", "dim"))

	tap.Intro(message)

	step := 1
	for {
		switch step {
		case 1:
			var initialOp *string
			if savedConf.Operation != "" {
				initialOp = &savedConf.Operation
			}

			options := []tap.SelectOption[string]{
				{Value: "configure", Label: "Configure shell", Hint: ".zshrc, .zprofile, .gitconfig, .gitignore"},
				{Value: "install", Label: "Install packages", Hint: "Using your package manager"},
				{Value: "delete", Label: "Delete backup files", Hint: "~/.config/stash/bak**"},
			}

			conf.Operation = tap.Select(ctx, tap.SelectOptions[string]{
				Message:      "What would you like to do?",
				InitialValue: initialOp,
				Options:      options,
			})
			step++
		case 2:
			if conf.Operation == "install" {
				detectedPM := utils.DetectPackageManager()
				var initialPM *string
				if savedConf.PackageManager != "" {
					initialPM = &savedConf.PackageManager
				} else {
					initialPM = &detectedPM
				}
				conf.PackageManager = tap.Select(ctx, tap.SelectOptions[string]{
					Message:      "Select your package manager:",
					InitialValue: initialPM,
					Options: []tap.SelectOption[string]{
						{Value: "back", Label: "⬅ Back"},
						{Value: "apt", Label: "apt", Hint: "Debian, Ubuntu"},
						{Value: "dnf", Label: "dnf", Hint: "Fedora, RHEL, AlmaLinux"},
						{Value: "homebrew", Label: "homebrew", Hint: "macOS"},
						{Value: "macports", Label: "macports", Hint: "macOS"},
						{Value: "pacman", Label: "pacman", Hint: "Arch Linux"},
					},
				})
				if conf.PackageManager == "back" {
					step = 1
					continue
				}
				step = 4
				continue
			}

			if conf.Operation == "configure" {
				options := []tap.SelectOption[string]{
					{Value: ".zshrc", Label: ".zshrc"},
					{Value: ".zprofile", Label: ".zprofile"},
					{Value: ".gitconfig", Label: ".gitconfig", Hint: "Requires name and email"},
					{Value: ".gitignore", Label: ".gitignore"},
				}

				conf.BuildFiles = tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
					Message:       "What do you want built?",
					Options:       options,
					InitialValues: savedConf.BuildFiles,
				})

				if len(conf.BuildFiles) == 0 {
					tap.Message(utils.Style("At least one file must be selected!", "orange"))
					continue
				}
				step = 3
			}

			if conf.Operation == "delete" {
				conf.Confirm = tap.Confirm(ctx, tap.ConfirmOptions{
					Message:      "Are you sure you want to delete backup files?",
					InitialValue: false,
				})

				if !conf.Confirm {
					tap.Outro(utils.Style("🛑 [ABORTED]: No actions performed.", "orange"))
					os.Exit(0)
				}

				step = 6
			}
		case 3:
			if slices.Contains(conf.BuildFiles, ".gitconfig") {
				conf.GitName = tap.Text(ctx, tap.TextOptions{
					Message:      "Git Name:",
					Placeholder:  "John Doe",
					InitialValue: savedConf.GitName,
					Validate: func(input string) error {
						if strings.TrimSpace(input) == "" {
							return errors.New("Name is required.")
						}
						return nil
					},
				})
				conf.GitEmail = tap.Text(ctx, tap.TextOptions{
					Message:      "Git Email:",
					Placeholder:  "email@example.com",
					InitialValue: savedConf.GitEmail,
					Validate: func(input string) error {
						if strings.TrimSpace(input) == "" {
							return errors.New("Email is required.")
						}
						if !strings.Contains(input, "@") || !strings.Contains(input, ".") {
							return errors.New("Email is invalid.")
						}
						return nil
					},
				})
				conf.GitBranch = tap.Text(ctx, tap.TextOptions{
					Message:      "Default Branch:",
					DefaultValue: "main",
					Placeholder:  "main",
					InitialValue: savedConf.GitBranch,
				})
			}
			step = 4
		case 4:

			if conf.Operation == "configure" && !slices.Contains(conf.BuildFiles, ".zshrc") {
				step = 5
				continue
			}

			categories := map[string][]string{
				"CLI tools": {},
				"Exports":   {"bun", "docker", "go", "nvm", "pipx", "pnpm"},
				"Plugins":   {"fzf", "zsh-autosuggestions", "zsh-syntax-highlighting"},
			}

			if conf.Operation == "install" {
				categories["CLI tools"] = []string{"bat", "fastfetch", "fd", "ffmpeg", "gh", "git", "jq", "just", "tree"}
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

				for _, p := range pkgs {
					if conf.Operation == "configure" && slices.Contains(savedConf.SelectedPkgs, p) {
						initial = append(initial, p)
					}
				}

				selected := tap.MultiSelect(ctx, tap.MultiSelectOptions[string]{
					Message:       "Select " + cat,
					Options:       opts,
					InitialValues: initial,
				})

				conf.SelectedPkgs = append(conf.SelectedPkgs, selected...)

			}

			step = 5
			continue
		case 5:
			headers := []string{"Summary", ""}
			rows := [][]string{
				{"Operation", utils.Style(conf.Operation, "bold", "cyan")},
			}

			if conf.Operation == "install" {
				rows = append(rows, []string{"Installing with", utils.Style(conf.PackageManager, "bold", "cyan")})
			}

			if conf.Operation == "configure" {
				rows = append(rows, []string{"Build files", utils.Style(strings.Join(conf.BuildFiles, ", "), "bold", "cyan")})
			}

			if len(conf.SelectedPkgs) > 0 {
				rows = append(rows, []string{"Packages", utils.Style(strings.Join(conf.SelectedPkgs, ", "), "bold", "cyan")})
			}

			hasPackages := len(conf.SelectedPkgs) > 0
			includesZshrc := slices.Contains(conf.BuildFiles, ".zshrc")

			showSummary := (hasPackages && (conf.Operation == "install" || includesZshrc)) ||
				(conf.Operation == "configure" && !includesZshrc)

			if showSummary {
				tap.Table(headers, rows, tap.TableOptions{
					ShowBorders:   true,
					IncludePrefix: true,
					HeaderStyle:   tap.TableStyleBold,
					HeaderColor:   tap.TableColorGreen,
				})

				conf.Confirm = tap.Confirm(ctx, tap.ConfirmOptions{
					Message:      "Are you sure you want to proceed?",
					InitialValue: false,
				})

				if !conf.Confirm {
					tap.Outro(utils.Style("🛑 [ABORTED]: No actions performed.", "orange"))
					os.Exit(0)
				}
			} else {
				var msg string
				if conf.Operation == "install" || (!hasPackages && includesZshrc) {
					msg = utils.Style("No packages selected, do you want to start over?", "orange")
				} else {
					msg = utils.Style("No build files selected, do you want to start over?", "orange")
				}

				conf.StartOver = tap.Confirm(ctx, tap.ConfirmOptions{
					Message:      msg,
					InitialValue: true,
				})

				if !conf.StartOver {
					tap.Outro(utils.Style("🛑 [ABORTED]: No actions performed.", "orange"))
					os.Exit(0)
				}

				conf.SelectedPkgs = []string{}
				step = 1
				continue
			}

			step++
		case 6:
			goto end
		}
	}

end:
	if !dryRun {
		savedConf.Operation = conf.Operation

		if conf.Operation == "install" {
			savedConf.PackageManager = conf.PackageManager
		}
		if conf.Operation == "configure" {
			savedConf.BuildFiles = conf.BuildFiles

			if slices.Contains(conf.BuildFiles, ".zshrc") {
				savedConf.SelectedPkgs = conf.SelectedPkgs
			}

			if slices.Contains(conf.BuildFiles, ".gitconfig") {
				savedConf.GitName = conf.GitName
				savedConf.GitEmail = conf.GitEmail
				savedConf.GitBranch = conf.GitBranch
			}
		}
		savedConf.Save()
	}

	if dryRun {
		tap.Message(utils.Style("___ [DRY_RUN]: No changes will be made to current files on disk. ___", "orange"))
	}

	return conf, nil
}
