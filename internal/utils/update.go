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

	execPath, err := os.Executable()
	if err != nil {
		installDir := "/usr/local/bin"
		if os.Getuid() != 0 {
			if home, err := os.UserHomeDir(); err == nil {
				installDir = filepath.Join(home, ".local", "bin")
			}
		}
		execPath = filepath.Join(installDir, "stash")
	}

	var cpCmd, chmodCmd *exec.Cmd
	if strings.HasPrefix(execPath, "/usr/local/bin") && os.Getuid() != 0 {
		cpCmd = exec.Command("sudo", "cp", extractedBinaryPath, execPath)
		chmodCmd = exec.Command("sudo", "chmod", "+x", execPath)
	} else {
		if err := os.MkdirAll(filepath.Dir(execPath), 0755); err != nil {
			spinner.Stop("❌ [FAILED]: creating install directory.", 2)
			os.Exit(1)
		}
		cpCmd = exec.Command("cp", extractedBinaryPath, execPath)
		chmodCmd = exec.Command("chmod", "+x", execPath)
	}

	cpCmd.Stdout = os.Stdout
	cpCmd.Stderr = os.Stderr
	if err := cpCmd.Run(); err != nil {
		spinner.Stop("❌ [FAILED]: installing binary to destination.", 2)
		os.Exit(1)
	}

	if err := chmodCmd.Run(); err != nil {
		spinner.Stop("❌ [FAILED]: setting executable permissions.", 2)
		os.Exit(1)
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
