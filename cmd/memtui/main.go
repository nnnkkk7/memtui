// Package main provides the entry point for the memtui CLI application.
package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/config"
)

// version is set by goreleaser via ldflags
var version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// CLI flags
	addr := flag.String("addr", "", "Memcached server address (overrides config)")
	showVersion := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *showVersion {
		fmt.Printf("memtui version %s\n", version)
		return nil
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// CLI flag overrides config
	serverAddr := cfg.Connection.DefaultAddress
	if *addr != "" {
		serverAddr = *addr
	}

	// Create and run the TUI with config
	m := app.NewModelWithConfig(serverAddr, cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run program: %w", err)
	}

	return nil
}
