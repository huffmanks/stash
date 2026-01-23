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

var version = "dev"

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

	if *showVersion {
		title := fmt.Sprintf("Version: %s", version)
		banner := ui.DisplayBanner(title)

		utils.HandleVersion(banner)
		os.Exit(0)
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

		title := fmt.Sprintf("Updating from version: %s", version)
		banner := ui.DisplayBanner(title)

		err := utils.HandleUpdate(banner, *force)
		if err != nil {
			fmt.Printf("❌ [ERROR]: Update failed: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)

	case "uninstall":
		banner := ui.DisplayBanner("Uninstalling stash", "This will remove the binary from your system.")
		utils.HandleUninstall(banner)

	case "version":
		title := fmt.Sprintf("Version: %s", version)
		banner := ui.DisplayBanner(title)

		utils.HandleVersion(banner)
		os.Exit(0)

	case "help":
		flag.Usage()

	case "":
		conf, err := ui.RunPrompts(*dryRun)
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
