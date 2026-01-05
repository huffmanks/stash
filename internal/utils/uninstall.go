package utils

import (
	"fmt"
	"os"
)

func HandleUninstall() {
	fmt.Println("🗑️  Uninstalling stash...")
    err := os.Remove("/usr/local/bin/stash")
    if err != nil {
        fmt.Printf("❌ Error removing binary: %v\n", err)
        fmt.Println("Try running with sudo: sudo stash --uninstall")
        return
    }
    fmt.Println("✅ stash has been removed.")
}