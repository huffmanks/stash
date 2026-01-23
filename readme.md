# stash

An interactive CLI tool to bootstrap system packages and dynamically build platform-specific zsh and git configurations.

---

## Features

- **Smart package detection:** Automatically identifies your package manager (`apt`, `brew`, `dnf`, `pacman`, `ports`).
- **Dynamic ZSH building:** Generates a `.zshrc` tailored to your OS (macOS/Linux) and architecture (Intel/ARM).
- **Modular configs:** Only includes exports and plugins for the packages you actually choose to install.

## Quick install

```sh
curl -sSL https://raw.githubusercontent.com/huffmanks/stash/main/install.sh | bash
```

### Force install

```sh
curl -sSL https://raw.githubusercontent.com/huffmanks/stash/main/install.sh | bash -s -- --force
```

Once installed, simply run the command to start the interactive prompt:

```sh
stash
```

## Commands/Flags

| Command / Flag       | Shorthand       | Description                                           |
| -------------------- | --------------- | ----------------------------------------------------- |
| stash                |                 | Runs interactive setup and configuration.             |
| stash --dry-run      | stash -d        | Preview changes without writing to disk.              |
| stash update         |                 | Updates stash to the latest version.                  |
| stash update --force | stash update -f | Bypasses version check and forces a reinstall.        |
| stash uninstall      | stash -u        | Removes stash and associated configs from the system. |
| stash version        | stash -v        | Displays the current installed version.               |
| stash help           | stash -h        | Shows the help menu and available commands.           |
