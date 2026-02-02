package setup

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"slices"
	"strings"
	"text/template"
	"time"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func ExecuteSetup(c *config.Config, dryRun bool) error {

	if c.Operation == "install" && len(c.SelectedPkgs) == 0 {
		tap.Outro(utils.Style("💡 [INFO]: No packages selected to install. Exiting.", "orange"))
		return nil
	}

	if c.Operation == "configure" && (len(c.BuildFiles) == 0 || (len(c.SelectedPkgs) == 0 && !slices.ContainsFunc(c.BuildFiles, func(f string) bool { return f != ".zshrc" }))) {
		tap.Outro(utils.Style("💡 [INFO]:  No shell files or packages selected to configure. Exiting.", "orange"))
		return nil
	}

	pkgCount := 0
	extraPkgs := 0
	needsSystemTools := false

	if c.Operation == "install" && len(c.SelectedPkgs) > 0 {

		if !dryRun {
			utils.PromptForSudo("❌ [ERROR]: sudo authentication failed.", "true", true)
		}

		if runtime.GOOS == "darwin" {
			if !utils.CommandExists("xcode-select") {
				extraPkgs++
				needsSystemTools = true
			}

			if c.PackageManager == "brew" && !utils.CommandExists("brew") {
				extraPkgs++
				needsSystemTools = true
			} else if c.PackageManager == "macports" && !utils.CommandExists("port") {
				extraPkgs++
				needsSystemTools = true
			}
		}

		if runtime.GOOS == "linux" {
			plugins := []string{"zsh-syntax-highlighting", "zsh-autosuggestions"}
			hasPlugin := false

			for _, p := range plugins {
				if slices.Contains(c.SelectedPkgs, p) {
					hasPlugin = true
					break
				}
			}

			if hasPlugin && !slices.Contains(c.SelectedPkgs, "zsh") {
				c.SelectedPkgs = append(c.SelectedPkgs, "zsh")
				extraPkgs++
			}
		}

		pkgCount = len(c.SelectedPkgs) + extraPkgs

		progress := tap.NewProgress(tap.ProgressOptions{
			Max:   pkgCount,
			Style: "heavy",
			Size:  40,
		})

		progress.Start("Installing packages...")
		time.Sleep(time.Millisecond * 100)

		var failedPkgs []string

		if needsSystemTools {
			ensureMacOSPrereqs(c.PackageManager, dryRun, progress, &failedPkgs)
		}

		if err := installSystemPkgs(c, dryRun, progress, &failedPkgs); err != nil {
			return err
		}

		time.Sleep(time.Millisecond * 100)
		progress.Stop("🏁 [FINISHED]", 0)
		time.Sleep(time.Millisecond * 100)

		successfulPkgs := c.SelectedPkgs

		if len(failedPkgs) > 0 {
			failedMap := make(map[string]bool)
			for _, p := range failedPkgs {
				failedMap[p] = true
			}

			successfulPkgs = []string{}
			for _, p := range c.SelectedPkgs {
				if !failedMap[p] {
					successfulPkgs = append(successfulPkgs, p)
				}
			}
		}

		installedPkgsMsg := fmt.Sprintf("📦 [INSTALLED]: %d packages\n\n   %s",
			len(successfulPkgs),
			strings.Join(successfulPkgs, ", "))

		if len(failedPkgs) > 0 {

			if len(successfulPkgs) > 0 {
				tap.Message(installedPkgsMsg)
			}

			failedPkgsMsg := fmt.Sprintf("❌ [FAILED]: %d packages\n\n   %s",
				len(failedPkgs),
				strings.Join(failedPkgs, ", "))

			tap.Outro(failedPkgsMsg)
		} else {
			tap.Outro(installedPkgsMsg)
		}

		time.Sleep(time.Millisecond * 100)

		os.Exit(0)
	}

	if c.Operation == "configure" && len(c.BuildFiles) > 0 {
		var created []string
		zshProcessed := false

		if slices.Contains(c.BuildFiles, ".zshrc") {
			zshProcessed = true
		}

		if slices.Contains(c.BuildFiles, ".zprofile") {
			zshProcessed = true
		}

		if zshProcessed {
			buildZshConfigs(c, runtime.GOOS, runtime.GOARCH, dryRun, &created)
		}

		if slices.Contains(c.BuildFiles, ".gitignore") {
			gitignoreSpinner := tap.NewSpinner(tap.SpinnerOptions{
				Delay: time.Millisecond * 100,
			})

			gitignoreSpinner.Start("Creating .gitignore...")
			time.Sleep(time.Millisecond * 100)

			copyGitIgnore(dryRun, &created, gitignoreSpinner)
		}

		if slices.Contains(c.BuildFiles, ".gitconfig") {
			gitconfigSpinner := tap.NewSpinner(tap.SpinnerOptions{
				Delay: time.Millisecond * 100,
			})

			gitconfigSpinner.Start("Creating .gitconfig...")
			time.Sleep(time.Millisecond * 100)

			createGitConfig(c, dryRun, &created, gitconfigSpinner)
		}

		success, missed := utils.Diff(c.BuildFiles, created)
		confMsg := fmt.Sprintf("⚙️  [CONFIGURED]: %d packages\n   🗂️  [FILES]: %d created, %d skipped",
			len(c.SelectedPkgs),
			len(success),
			len(missed),
		)
		tap.Message(confMsg)

		var sections []string
		prefix := ""
		if dryRun {
			prefix = "test_"
		}

		if len(success) > 0 {
			for i, f := range success {
				success[i] = prefix + f
			}
			msg := fmt.Sprintf("The following files were created in your home directory:\n   %s",
				utils.Style(strings.Join(success, ", "), "cyan"))
			sections = append(sections, msg)
		}

		if len(missed) > 0 {
			msg := fmt.Sprintf("The following files were skipped:\n   %s",
				utils.Style(strings.Join(missed, ", "), "orange"))
			sections = append(sections, msg)
		}

		outroMsg := strings.Join(sections, "\n\n")
		if outroMsg == "" {
			outroMsg = "✨ No files were processed."
		}
		tap.Outro(outroMsg)
		time.Sleep(time.Millisecond * 100)

		os.Exit(0)
	}

	if c.Operation == "delete" {
		spinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})

		spinner.Start("Scanning for backups...")
		time.Sleep(time.Millisecond * 100)

		report := utils.DeleteFiles(dryRun, spinner)

		spinner.Stop("Cleanup process finished", 0)
		time.Sleep(time.Millisecond * 100)

		var outroMsg string

		if len(report.Failed) > 0 {
			outroMsg += fmt.Sprintf(utils.Style("❌ [FAILED]: %d\n\n", "red"), len(report.Failed))

			for _, f := range report.Failed {
				outroMsg += fmt.Sprintf(utils.Style("     - %s\n", "cyan"), f)
			}
			outroMsg += "\n"
		}

		if len(report.Deleted) > 0 {
			header := fmt.Sprintf("🗑️  [DELETED]: %d\n\n", len(report.Deleted))
			if dryRun {
				header = fmt.Sprintf("___ [DRY_RUN]: Would delete: %d ___\n\n", len(report.Deleted))
			}
			outroMsg += fmt.Sprintf(utils.Style("%s", "orange"), header)

			for _, f := range report.Deleted {
				outroMsg += fmt.Sprintf(utils.Style("     - %s\n", "cyan"), f)
			}

		}

		if len(report.Failed) == 0 && len(report.Deleted) == 0 {
			outroMsg = "✨ [EMPTY]: No files found to delete."
		}

		tap.Outro(strings.TrimSpace(outroMsg))
		time.Sleep(time.Millisecond * 100)

		os.Exit(0)

	}

	return nil
}

const gitConfigTmpl = `[init]
    defaultBranch = {{.GitBranch}}

[user]
    name = {{.GitName}}
    email = {{.GitEmail}}

[core]
    excludesfile = ~/.gitignore

[http]
    postBuffer = 10485760
{{if .GHPath}}
[credential "https://github.com"]
    helper =
    helper = !{{.GHPath}} auth git-credential
[credential "https://gist.github.com"]
    helper =
    helper = !{{.GHPath}} auth git-credential
{{end}}`

func createGitConfig(c *config.Config, dryRun bool, created *[]string, spinner *tap.Spinner) {
	spinner.Message(("🔨 [BUILDING]: .gitconfig from template..."))
	time.Sleep(time.Millisecond * 100)

	ghPath, err := exec.LookPath("gh")
	if err == nil {
		c.GHPath = ghPath
	} else {
		c.GHPath = ""
	}

	tmpl, err := template.New("gitconfig").Parse(gitConfigTmpl)
	if err != nil {
		spinner.Stop("❌ [FAILED]: creating .gitconfig", 1)
		time.Sleep(time.Millisecond * 100)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		spinner.Stop("❌ [FAILED]: creating .gitconfig", 1)
		time.Sleep(time.Millisecond * 100)
		return
	}

	time.Sleep(time.Millisecond * 500)

	err = utils.WriteFiles(".gitconfig", buf.Bytes(), dryRun, spinner)
	if err != nil {
		spinner.Stop("❌ [FAILED]: writing .gitconfig", 1)
		time.Sleep(time.Millisecond * 100)
		return
	}

	*created = append(*created, ".gitconfig")
	spinner.Stop("✅ [CREATED]: .gitconfig", 0)
	time.Sleep(time.Millisecond * 100)
}

func copyGitIgnore(dryRun bool, created *[]string, spinner *tap.Spinner) {
	spinner.Message(("🔍 [SEARCHING]: Looking for .gitignore..."))
	time.Sleep(time.Millisecond * 100)

	sourcePath := ".dotfiles/git/.gitignore"

	data, err := assets.Files.ReadFile(sourcePath)
	if err != nil {
		spinner.Stop(fmt.Sprintf("⚠️ [SKIPPED]: No .gitignore found at: %s", sourcePath), 1)
		time.Sleep(time.Millisecond * 100)
		return
	}

	spinner.Message(fmt.Sprintf("📍 [FOUND]: .gitignore at: %s", sourcePath))
	time.Sleep(time.Millisecond * 500)

	err = utils.WriteFiles(".gitignore", data, dryRun, spinner)
	if err != nil {
		spinner.Stop("❌ [FAILED]: writing .gitignore", 1)
		time.Sleep(time.Millisecond * 100)
		return
	}

	*created = append(*created, ".gitignore")
	spinner.Stop("✅ [CREATED]: .gitignore", 0)
	time.Sleep(time.Millisecond * 100)
}
