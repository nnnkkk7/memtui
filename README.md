<div align="center">
  <img src="assets/logo.svg" alt="memtui logo" width="140">
  <h1>memtui</h1>
  <p><strong>A modern, intuitive TUI (Terminal User Interface) client for Memcached, built with Go.</strong></p>

  <br>

  [![CI](https://github.com/nnnkkk7/memtui/actions/workflows/ci.yaml/badge.svg)](https://github.com/nnnkkk7/memtui/actions/workflows/ci.yaml)
  [![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
  [![Go Report Card](https://goreportcard.com/badge/github.com/nnnkkk7/memtui)](https://goreportcard.com/report/github.com/nnnkkk7/memtui)
  [![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

  <br>

  <img src="assets/demo.gif" alt="memtui demo" width="700">
</div>

---

## Table of Contents

- [Why memtui?](#why-memtui)
- [Highlights](#highlights)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Use Cases](#use-cases)
- [Keyboard Shortcuts](#keyboard-shortcuts)
- [Features](#features)
- [Configuration](#configuration)
- [Built With](#built-with)
- [Contributing](#contributing)
- [Support](#support)
- [License](#license)

---

## Why memtui?

Ever tried to debug Memcached data? Traditional tools like `telnet` or `nc` make it painful to browse keys, understand data formats, or safely edit values.

**memtui** brings the developer experience you expect from modern tools:

- **No more raw `stats cachedump`** — See all your keys in a tree structure
- **No more guessing data formats** — Auto-detect JSON, compressed data, binary
- **No more accidental overwrites** — CAS support prevents data loss
- **No more context switching** — Stay in your terminal workflow

---

## Highlights

- **Hierarchical Key Navigation** — Browse keys organized in a tree structure with folder-like grouping
- **Smart Value Viewer** — Auto-detect and format JSON, with syntax highlighting
- **VS Code-style Command Palette** — Quick access to all commands with `Ctrl+P`
- **Real-time Key Filtering** — Fuzzy search through thousands of keys instantly
- **Safe Operations** — Confirmation dialogs for destructive operations
- **Vim-style Keybindings** — Navigate with `j`/`k`, familiar to terminal users

---

## Installation

### Homebrew (macOS / Linux)

```bash
brew install nnnkkk7/tap/memtui
```

### Using Go Install

```bash
go install github.com/nnnkkk7/memtui/cmd/memtui@latest
```

### From Source

```bash
git clone https://github.com/nnnkkk7/memtui.git
cd memtui
go build -o memtui ./cmd/memtui
```

### Requirements

| Requirement | Version | Note |
|-------------|---------|------|
| Go | 1.25+ | For building from source |
| Memcached | **1.4.31+** | Required for `lru_crawler metadump` support |

---

## Quick Start

1. **Start Memcached**:

```bash
# Using Docker
docker run -d -p 11211:11211 memcached:latest
```

2. **Run memtui**:

```bash
memtui                        # Connect to localhost:11211 (default)
memtui -addr localhost:11211  # Specify address
memtui --help                 # Show help
```

---

## Use Cases

- **Debugging** — Quickly inspect cached values during development
- **Data Migration** — View and verify data before/after migrations
- **Cache Analysis** — Understand key naming patterns and data distribution
- **Incident Response** — Rapidly inspect cache state during outages

---

## Keyboard Shortcuts

<details>
<summary><strong>Navigation</strong></summary>

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select key / Expand folder |
| `Tab` | Switch between key list and viewer |
| `Esc` | Return to key list / Close dialog |

</details>

<details>
<summary><strong>Commands</strong></summary>

| Key | Action |
|-----|--------|
| `Ctrl+P` | Open command palette |
| `/` | Filter keys (fuzzy search) |
| `r` | Refresh key list |
| `n` | Create new key |
| `e` | Edit selected key's value |
| `d` | Delete selected key |
| `c` | Copy value to clipboard |
| `?` | Show help |
| `q` | Quit |

</details>

<details>
<summary><strong>In Value Viewer</strong></summary>

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `g` | Go to top |
| `G` | Go to bottom |
| `J` | JSON view mode |
| `H` | Hex view mode |
| `T` | Text view mode |
| `A` | Auto view mode |

</details>

---

## Features

### Tree-structured Key List

Keys are automatically organized into a hierarchical tree based on common delimiters (`:`, `/`, `.`). For example:

```
user:1001:profile  ─┐
user:1001:session  ─┼→  user/
user:1002:profile  ─┘     ├── 1001/
cache:api:users    ─┐     │     ├── profile
cache:api:posts    ─┘     │     └── session
                          ├── 1002/
                          │     └── profile
                          cache/
                            └── api/
                                  ├── users
                                  └── posts
```

### Smart Value Detection

The viewer automatically detects and formats:

- **JSON** — Pretty-printed with syntax highlighting
- **Compressed data** — Auto-decompresses gzip/zlib/zstd
- **Binary data** — Displays hex dump
- **Plain text** — Shows as-is

### Command Palette

Press `Ctrl+P` to open the VS Code-style command palette for quick access to all features with fuzzy search.

---

## Configuration

Configuration file location: `~/.config/memtui/config.yaml`

```yaml
# Connection settings
connection:
  default_address: localhost:11211  # Default server address

# UI settings
ui:
  theme: dark           # dark or light
  key_delimiter: ":"    # Key hierarchy delimiter (e.g., ":", "/", ".")
```

CLI flags override config file settings:

```bash
memtui -addr localhost:11212  # Overrides connection.default_address
```

---

## Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework based on The Elm Architecture
- [Bubbles](https://github.com/charmbracelet/bubbles) — Common Bubble Tea components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Style definitions for terminal apps
- [gomemcache](https://github.com/bradfitz/gomemcache) — Memcached client for Go

---

## Contributing

Contributions are welcome! Here's how you can help:

---

## Support

If you find memtui useful, please consider giving it a star on GitHub!
It helps others discover the project and motivates continued development.

[![Star on GitHub](https://img.shields.io/github/stars/nnnkkk7/memtui?style=social)](https://github.com/nnnkkk7/memtui)

---

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
