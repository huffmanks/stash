package setup

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/utils"
)

func installSystemPkgs(c *config.Config, dryRun bool) error {
	if runtime.GOOS == "darwin" {
		ensureMacOSPrereqs(c.PackageManager, dryRun)
	}

	if runtime.GOOS == "linux" {
        hasPlugins := slices.Contains(c.SelectedPkgs, "zsh-syntax-highlighting") ||
                      slices.Contains(c.SelectedPkgs, "zsh-autosuggestions")
        if hasPlugins && !slices.Contains(c.SelectedPkgs, "zsh") {
            c.SelectedPkgs = append(c.SelectedPkgs, "zsh")
        }
    }

	for _, pkg := range c.SelectedPkgs {
		if runtime.GOOS != "linux" && pkg != "docker" {
			fmt.Printf("\n📦 Installing %s...\n", pkg)
		}
		switch pkg {
		case "bun":
			utils.RunCmd("curl -fsSL https://bun.com/install | bash", dryRun)
		case "docker":
			if runtime.GOOS == "linux" {
				installDocker(dryRun)
    		}
		case "go":
    		installGo(dryRun)
		case "java-android-studio":
    		installZulu(c.PackageManager, dryRun)
		case "nvm":
			utils.RunCmd("curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.2/install.sh | bash", dryRun)
		case "pnpm":
			utils.RunCmd("curl -fsSL https://get.pnpm.io/install.sh | sh -", dryRun)
		default:
			installViaPM(c.PackageManager, pkg, dryRun)
		}

		if pkg == "zsh" && runtime.GOOS == "linux" {
            utils.RunCmd("sudo chsh -s $(which zsh) $(whoami)", dryRun)
        }
        if pkg == "bat" && runtime.GOOS == "linux" {
            aliasCmd := `if command -v batcat &>/dev/null && ! command -v bat &>/dev/null; then sudo update-alternatives --install /usr/local/bin/bat bat /usr/bin/batcat 1; fi`
            utils.RunCmd(aliasCmd, dryRun)
        }
	}
	return nil
}

func installViaPM(pm, pkg string, dryRun bool) {
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

	if cmdStr != "" {
		utils.RunCmd(cmdStr, dryRun)
	}
}

var dockerScript embed.FS

func installDocker(dryRun bool) {
    tempScript := filepath.Join(os.TempDir(), "get-docker.sh")

    if !dryRun {
        data, err := dockerScript.ReadFile("scripts/get-docker.sh")
        if err != nil {
            fmt.Printf("[ERROR]: Failed to read embedded script: %v\n", err)
            return
        }

        err = os.WriteFile(tempScript, data, 0755)
        if err != nil {
            fmt.Printf("[ERROR]: Failed to write temp script: %v\n", err)
            return
        }

        defer os.Remove(tempScript)
    }

    utils.RunCmd(fmt.Sprintf("sudo sh %s", tempScript), dryRun)
}

func installGo(dryRun bool) {
	version := "1.25.5"
	if !dryRun {
		out, err := exec.Command("sh", "-c", "curl -s 'https://go.dev/VERSION?m=text' | head -n 1").Output()
		if err == nil {
			version = strings.TrimPrefix(strings.TrimSpace(string(out)), "go")
		}
	}

	var url string
	if runtime.GOOS == "darwin" {
		url = fmt.Sprintf("https://go.dev/dl/go%s.darwin-%s.pkg", version, runtime.GOARCH)
		utils.RunCmd(fmt.Sprintf("curl -LO %s && sudo installer -pkg go%s.darwin-%s.pkg -target /", url, version, runtime.GOARCH), dryRun)
	} else {
		url = fmt.Sprintf("https://go.dev/dl/go%s.linux-%s.tar.gz", version, runtime.GOARCH)
		utils.RunCmd(fmt.Sprintf("curl -L %s | sudo tar -C /usr/local -xzf -", url), dryRun)
	}
}

func installZulu(pm string, dryRun bool) {
	var cmdStr string
	switch pm {
	case "homebrew":
		cmdStr = "brew install --cask zulu@17"
	case "macports":
		cmdStr = "sudo port install openjdk17-zulu"
	}

	if cmdStr != "" {
		utils.RunCmd(cmdStr, dryRun)
	}
}

func ensureMacOSPrereqs(pm string, dryRun bool) {
	_, err := exec.LookPath("xcode-select")
    if err != nil {
        if dryRun {
            fmt.Println("[DRY-RUN]: Would ensure xcode-select is installed")
        } else {
            fmt.Println("📦 [INSTALLING]: Xcode Command Line Tools...")
            _ = exec.Command("xcode-select", "--install").Run()
        }
    }

	switch pm {
	case "homebrew":
		if _, err := exec.LookPath("brew"); err != nil {
			utils.RunCmd(`/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`, dryRun)
		}
	case "macports":
		if _, err := exec.LookPath("port"); err != nil {
			installMacPorts(dryRun)
		}
	}
}

func installMacPorts(dryRun bool) {
	out, _ := exec.Command("sw_vers", "-productVersion").Output()
	versionStr := strings.TrimSpace(string(out))

	var osName string
	switch {
	case strings.HasPrefix(versionStr, "26"): osName = "26-Tahoe"
	case strings.HasPrefix(versionStr, "15"): osName = "15-Sequoia"
	case strings.HasPrefix(versionStr, "14"): osName = "14-Sonoma"
	case strings.HasPrefix(versionStr, "13"): osName = "13-Ventura"
	case strings.HasPrefix(versionStr, "12"): osName = "12-Monterey"
	case strings.HasPrefix(versionStr, "11"): osName = "11-BigSur"
	default:
		fmt.Printf("[WARNING]: macOS %s not in auto-install list.\n", versionStr)
		return
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
		fmt.Printf("[DRY-RUN] OS: %s. Would download: %s\n", versionStr, downloadURL)
		return
	}

	fmt.Printf("[DOWNLOADING]: MacPorts %s for %s...\n", pkgName, osName)
	utils.RunCmd(fmt.Sprintf("curl -O %s", downloadURL), false)

	fmt.Println("[WARNING]: Root privileges required to run installer...")
	utils.RunCmd(fmt.Sprintf("sudo installer -pkg %s -target /", pkgName), false)
	_ = os.Remove(pkgName)
}