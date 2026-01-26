package utils

import (
	"os"

	"github.com/yarlson/tap"
)

func HandleVersion(banner string) {
	tap.Message(banner)
	os.Exit(0)
}
