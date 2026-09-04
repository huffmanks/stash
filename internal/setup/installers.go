package setup

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
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
	if c.PackageManager == "macports" && !dryRun {
		utils.PromptForSudo("❌ [ERROR]: sudo authentication failed.", true)

		if err := utils.RunCmd("sudo port sync", dryRun, progress); err != nil {
			progress.Message(fmt.Sprintf("❌ [ERROR]: MacPorts sync failed: %v", err))
			time.Sleep(time.Millisecond * 500)
		}
	}

	for _, pkg := range c.SelectedPkgs {
		var err error

		if runtime.GOOS != "linux" && pkg != "docker" {
			msg := fmt.Sprintf("📦 Installing %s...", pkg)
			progress.Message(msg)
			time.Sleep(time.Millisecond * 100)

			if dryRun {
				time.Sleep(time.Millisecond * 500)
			}
		}

		isZshPlugin := strings.HasPrefix(pkg, "zsh-") && runtime.GOOS == "linux"

		switch {
		case pkg == "bat":
			err = installViaPM(c.PackageManager, pkg, dryRun, progress)
			if err == nil && runtime.GOOS == "linux" {
				aliasCmd := `if command -v batcat &>/dev/null && ! command -v bat &>/dev/null; then sudo update-alternatives --install /usr/local/bin/bat bat /usr/bin/batcat 1; fi`
				utils.RunCmd(aliasCmd, dryRun, progress)
			}
		case pkg == "bun":
			err = utils.RunCmd("curl -fsSL https://bun.com/install | bash", dryRun, progress)
		case pkg == "docker":
			if runtime.GOOS == "linux" {
				err = installDocker(dryRun, progress)
			} else {
				progress.Message(fmt.Sprintln("⚠️ [docker]: Skipping installation (only automated for Linux)"))
				progress.Advance(1, "⚠️ [docker]: skipped")
				continue
			}
		case pkg == "go":
			err = installGo(dryRun, progress)
		case pkg == "nvm":
			err = utils.RunCmd("curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash", dryRun, progress)
		case pkg == "pnpm":
			isMacIntel := runtime.GOOS == "darwin" && runtime.GOARCH == "amd64"

			if isMacIntel {
				if _, lookErr := exec.LookPath("npm"); lookErr == nil {
					err = utils.RunCmd("npm i -g pnpm", dryRun, progress)
				} else {
					progress.Message("⚠️ [pnpm]: npm not found. Skipping pnpm installation...")
					progress.Advance(1, "⚠️ [pnpm]: skipped (npm missing)")
					continue
				}
			} else {
				err = utils.RunCmd("curl -fsSL https://get.pnpm.io/install.sh | sh -", dryRun, progress)
			}
		case pkg == "uv":
			err = utils.RunCmd("curl -LsSf https://astral.sh/uv/install.sh | sh", dryRun, progress)
		case pkg == "zsh":
			err = installViaPM(c.PackageManager, pkg, dryRun, progress)
			if err == nil && runtime.GOOS == "linux" {
				utils.RunCmd("sudo chsh -s $(which zsh) $(whoami)", dryRun, progress)
			}
		case isZshPlugin:
			repo := fmt.Sprintf("https://github.com/zsh-users/%s", pkg)
			home, _ := os.UserHomeDir()
			target := filepath.Join(home, ".zsh", pkg)
			err = gitClone(repo, target, dryRun, progress)
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

		time.Sleep(time.Millisecond * 500)

	}
	return nil
}

func installViaPM(pm, pkg string, dryRun bool, progress *tap.Progress) error {
	resolvedPkg := utils.ResolvePkgName(pm, pkg)
	var cmdStr string

	switch pm {
	case "apk":
		cmdStr = fmt.Sprintf("sudo apk add %s", resolvedPkg)
	case "apt":
		cmdStr = fmt.Sprintf("sudo DEBIAN_FRONTEND=noninteractive apt install -y %s", resolvedPkg)
	case "dnf":
		cmdStr = fmt.Sprintf("sudo dnf install -y %s", resolvedPkg)
	case "brew", "homebrew":
		cmdStr = fmt.Sprintf("NONINTERACTIVE=1 brew install %s", resolvedPkg)
	case "macports":
		cmdStr = fmt.Sprintf("sudo port -N install %s", resolvedPkg)
	case "pacman":
		cmdStr = fmt.Sprintf("sudo pacman -S --noconfirm %s", resolvedPkg)
	}

	if cmdStr == "" {
		return fmt.Errorf("⚠️ [WARNING]: Unsupported package manager.")
	}

	return utils.RunCmd(cmdStr, dryRun, progress)
}

func gitClone(repoURL, targetPath string, dryRun bool, progress *tap.Progress) error {
	if _, err := exec.LookPath("git"); err != nil {
		msg := fmt.Sprintf("❌ [ERROR]: git is not installed; %s", repoURL)
		progress.Message(msg)
		time.Sleep(time.Millisecond * 100)

		return fmt.Errorf("%s", msg)
	}

	parentDir := filepath.Dir(targetPath)
	if !dryRun {
		if err := os.MkdirAll(parentDir, 0755); err != nil {
			msg := fmt.Sprintf("❌ [FAILED]: to create directory: %s", parentDir)
			progress.Message(msg)
			time.Sleep(time.Millisecond * 100)

			return fmt.Errorf("%s", msg)
		}
	}

	if _, err := os.Stat(targetPath); err == nil {
		msg := fmt.Sprintf("⚠️ [SKIPPED]: %s already exists.", filepath.Base(targetPath))
		progress.Message(msg)
		time.Sleep(time.Millisecond * 100)

		return fmt.Errorf("%s", msg)
	}

	cmdStr := fmt.Sprintf("git clone --depth 1 %s %s", repoURL, targetPath)
	return utils.RunCmd(cmdStr, dryRun, progress)
}

func installDocker(dryRun bool, progress *tap.Progress) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("automated docker installation is only supported on Linux")
	}

	if dryRun {
		return utils.RunCmd("sudo sh /tmp/get-docker.sh", dryRun, progress)
	}

	tempScript := filepath.Join(os.TempDir(), "get-docker.sh")

	data, err := assets.Files.ReadFile("scripts/get-docker.sh")
	if err != nil {
		progress.Message(fmt.Sprintf("❌ [ERROR]: Failed to read docker script: %v", err))
		return fmt.Errorf("read docker script: %w", err)
	}

	if err := os.WriteFile(tempScript, data, 0755); err != nil {
		progress.Message(fmt.Sprintf("❌ [ERROR]: Failed to write temp script: %v", err))
		return fmt.Errorf("write temp script: %w", err)
	}
	defer os.Remove(tempScript)

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

func ensureMacOSPrereqs(pm string, dryRun bool, progress *tap.Progress, failedPkgs *[]string) {
	_, err := exec.LookPath("xcode-select")
	if err != nil {
		if dryRun {
			progress.Advance(1, utils.Style("___ [DRY_RUN]: Would ensure xcode-select is installed ___", "orange"))
			time.Sleep(time.Millisecond * 100)
		} else {
			cmdErr := utils.RunCmd("xcode-select --install", dryRun, progress)
			if cmdErr != nil {
				*failedPkgs = append(*failedPkgs, "xcode")
			}
			progress.Advance(1, "📦 [INSTALLING]: Xcode Command Line Tools...")
			time.Sleep(time.Millisecond * 100)
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
			time.Sleep(time.Millisecond * 100)
		}
	case "macports":
		if _, err := exec.LookPath("port"); err != nil {
			cmdErr := installMacPorts(dryRun, progress)

			if cmdErr != nil {
				*failedPkgs = append(*failedPkgs, "macports")
			}
			progress.Advance(1, "📦 [INSTALLING]: Macports...")
			time.Sleep(time.Millisecond * 100)
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
		time.Sleep(time.Millisecond * 100)
		return fmt.Errorf("macOS %s not in auto-install list", versionStr)
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
		msg := fmt.Sprintf(utils.Style("___ [DRY_RUN]: %s. Would download: %s ___", "orange"), versionStr, downloadURL)
		progress.Message(msg)
		time.Sleep(time.Millisecond * 100)
		return nil
	}

	dlMsg := fmt.Sprintf("↓ [DOWNLOADING]: MacPorts %s for %s...", pkgName, osName)
	progress.Message(dlMsg)
	time.Sleep(time.Millisecond * 100)

	cmdErrDownload := utils.RunCmd(fmt.Sprintf("curl -O %s", downloadURL), false, progress)
	if cmdErrDownload != nil {
		return cmdErrDownload
	}

	cmdErrInstall := utils.RunCmd(fmt.Sprintf("sudo installer -pkg %s -target /", pkgName), false, progress)
	_ = os.Remove(pkgName)

	if cmdErrInstall != nil {
		return cmdErrInstall
	}

	if !dryRun {
		_ = utils.RunCmd("sudo /opt/local/bin/port selfupdate", false, progress)
	}

	return nil
}
