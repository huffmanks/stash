package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/huffmanks/config-stash/internal/setup"
	"github.com/huffmanks/config-stash/internal/ui"
)

func main() {
    dryRun := flag.Bool("dry-run", false, "description")
    flag.Parse()

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