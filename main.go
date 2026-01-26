package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/huffmanks/stash/internal/setup"
	"github.com/huffmanks/stash/internal/ui"
	"github.com/huffmanks/stash/internal/utils"
)

var version = "v0.0.dev"

func main() {
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

	latest := utils.GetLatestVersion(version)

	if *showVersion {
		title := fmt.Sprintf("Current version: [%s]", utils.Style(version, "bold", "green"))
		description := fmt.Sprintf(utils.Style("Latest version: [%s]", "bold"), utils.Style(latest, "bold", "cyan"))
		banner := ui.DisplayBanner(title, description)

		utils.HandleVersion(banner)
	}

	args := flag.Args()
	command := ""
	if len(args) > 0 {
		command = args[0]
	}

	switch command {
	case "update":
		updateCmd := flag.NewFlagSet("update", flag.ExitOnError)
		force := updateCmd.Bool("force", false, "Force reinstall")
		updateCmd.BoolVar(force, "f", false, "Force reinstall (shorthand)")

		updateCmd.Parse(args[1:])

		description := fmt.Sprintf("Current version: [%s]", utils.Style(version, "bold", "green"))
		banner := ui.DisplayBanner("Update", description)

		utils.HandleUpdate(banner, *force, latest)

	case "uninstall":
		title := fmt.Sprintf("Uninstalling stash: [%s]", utils.Style(version, "bold", "green"))
		banner := ui.DisplayBanner(title, utils.Style("This will remove the binary from your system.", "dim"))
		utils.HandleUninstall(banner)

	case "version":
		title := fmt.Sprintf("Current version: [%s]", utils.Style(version, "bold", "green"))
		description := fmt.Sprintf("Latest version: [%s]", utils.Style(latest, "bold", "cyan"))
		banner := ui.DisplayBanner(title, description)

		utils.HandleVersion(banner)

	case "help":
		flag.Usage()

	case "":
		conf, err := ui.RunPrompts(*dryRun, version)
		if err != nil {
			log.Fatal(err)
		}

		err = setup.ExecuteSetup(conf, *dryRun)
		if err != nil {
			log.Fatal(err)
		}

	default:
		fmt.Printf("Unknown command: %s\n", command)
		flag.Usage()
		os.Exit(1)
	}

}
