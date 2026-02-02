package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

var Version = "dev_x.x.x"

type Config struct {
	App            string   `json:"app"`
	Version        string   `json:"version"`
	Operation      string   `json:"operation"`
	PackageManager string   `json:"package_manager"`
	BuildFiles     []string `json:"build_files"`
	GitName        string   `json:"git_name"`
	GitEmail       string   `json:"git_email"`
	GitBranch      string   `json:"git_branch"`
	GHPath         string   `json:"-"`
	SelectedPkgs   []string `json:"selected_pkgs"`
	Confirm        bool     `json:"-"`
	StartOver      bool     `json:"-"`
}

func Load() (*Config, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".config", "stash", "config.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{
			SelectedPkgs: []string{},
			BuildFiles:   []string{},
		}, err
	}

	var conf Config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}

func (c *Config) Save() error {
	c.App = "stash"
	c.Version = Version

	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "stash")
	path := filepath.Join(configDir, "config.json")

	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return err
		}
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

type MacPortRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

type DeleteResult struct {
	Deleted []string
	Failed  []string
}
