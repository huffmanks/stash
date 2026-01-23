package assets

import "embed"

//go:embed all:.dotfiles all:scripts
var Files embed.FS
