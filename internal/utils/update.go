package utils

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/huffmanks/stash/internal/config"
	"github.com/yarlson/tap"
)

func HandleUpdate(banner string, force bool, latest string) {
	ctx := context.Background()

	tap.Intro(banner)

	if !force {
		msg := fmt.Sprintf("Update to version: [%s]?", Style(latest, "bold", "cyan"))
		confirmed := tap.Confirm(ctx, tap.ConfirmOptions{
			Message:      msg,
			InitialValue: false,
		})

		if !confirmed {
			tap.Outro(Style("🛑 [ABORTED]: stash remains installed.", "orange"))
			os.Exit(0)
		}
	}

	PromptForSudo("❌ [ERROR]: sudo authentication failed.", true)

	spinner := tap.NewSpinner(tap.SpinnerOptions{
		Delay: time.Millisecond * 100,
	})
	spinner.Start("Updating...")
	time.Sleep(time.Millisecond * 100)

	versionClean := strings.TrimPrefix(latest, "v")
	binaryName := fmt.Sprintf("stash_%s_%s_%s", versionClean, runtime.GOOS, runtime.GOARCH)
	downloadURL := fmt.Sprintf("https://github.com/huffmanks/stash/releases/download/%s/%s.tar.gz", latest, binaryName)

	tmpDir, err := os.MkdirTemp("", "stash-update-*")
	if err != nil {
		spinner.Stop("❌ [FAILED]: creating temporary directory.", 2)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	resp, err := http.Get(downloadURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		if resp != nil {
			resp.Body.Close()
		}
		spinner.Stop(fmt.Sprintf("❌ [FAILED]: downloading release from %s", downloadURL), 2)
		os.Exit(1)
	}
	defer resp.Body.Close()

	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		spinner.Stop("❌ [FAILED]: reading gzip stream.", 2)
		os.Exit(1)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	var extractedBinaryPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			spinner.Stop("❌ [FAILED]: reading archive.", 2)
			os.Exit(1)
		}

		target := filepath.Join(tmpDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				spinner.Stop("❌ [FAILED]: creating directory structure.", 2)
				os.Exit(1)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				spinner.Stop("❌ [FAILED]: creating parent directory.", 2)
				os.Exit(1)
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				spinner.Stop("❌ [FAILED]: extracting file.", 2)
				os.Exit(1)
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				spinner.Stop("❌ [FAILED]: writing extracted file.", 2)
				os.Exit(1)
			}
			f.Close()

			if filepath.Base(target) == "stash" {
				extractedBinaryPath = target
			}
		}
	}

	if extractedBinaryPath == "" {
		spinner.Stop("❌ [FAILED]: 'stash' binary not found in archive.", 2)
		os.Exit(1)
	}

	spinner.Message("Installing binary...")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		spinner.Stop("❌ [FAILED]: determining home directory.", 2)
		os.Exit(1)
	}

	execPath := filepath.Join(homeDir, ".local", "bin", "stash")

	if err := os.MkdirAll(filepath.Dir(execPath), 0755); err != nil {
		spinner.Stop("❌ [FAILED]: creating install directory.", 2)
		os.Exit(1)
	}

	if err := os.Chmod(extractedBinaryPath, 0755); err != nil {
		spinner.Stop("❌ [FAILED]: setting executable permissions.", 2)
		os.Exit(1)
	}

	_ = os.Remove(execPath)
	if err := os.Rename(extractedBinaryPath, execPath); err != nil {
		cpCmd := exec.Command("cp", extractedBinaryPath, execPath)
		if err := cpCmd.Run(); err != nil {
			spinner.Stop("❌ [FAILED]: installing binary to destination.", 2)
			os.Exit(1)
		}
	}

	if path, err := exec.LookPath("stash"); err == nil {
		var legacyPaths []string

		out, err := exec.Command("sh", "-c", "type -a -p stash || which -a stash").Output()
		if err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for _, line := range lines {
				p := strings.TrimSpace(line)
				if p != "" && p != execPath {
					legacyPaths = append(legacyPaths, p)
				}
			}
		} else if path != execPath {
			legacyPaths = append(legacyPaths, path)
		}

		for _, legacyPath := range legacyPaths {
			if _, err := os.Stat(legacyPath); err == nil {
				if err := os.Remove(legacyPath); err != nil {
					_ = exec.Command("sudo", "rm", "-f", legacyPath).Run()
				}
			}
		}
	}

	time.Sleep(time.Millisecond * 1000)
	spinner.Stop("Updating...", 0)

	if savedConf, err := config.Load(); err == nil {
		savedConf.Version = latest
		_ = savedConf.Save()
	}

	time.Sleep(time.Millisecond * 100)
	tap.Outro(fmt.Sprintf("✅ [UPDATED]: successfully to version [%s]", Style(latest, "bold", "cyan")))
	time.Sleep(time.Millisecond * 100)

	os.Exit(0)
}
