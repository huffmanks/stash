package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/huffmanks/stash/internal/config"
	"github.com/huffmanks/stash/internal/setup"
	"github.com/huffmanks/stash/internal/ui"
	"github.com/huffmanks/stash/internal/utils"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func run() error {

	dryRun := flag.Bool("dry-run", false, "Run without making changes")
	flag.BoolVar(dryRun, "d", false, "Run without making changes (shorthand)")

	showVersion := flag.Bool("version", false, "Show version")
	flag.BoolVar(showVersion, "v", false, "Show version (shorthand)")

	flag.Usage = func() {
		fmt.Println("Usage: stash [command] [flags]")
		fmt.Println("\nCommands:")
		fmt.Println("  (default)   Run setup and configuration")
		fmt.Println("  update      Update stash to the latest version")
		fmt.Println("  uninstall   Remove stash and configs")
		fmt.Println("  version     Show version information")
		fmt.Println("  help        Show this help menu")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
	}

	flag.Parse()
	args := flag.Args()
	command := ""

	if len(args) > 0 {
		command = args[0]
	}

	latest := utils.GetLatestVersion(config.Version)

	if *showVersion || command == "version" {
		title := fmt.Sprintf("Current version: [%s]", utils.Style(config.Version, "bold", "green"))
		description := fmt.Sprintf(utils.Style("Latest version: [%s]", "bold"), utils.Style(latest, "bold", "cyan"))
		banner := ui.DisplayBanner(title, description)

		utils.HandleVersion(banner)
		return nil
	}

	switch command {
	case "update":
		updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
		force := updateCmd.Bool("force", false, "Force reinstall")
		updateCmd.BoolVar(force, "f", false, "Force reinstall (shorthand)")

		updateCmd.Parse(args[1:])

		description := fmt.Sprintf("Current version: [%s]", utils.Style(config.Version, "bold", "green"))
		banner := ui.DisplayBanner("Update", description)

		return utils.HandleUpdate(banner, *force, latest)

	case "uninstall":
		title := fmt.Sprintf("Uninstalling stash: [%s]", utils.Style(config.Version, "bold", "green"))
		banner := ui.DisplayBanner(title, utils.Style("This will remove the binary from your system.", "dim"))
		return utils.HandleUninstall(banner)

	case "help":
		flag.Usage()

	case "":
		conf, err := ui.RunPrompts(*dryRun, config.Version)
		if err != nil {
			return err
		}

		if conf == nil {
			return nil
		}

		return setup.ExecuteSetup(conf, *dryRun)

	default:
		return fmt.Errorf("unknown command: %s", command)
	}

	return nil
}
