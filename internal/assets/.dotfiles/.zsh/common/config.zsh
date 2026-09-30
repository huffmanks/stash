# =====================================
# Config (common)
# =====================================

export TZ="America/New_York"

if [ "$TERM" = "xterm" ]; then
  export TERM="xterm-256color"
fi

# ----- zsh config -----
HISTFILE=~/.zsh_history
HISTSIZE=100000
SAVEHIST=100000

setopt EXTENDED_HISTORY
setopt HIST_EXPIRE_DUPS_FIRST
setopt HIST_IGNORE_ALL_DUPS
setopt APPEND_HISTORY
setopt SHARE_HISTORY
unsetopt HIST_IGNORE_SPACE
setopt PROMPT_SUBST
