# ----- Aliases (macOS:amd64) -----
# --- Map brew commands to macports ---
brew() {
  case "$1" in
    search)   port search "$2" ;;
    update)   sudo port selfupdate ;;
    list)     port installed ;;
    outdated) port outdated ;;
    upgrade)  sudo port upgrade outdated ;;
    cleanup)  sudo port reclaim ;;
    doctor)   port diagnose ;;
    *)        echo "Usage: brew {search|update|list|outdated|upgrade|cleanup|doctor}" ;;
  esac
}
