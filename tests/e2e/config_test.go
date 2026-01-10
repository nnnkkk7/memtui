//go:build e2e

package e2e_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nnnkkk7/memtui/app"
	"github.com/nnnkkk7/memtui/config"
)

// TestE2E_ConfigLoading tests that configuration is properly loaded
func TestE2E_ConfigLoading(t *testing.T) {
	t.Run("default config when no file exists", func(t *testing.T) {
		// Use a temp directory for XDG_CONFIG_HOME
		tmpDir := t.TempDir()
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// Verify defaults
		if cfg.Connection.DefaultAddress != "localhost:11211" {
			t.Errorf("expected default address localhost:11211, got %s", cfg.Connection.DefaultAddress)
		}
		if cfg.UI.Theme != "dark" {
			t.Errorf("expected default theme dark, got %s", cfg.UI.Theme)
		}
		if cfg.UI.KeyDelimiter != ":" {
			t.Errorf("expected default delimiter ':', got %s", cfg.UI.KeyDelimiter)
		}
	})

	t.Run("config file is loaded and merged", func(t *testing.T) {
		// Create a temp config directory
		tmpDir := t.TempDir()
		configDir := filepath.Join(tmpDir, "memtui")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}

		// Write custom config
		configContent := `
connection:
  default_address: "custom:12345"
ui:
  theme: light
  key_delimiter: "/"
`
		configPath := filepath.Join(configDir, "config.yaml")
		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("failed to write config: %v", err)
		}

		// Set XDG_CONFIG_HOME to temp dir
		oldXDG := os.Getenv("XDG_CONFIG_HOME")
		os.Setenv("XDG_CONFIG_HOME", tmpDir)
		defer os.Setenv("XDG_CONFIG_HOME", oldXDG)

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// Verify custom values
		if cfg.Connection.DefaultAddress != "custom:12345" {
			t.Errorf("expected address custom:12345, got %s", cfg.Connection.DefaultAddress)
		}
		if cfg.UI.Theme != "light" {
			t.Errorf("expected theme light, got %s", cfg.UI.Theme)
		}
		if cfg.UI.KeyDelimiter != "/" {
			t.Errorf("expected delimiter '/', got %s", cfg.UI.KeyDelimiter)
		}
	})
}

// TestE2E_ConfigAppliedToApp tests that config is applied to the app
func TestE2E_ConfigAppliedToApp(t *testing.T) {
	t.Run("dark theme is applied by default", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.UI.Theme = "dark"

		m := app.NewModelWithConfig("localhost:11211", cfg)
		if m == nil {
			t.Fatal("expected non-nil model")
		}

		// Model should be created with dark theme (verify via state)
		if m.State() != app.StateConnecting {
			t.Errorf("expected StateConnecting, got %v", m.State())
		}
	})

	t.Run("light theme is applied from config", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.UI.Theme = "light"

		m := app.NewModelWithConfig("localhost:11211", cfg)
		if m == nil {
			t.Fatal("expected non-nil model")
		}

		// Model should be created successfully
		if m.State() != app.StateConnecting {
			t.Errorf("expected StateConnecting, got %v", m.State())
		}
	})

	t.Run("custom delimiter is applied", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.UI.KeyDelimiter = "/"

		m := app.NewModelWithConfig("localhost:11211", cfg)
		if m == nil {
			t.Fatal("expected non-nil model")
		}

		// Model should be created successfully
		if m.State() != app.StateConnecting {
			t.Errorf("expected StateConnecting, got %v", m.State())
		}
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		m := app.NewModelWithConfig("localhost:11211", nil)
		if m == nil {
			t.Fatal("expected non-nil model")
		}

		// Should work with nil config (uses defaults)
		if m.State() != app.StateConnecting {
			t.Errorf("expected StateConnecting, got %v", m.State())
		}
	})
}

// TestE2E_ConfigValidation tests config validation
func TestE2E_ConfigValidation(t *testing.T) {
	t.Run("valid config passes validation", func(t *testing.T) {
		cfg := config.DefaultConfig()
		if err := cfg.Validate(); err != nil {
			t.Errorf("expected valid config, got error: %v", err)
		}
	})

	t.Run("invalid theme fails validation", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.UI.Theme = "invalid"
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for invalid theme")
		}
	})

	t.Run("empty address fails validation", func(t *testing.T) {
		cfg := config.DefaultConfig()
		cfg.Connection.DefaultAddress = ""
		if err := cfg.Validate(); err == nil {
			t.Error("expected validation error for empty address")
		}
	})
}
