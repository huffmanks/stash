# =====================================
# Aliases (common)
# =====================================

# ----- Opinionated defaults -----
alias grep='grep --color=auto'

# ----- Traversing -----
alias ..='cd ..'
alias ...='cd ../../../'
alias ....='cd ../../../../'

# ----- Git -----
alias gs='git status'
alias ga='git add'
alias gc='git commit'
alias gp='git push'
alias gd='git diff'
alias glog='git log --oneline --graph --decorate'
alias gfu='git fetch origin && git reset --hard origin/main && git clean -fd'
alias gsu='git submodule update --remote --merge'
alias gfix='git add . && git commit --amend --no-edit && git push origin main --force-with-lease'

# ----- Docker -----
alias dcu='docker compose up -d'
alias dcd='docker compose down'
alias dcl='docker compose logs -f'

# ----- Functions -----
dri() {
    local image

    image=$(docker image ls --format '{{.Repository}}:{{.Tag}}\t{{.ID}}' | fzf) || return
    docker image rm "${image##*$'\t'}"
}

nuke-cache() {
  echo "Starting cache purge..."

  if type docker &>/dev/null && docker info &>/dev/null; then
    docker system prune -a -f
  else
    echo "Docker not running or not installed, skipping..."
  fi

  type pnpm &>/dev/null && pnpm store prune || echo "pnpm not found, skipping..."
  type npm &>/dev/null && npm cache clean --force || echo "npm not found, skipping..."
  type go &>/dev/null && go clean -cache -modcache || echo "Go not found, skipping..."
  type uv &>/dev/null && uv cache clean || echo "uv not found, skipping..."

  echo "Cache purge complete!"
}
