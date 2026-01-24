package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Operation      string   `json:"operation"`
	PackageManager string   `json:"package_manager"`
	BuildFiles     []string `json:"build_files"`
	GitName        string   `json:"git_name"`
	GitEmail       string   `json:"git_email"`
	GitBranch      string   `json:"git_branch"`
	SelectedPkgs   []string `json:"selected_pkgs"`
	Confirm        bool
}

func Load() (*Config, error) {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".stash_config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return &Config{
			SelectedPkgs: []string{},
			BuildFiles:   []string{},
		}, err
	}
	var conf Config
	err = json.Unmarshal(data, &conf)
	return &conf, err
}

func (c *Config) Save() error {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".stash_config.json")
	data, _ := json.MarshalIndent(c, "", "  ")
	return os.WriteFile(path, data, 0644)
}

type MacPortRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}
