package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/huffmanks/stash/internal/setup"
	"github.com/huffmanks/stash/internal/ui"
)

var version = "version:dev"

func main() {
    dryRun := flag.Bool("dry-run", false, "Run the setup without making actual changes")
    showVersion := flag.Bool("version", false, "Show the current version of stash")
    flag.Parse()

    if *showVersion {
		fmt.Printf("stash %s\n", version)
		os.Exit(0)
	}

    if *dryRun {
        fmt.Println("⚠️  DRY RUN MODE")
    }

    conf, err := ui.RunPrompts()
    if err != nil {
        log.Fatal(err)
    }

    err = setup.ExecuteSetup(conf, *dryRun)
    if err != nil {
        log.Fatal(err)
    }
}