# ----- Docker -----
fpath=("$HOME/.docker/completions" $fpath)
autoload -Uz compinit
compinit

# ----- Functions -----
dr() {
    local target="${1:-$(basename "${(%):-%N}")}"
    local selected

    case "$target" in
        (dri|image|images)
            selected=$(docker image ls --format '{{.Repository}}:{{.Tag}}\t{{.ID}}' | fzf --ansi) || return
            docker image rm "${selected##*$'\t'}"
            ;;
        (drc|container|containers)
            selected=$(docker ps -a --format '{{.Names}}\t{{.ID}}' | fzf --ansi) || return
            docker rm "${selected##*$'\t'}"
            ;;
        (drv|volume|volumes)
            selected=$(docker volume ls --format '{{.Name}}' | fzf --ansi) || return
            docker volume rm "$selected"
            ;;
        (drn|network|networks)
            selected=$(docker network ls --format '{{.Name}}\t{{.ID}}' | fzf --ansi) || return
            docker network rm "${selected##*$'\t'}"
            ;;
        (*)
            echo "Usage: dr [dri|drc|drv|drn]"
            return 1
            ;;
    esac
}

# ----- Aliases -----
alias dcp='docker compose pull'
alias dcu='docker compose up -d'
alias dcd='docker compose down'
alias dcl='docker compose logs -f'
alias dri='dr dri'
alias drc='dr drc'
alias drv='dr drv'
alias drn='dr drn'
