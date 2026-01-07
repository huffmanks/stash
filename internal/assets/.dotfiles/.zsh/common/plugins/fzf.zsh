export FZF_DEFAULT_OPTS="
  --layout=reverse
  --height=80%
  --border
"

export FZF_CTRL_T_OPTS="--preview 'cat {}'"
export FZF_COMPLETION_TRIGGER='**'

if command -v fd > /dev/null; then
  export FZF_DEFAULT_COMMAND='fd --type f --strip-cwd-prefix --hidden --exclude .git'
  export FZF_CTRL_T_COMMAND="$FZF_DEFAULT_COMMAND"
fi

typeset -g ZSH_FZF_HISTORY_SEARCH_DATES_IN_SEARCH=1
typeset -g ZSH_FZF_HISTORY_SEARCH_EVENT_NUMBERS=0
typeset -g ZSH_FZF_HISTORY_SEARCH_FZF_ARGS='--layout=reverse --height=80% --border --no-preview'

fzf_history_search() {
  setopt extendedglob
  local FC_ARGS="-l"
  local CANDIDATE_LEADING_FIELDS=2

  if (( ! $ZSH_FZF_HISTORY_SEARCH_EVENT_NUMBERS )); then
    FC_ARGS+=" -n"
    ((CANDIDATE_LEADING_FIELDS--))
  fi

  if (( $ZSH_FZF_HISTORY_SEARCH_DATES_IN_SEARCH )); then
    FC_ARGS+=" -i"
    ((CANDIDATE_LEADING_FIELDS+=2))
  fi

  local history_cmd="fc ${=FC_ARGS} -1 0"

  local candidates
  if (( $#BUFFER )); then
    candidates=(${(f)"$(eval $history_cmd | fzf ${=ZSH_FZF_HISTORY_SEARCH_FZF_ARGS} -q "$BUFFER")"})
  else
    candidates=(${(f)"$(eval $history_cmd | fzf ${=ZSH_FZF_HISTORY_SEARCH_FZF_ARGS})"})
  fi

  local ret=$?
  if [ -n "$candidates" ]; then
    BUFFER="${candidates[@]/(#m)[0-9 \-\:\*]##/$(
      printf '%s' "${${(As: :)MATCH}[${CANDIDATE_LEADING_FIELDS},-1]}" | sed 's/%/%%/g'
    )}"
    zle end-of-line
  fi
  zle reset-prompt
  return $ret
}

zle -N fzf_history_search
bindkey '^r' fzf_history_search
