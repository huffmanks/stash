package config

type Config struct {
	InstallPackages bool
	PackageManager  string
	SelectedPkgs    []string
	BuildFiles      []string
	GitName         string
	GitEmail        string
	GitBranch       string
}

type MacPortRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}
