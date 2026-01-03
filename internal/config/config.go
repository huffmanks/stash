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