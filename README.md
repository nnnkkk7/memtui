# memtui

[![Go Version](https://img.shields.io/badge/Go-1.25.5+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen.svg)]()

A modern, intuitive TUI (Terminal User Interface) client for Memcached, built with Go and [Bubble Tea](https://github.com/charmbracelet/bubbletea).

<!-- TODO: Add animated GIF demo here
![memtui demo](docs/demo.gif)
-->

## Highlights

- **Hierarchical Key Navigation** - Browse keys organized in a tree structure with folder-like grouping
- **Smart Value Viewer** - Auto-detect and format JSON, with syntax highlighting
- **VS Code-style Command Palette** - Quick access to all commands with `Ctrl+P`
- **Real-time Key Filtering** - Fuzzy search through thousands of keys instantly
- **Safe Operations** - Confirmation dialogs for destructive operations
- **Vim-style Keybindings** - Navigate with `j`/`k`, familiar to terminal users

## Requirements

| Requirement | Version | Note |
|-------------|---------|------|
| Go | 1.25+ | For building from source |
| Memcached | **1.4.31+** | Required for `lru_crawler metadump` support (enabled by default in 1.5+) |

## Installation

### Homebrew (macOS / Linux)

```bash
brew install nnnkkk7/tap/memtui
```

### From Source

```bash
git clone https://github.com/nnnkkk7/memtui.git
cd memtui
go build -o memtui ./cmd/memtui
```

### Using Go Install

```bash
go install github.com/nnnkkk7/memtui/cmd/memtui@latest
```

## Quick Start

1. **Start Memcached**:

```bash
# Using Docker
docker run -d -p 11211:11211 memcached:latest

# Or using docker compose
docker compose up -d
```

2. **Run memtui**:

```bash
./memtui                        # Connect to localhost:11211 (default)
./memtui -addr localhost:11211  # Specify address
./memtui --help                 # Show help
```

## Keyboard Shortcuts

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `Enter` | Select key / Expand folder |
| `Tab` | Switch between key list and viewer |
| `Esc` | Return to key list / Close dialog |

### Commands

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

### In Value Viewer

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down |
| `k` / `↑` | Scroll up |
| `g` | Go to top |
| `G` | Go to bottom |

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

- **JSON** - Pretty-printed with syntax highlighting
- **Compressed data** - Auto-decompresses gzip/zlib
- **Binary data** - Displays hex dump
- **Plain text** - Shows as-is

### Command Palette

Press `Ctrl+P` to open the VS Code-style command palette for quick access to all features with fuzzy search.

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
./memtui -addr localhost:11212  # Overrides connection.default_address
```

## Architecture

```
memtui/
├── cmd/memtui/     # Application entry point
├── app/            # Main application logic (Bubble Tea model)
├── client/         # Memcached client wrapper
│   ├── capability.go   # Server capability detection
│   └── enumerator.go   # Key enumeration via metadump
├── models/         # Data models (KeyInfo, Item)
├── ui/
│   ├── components/     # Reusable UI components
│   │   ├── keylist/    # Hierarchical key list
│   │   ├── viewer/     # Value viewer with formatting
│   │   ├── command/    # Command palette
│   │   ├── dialog/     # Confirm/Input dialogs
│   │   ├── editor/     # Value editor
│   │   └── help/       # Help overlay
│   ├── layout/         # Layout utilities
│   └── styles/         # Lipgloss styles
├── viewer/         # Value parsing and formatting
└── config/         # Configuration management
```

### Built With

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework based on The Elm Architecture
- [Bubbles](https://github.com/charmbracelet/bubbles) - Common Bubble Tea components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Style definitions for terminal apps
- [gomemcache](https://github.com/bradfitz/gomemcache) - Memcached client for Go


## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.


## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

