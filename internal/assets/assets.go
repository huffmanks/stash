package assets

import "embed"

//go:embed all:.dotfiles/.zsh all:.dotfiles/git scripts
var Files embed.FS
