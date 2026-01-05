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

var version = ":dev"

func main() {
    var dryRun bool
	var showVersion bool
    var uninstall bool

	flag.BoolVar(&dryRun, "dry-run", false, "Run without making changes")
	flag.BoolVar(&dryRun, "d", false, "Run without making changes (shorthand)")

	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&showVersion, "v", false, "Show version (shorthand)")

	flag.BoolVar(&uninstall, "uninstall", false, "Remove stash and associated configs")
	flag.BoolVar(&uninstall, "u", false, "Remove stash (shorthand)")

    flag.Parse()

    if showVersion {
		fmt.Printf("stash v%s\n", version)
		os.Exit(0)
	}

	if uninstall {
		utils.HandleUninstall()
		os.Exit(0)
	}

	if dryRun {
		fmt.Println("⚠️  DRY RUN MODE: No changes will be written to disk.")
	}

	conf, err := ui.RunPrompts()
	if err != nil {
		log.Fatal(err)
	}

	err = setup.ExecuteSetup(conf, dryRun)
	if err != nil {
		log.Fatal(err)
	}
}