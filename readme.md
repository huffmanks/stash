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

Once installed, simply run the command to start the interactive prompt:

```sh
stash
```

## Flags

| Flag      | Shorthand | Description                                                            |
| --------- | --------- | ---------------------------------------------------------------------- |
| --dry-run | -d        | Preview which files would be created/modified without writing to disk. |
| --version | -v        | Display the current installed version of stash.                        |
