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

		if !dryRun && !utils.HasSudoPrivilege() {
			tap.Message("Root privileges are required for installing packages.")

			cmd := exec.Command("sh", "-c", "sudo", "-v")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin

			if err := cmd.Run(); err != nil {
				tap.Outro("❌ [ERROR]: Sudo authentication failed. Exiting.")
				os.Exit(1)
			}
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
			hasPlugins := slices.Contains(c.SelectedPkgs, "zsh-syntax-highlighting") ||
				slices.Contains(c.SelectedPkgs, "zsh-autosuggestions")
			if hasPlugins && !slices.Contains(c.SelectedPkgs, "zsh") {
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

		var failedPkgs []string

		if needsSystemTools {
			ensureMacOSPrereqs(c.PackageManager, dryRun, progress, &failedPkgs)
		}

		if err := installSystemPkgs(c, dryRun, progress, &failedPkgs); err != nil {
			return err
		}

		time.Sleep(time.Second * 1)
		progress.Stop("🏁 [FINISHED]", 0)
		time.Sleep(time.Second * 1)

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
	}

	if c.Operation == "configure" && len(c.BuildFiles) > 0 {

		gitignoreSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})
		gitconfigSpinner := tap.NewSpinner(tap.SpinnerOptions{
			Delay: time.Millisecond * 100,
		})

		zshProcessed := false

		if slices.Contains(c.BuildFiles, ".zshrc") {
			zshProcessed = true
		}

		if slices.Contains(c.BuildFiles, ".zprofile") {
			zshProcessed = true
		}

		if zshProcessed {
			buildZshConfigs(c, runtime.GOOS, runtime.GOARCH, dryRun)
		}

		if slices.Contains(c.BuildFiles, ".gitignore") {
			gitignoreSpinner.Start("Creating .gitignore...")
			time.Sleep(time.Second * 2)

			copyGitIgnore(dryRun, gitignoreSpinner)
		}

		if slices.Contains(c.BuildFiles, ".gitconfig") {
			gitconfigSpinner.Start("Creating .gitconfig...")
			time.Sleep(time.Second * 2)

			createGitConfig(c, dryRun, gitconfigSpinner)
		}

		confMsg := fmt.Sprintf("⚙️  [CONFIGURED]: %d packages\n   🗂️  [FILES]: %d created",
			len(c.SelectedPkgs),
			len(c.BuildFiles),
		)
		tap.Message(confMsg)

		var displayNames []string
		prefix := ""
		if dryRun {
			prefix = "test"
		}

		for _, f := range c.BuildFiles {
			displayNames = append(displayNames, prefix+f)
		}
		outroMsg := fmt.Sprintf("The following files were created in your home directory:\n   %s",
			utils.Style(strings.Join(displayNames, ", "), "cyan"))
		tap.Outro(outroMsg)

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
`

func createGitConfig(c *config.Config, dryRun bool, spinner *tap.Spinner) {
	spinner.Message(("🔨 [BUILDING]: .gitconfig from template..."))

	tmpl, err := template.New("gitconfig").Parse(gitConfigTmpl)
	if err != nil {
		spinner.Stop("❌ [FAILED]: creating .gitconfig", 1)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, c); err != nil {
		spinner.Stop("❌ [FAILED]: creating .gitconfig", 1)
		return
	}

	time.Sleep(time.Second * 1)
	utils.WriteFiles(".gitconfig", buf.Bytes(), dryRun, spinner)
	spinner.Stop("✅ [CREATED]: .gitconfig", 0)
}

func copyGitIgnore(dryRun bool, spinner *tap.Spinner) {
	spinner.Message(("🔍 [SEARCHING]: Looking for .gitignore..."))
	sourcePath := ".dotfiles/git/.gitignore"

	data, err := assets.Files.ReadFile(sourcePath)
	if err != nil {
		spinner.Stop(fmt.Sprintf("⚠️ [SKIPPED]: No .gitignore found at: %s", sourcePath), 1)
		return
	}

	spinner.Message(fmt.Sprintf("📍 [FOUND]: .gitignore at: %s", sourcePath))
	time.Sleep(time.Second * 1)

	utils.WriteFiles(".gitignore", data, dryRun, spinner)

	spinner.Stop("✅ [CREATED]: .gitignore", 0)
}
