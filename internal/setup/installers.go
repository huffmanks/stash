package setup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/assets"
	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
	"github.com/yarlson/tap"
)

func installSystemPkgs(c *config.Config, dryRun bool, progress *tap.Progress, failedPkgs *[]string) error {

	for _, pkg := range c.SelectedPkgs {
		var err error

		if runtime.GOOS != "linux" && pkg != "docker" {
			msg := fmt.Sprintf("📦 Installing %s...", pkg)
			progress.Message(msg)

			if dryRun {
				time.Sleep(time.Second * 2)
			}
		}

		switch pkg {
		case "bun":
			err = utils.RunCmd("curl -fsSL https://bun.com/install | bash", dryRun, progress)
		case "docker":
			if runtime.GOOS == "linux" {
				err = installDocker(dryRun, progress)
			}
		case "go":
			err = installGo(dryRun, progress)
		case "java-android-studio":
			err = installZulu(c.PackageManager, dryRun, progress)
		case "nvm":
			err = utils.RunCmd("curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash", dryRun, progress)
		case "pnpm":
			err = utils.RunCmd("curl -fsSL https://get.pnpm.io/install.sh | sh -", dryRun, progress)
		default:
			err = installViaPM(c.PackageManager, pkg, dryRun, progress)
		}

		if err != nil {
			if !slices.Contains(*failedPkgs, pkg) {
				*failedPkgs = append(*failedPkgs, pkg)
			}

			progress.Advance(1, fmt.Sprintf("❌ [%s]: failed", pkg))
		} else {
			progress.Advance(1, fmt.Sprintf("✅ [%s]: installed", pkg))
		}

		time.Sleep(time.Second * 1)

		if pkg == "zsh" && runtime.GOOS == "linux" {
			if utils.HasSudoPrivilege() {
				utils.RunCmd("sudo chsh -s $(which zsh) $(whoami)", dryRun, progress)
			}
		}
		if pkg == "bat" && runtime.GOOS == "linux" {
			if utils.HasSudoPrivilege() {
				aliasCmd := `if command -v batcat &>/dev/null && ! command -v bat &>/dev/null; then sudo update-alternatives --install /usr/local/bin/bat bat /usr/bin/batcat 1; fi`
				utils.RunCmd(aliasCmd, dryRun, progress)
			}
		}
	}
	return nil
}

func installViaPM(pm, pkg string, dryRun bool, progress *tap.Progress) error {
	var cmdStr string
	switch pm {
	case "apt":
		cmdStr = fmt.Sprintf("sudo apt install -y %s", pkg)
	case "dnf":
		cmdStr = fmt.Sprintf("sudo dnf install -y %s", pkg)
	case "homebrew":
		cmdStr = fmt.Sprintf("brew install %s", pkg)
	case "macports":
		cmdStr = fmt.Sprintf("sudo port install %s", pkg)
	case "pacman":
		cmdStr = fmt.Sprintf("sudo pacman -S --noconfirm %s", pkg)
	}

	if cmdStr == "" {
		return fmt.Errorf("Unsupported package manager.")
	}

	return utils.RunCmd(cmdStr, dryRun, progress)
}

func installDocker(dryRun bool, progress *tap.Progress) error {
	tempScript := path.Join(os.TempDir(), "get-docker.sh")

	if !dryRun {
		data, err := assets.Files.ReadFile("scripts/get-docker.sh")
		if err != nil {
			msg := fmt.Sprintf("⛔️ [ERROR]: Failed to read docker script: %v", err)
			progress.Message(msg)

			return fmt.Errorf("Failed to read docker script: %w", err)
		}

		err = os.WriteFile(tempScript, data, 0755)
		if err != nil {
			msg := fmt.Sprintf("⛔️ [ERROR]: Failed to write temp script: %v", err)
			progress.Message(msg)

			return fmt.Errorf("Failed to write temp script: %w", err)
		}

		defer os.Remove(tempScript)
	}

	return utils.RunCmd(fmt.Sprintf("sudo sh %s", tempScript), dryRun, progress)
}

func installGo(dryRun bool, progress *tap.Progress) error {
	version := "1.25.5"
	if !dryRun {
		out, err := exec.Command("sh", "-c", "curl -s 'https://go.dev/VERSION?m=text' | head -n 1").Output()
		if err == nil {
			version = strings.TrimPrefix(strings.TrimSpace(string(out)), "go")
		}
	}

	var cmd string
	if runtime.GOOS == "darwin" {
		url := fmt.Sprintf("https://go.dev/dl/go%s.darwin-%s.pkg", version, runtime.GOARCH)
		cmd = fmt.Sprintf("curl -LO %s && sudo installer -pkg go%s.darwin-%s.pkg -target /", url, version, runtime.GOARCH)
	} else {
		url := fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", version, runtime.GOARCH)
		cmd = fmt.Sprintf("curl -L %s | sudo tar -C /usr/local -xzf -", url)
	}

	return utils.RunCmd(cmd, dryRun, progress)
}

func installZulu(pm string, dryRun bool, progress *tap.Progress) error {
	var cmdStr string
	switch pm {
	case "homebrew":
		cmdStr = "brew install --cask zulu@17"
	case "macports":
		cmdStr = "sudo port install openjdk17-zulu"
	}

	if cmdStr == "" {
		return fmt.Errorf("Unsupported package manager for java: %s", pm)
	}

	return utils.RunCmd(cmdStr, dryRun, progress)
}

func ensureMacOSPrereqs(pm string, dryRun bool, progress *tap.Progress, failedPkgs *[]string) {
	_, err := exec.LookPath("xcode-select")
	if err != nil {
		if dryRun {
			progress.Advance(1, "___ [DRY-RUN]: Would ensure xcode-select is installed")
		} else {
			cmdErr := utils.RunCmd("xcode-select --install", dryRun, progress)
			if cmdErr != nil {
				*failedPkgs = append(*failedPkgs, "xcode")
			}
			progress.Advance(1, "📦 [INSTALLING]: Xcode Command Line Tools...")
		}
	}

	switch pm {
	case "homebrew":
		if _, err := exec.LookPath("brew"); err != nil {
			cmdErr := utils.RunCmd(`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`, dryRun, progress)

			if cmdErr != nil {
				*failedPkgs = append(*failedPkgs, "homebrew")
			}
			progress.Advance(1, "📦 [INSTALLING]: Homebrew...")
		}
	case "macports":
		if _, err := exec.LookPath("port"); err != nil {
			cmdErr := installMacPorts(dryRun, progress)

			if cmdErr != nil {
				*failedPkgs = append(*failedPkgs, "macports")
			}
			progress.Advance(1, "📦 [INSTALLING]: Macports...")
		}
	}
}

func installMacPorts(dryRun bool, progress *tap.Progress) error {
	out, _ := exec.Command("sw_vers", "-productVersion").Output()
	versionStr := strings.TrimSpace(string(out))

	var osName string
	switch {
	case strings.HasPrefix(versionStr, "26"):
		osName = "26-Tahoe"
	case strings.HasPrefix(versionStr, "15"):
		osName = "15-Sequoia"
	case strings.HasPrefix(versionStr, "14"):
		osName = "14-Sonoma"
	case strings.HasPrefix(versionStr, "13"):
		osName = "13-Ventura"
	case strings.HasPrefix(versionStr, "12"):
		osName = "12-Monterey"
	case strings.HasPrefix(versionStr, "11"):
		osName = "11-BigSur"
	default:
		msg := fmt.Sprintf("⚠️ [WARNING]: macOS %s not in auto-install list.", versionStr)
		progress.Message(msg)
		return fmt.Errorf("macOS %s not in auto-install list.", versionStr)
	}

	pkgName := "MacPorts-Latest.pkg"
	downloadURL := ""

	if !dryRun {
		resp, err := http.Get("https://api.github.com/repos/macports/macports-base/releases/latest")
		if err == nil {
			defer resp.Body.Close()
			var release config.MacPortRelease

			if err := json.NewDecoder(resp.Body).Decode(&release); err == nil {
				for _, asset := range release.Assets {
					if strings.Contains(asset.Name, osName) && strings.HasSuffix(asset.Name, ".pkg") {
						pkgName = asset.Name
						downloadURL = asset.BrowserDownloadURL
						break
					}
				}
			}
		}
	}

	if downloadURL == "" {
		downloadURL = fmt.Sprintf("https://distfiles.macports.org/MacPorts/%s", pkgName)
	}

	if dryRun {
		msg := fmt.Sprintf("___ [DRY-RUN]: %s. Would download: %s", versionStr, downloadURL)
		progress.Message(msg)
		return nil
	}

	dlMsg := fmt.Sprintf("↓ [DOWNLOADING]: MacPorts %s for %s...", pkgName, osName)
	progress.Message(dlMsg)
	cmdErrDownload := utils.RunCmd(fmt.Sprintf("curl -O %s", downloadURL), false, progress)
	if cmdErrDownload != nil {
		return cmdErrDownload
	}

	cmdErrInstall := utils.RunCmd(fmt.Sprintf("sudo installer -pkg %s -target /", pkgName), false, progress)
	_ = os.Remove(pkgName)

	if cmdErrInstall != nil {
		return cmdErrInstall
	}

	return nil
}
