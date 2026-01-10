package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nnnkkk7/memtui/config"
)

func TestConfig_Defaults(t *testing.T) {
	cfg := config.DefaultConfig()

	// Test connection defaults
	if cfg.Connection.DefaultAddress != "localhost:11211" {
		t.Errorf("expected default address 'localhost:11211', got '%s'", cfg.Connection.DefaultAddress)
	}
	if cfg.Connection.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got '%v'", cfg.Connection.Timeout)
	}

	// Test timeout defaults
	if cfg.Timeouts.Connection != 3*time.Second {
		t.Errorf("expected default timeouts.connection 3s, got '%v'", cfg.Timeouts.Connection)
	}
	if cfg.Timeouts.KeyEnumeration != 30*time.Second {
		t.Errorf("expected default timeouts.key_enumeration 30s, got '%v'", cfg.Timeouts.KeyEnumeration)
	}
	if cfg.Timeouts.Capability != 5*time.Second {
		t.Errorf("expected default timeouts.capability 5s, got '%v'", cfg.Timeouts.Capability)
	}

	// Test layout defaults
	if cfg.Layout.KeyListWidthPercent != 30 {
		t.Errorf("expected default layout.keylist_width_percent 30, got '%d'", cfg.Layout.KeyListWidthPercent)
	}
	if cfg.Layout.ContentPadding != 4 {
		t.Errorf("expected default layout.content_padding 4, got '%d'", cfg.Layout.ContentPadding)
	}

	// Test UI defaults
	if cfg.UI.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", cfg.UI.Theme)
	}
	if cfg.UI.KeyDelimiter != ":" {
		t.Errorf("expected default key delimiter ':', got '%s'", cfg.UI.KeyDelimiter)
	}
	if cfg.UI.DefaultViewMode != "auto" {
		t.Errorf("expected default view mode 'auto', got '%s'", cfg.UI.DefaultViewMode)
	}

	// Test keybindings defaults
	if cfg.Keybindings.CommandPalette != "ctrl+p" {
		t.Errorf("expected default keybindings.command_palette 'ctrl+p', got '%s'", cfg.Keybindings.CommandPalette)
	}
	if cfg.Keybindings.Help != "?" {
		t.Errorf("expected default keybindings.help '?', got '%s'", cfg.Keybindings.Help)
	}
	if cfg.Keybindings.Refresh != "r" {
		t.Errorf("expected default keybindings.refresh 'r', got '%s'", cfg.Keybindings.Refresh)
	}
	if cfg.Keybindings.Delete != "d" {
		t.Errorf("expected default keybindings.delete 'd', got '%s'", cfg.Keybindings.Delete)
	}
	if cfg.Keybindings.Quit != "q" {
		t.Errorf("expected default keybindings.quit 'q', got '%s'", cfg.Keybindings.Quit)
	}
}

func TestConfig_Load_NoFile(t *testing.T) {
	// Create a temp directory for config
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Load should return defaults when no file exists
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have default values
	if cfg.Connection.DefaultAddress != "localhost:11211" {
		t.Errorf("expected default address, got '%s'", cfg.Connection.DefaultAddress)
	}
	if cfg.UI.Theme != "dark" {
		t.Errorf("expected default theme, got '%s'", cfg.UI.Theme)
	}
}

func TestConfig_Load_FromYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory and file
	configDir := filepath.Join(tmpDir, "memtui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	yamlContent := `connection:
  default_address: "myserver:11211"
  timeout: 10s

ui:
  theme: "light"
  key_delimiter: "/"
  default_view_mode: "json"
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify loaded values
	if cfg.Connection.DefaultAddress != "myserver:11211" {
		t.Errorf("expected address 'myserver:11211', got '%s'", cfg.Connection.DefaultAddress)
	}
	if cfg.Connection.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got '%v'", cfg.Connection.Timeout)
	}
	if cfg.UI.Theme != "light" {
		t.Errorf("expected theme 'light', got '%s'", cfg.UI.Theme)
	}
	if cfg.UI.KeyDelimiter != "/" {
		t.Errorf("expected key delimiter '/', got '%s'", cfg.UI.KeyDelimiter)
	}
	if cfg.UI.DefaultViewMode != "json" {
		t.Errorf("expected view mode 'json', got '%s'", cfg.UI.DefaultViewMode)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := &config.Config{
		Connection: config.ConnectionConfig{
			DefaultAddress: "testserver:11211",
			Timeout:        15 * time.Second,
		},
		Timeouts: config.TimeoutConfig{
			Connection:     5 * time.Second,
			KeyEnumeration: 60 * time.Second,
			Capability:     10 * time.Second,
		},
		Layout: config.LayoutConfig{
			KeyListWidthPercent: 40,
			ContentPadding:      6,
		},
		UI: config.UIConfig{
			Theme:           "light",
			KeyDelimiter:    "-",
			DefaultViewMode: "hex",
		},
	}

	err := config.Save(cfg)
	if err != nil {
		t.Fatalf("unexpected error saving config: %v", err)
	}

	// Verify file was created
	configPath := filepath.Join(tmpDir, "memtui", "config.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatalf("config file was not created at %s", configPath)
	}

	// Load it back and verify
	loadedCfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error loading saved config: %v", err)
	}

	if loadedCfg.Connection.DefaultAddress != cfg.Connection.DefaultAddress {
		t.Errorf("address mismatch: expected '%s', got '%s'",
			cfg.Connection.DefaultAddress, loadedCfg.Connection.DefaultAddress)
	}
	if loadedCfg.Connection.Timeout != cfg.Connection.Timeout {
		t.Errorf("timeout mismatch: expected '%v', got '%v'",
			cfg.Connection.Timeout, loadedCfg.Connection.Timeout)
	}
	if loadedCfg.Timeouts.Connection != cfg.Timeouts.Connection {
		t.Errorf("timeouts.connection mismatch: expected '%v', got '%v'",
			cfg.Timeouts.Connection, loadedCfg.Timeouts.Connection)
	}
	if loadedCfg.Layout.KeyListWidthPercent != cfg.Layout.KeyListWidthPercent {
		t.Errorf("layout.keylist_width_percent mismatch: expected '%d', got '%d'",
			cfg.Layout.KeyListWidthPercent, loadedCfg.Layout.KeyListWidthPercent)
	}
	if loadedCfg.UI.Theme != cfg.UI.Theme {
		t.Errorf("theme mismatch: expected '%s', got '%s'",
			cfg.UI.Theme, loadedCfg.UI.Theme)
	}
	if loadedCfg.UI.KeyDelimiter != cfg.UI.KeyDelimiter {
		t.Errorf("key delimiter mismatch: expected '%s', got '%s'",
			cfg.UI.KeyDelimiter, loadedCfg.UI.KeyDelimiter)
	}
	if loadedCfg.UI.DefaultViewMode != cfg.UI.DefaultViewMode {
		t.Errorf("view mode mismatch: expected '%s', got '%s'",
			cfg.UI.DefaultViewMode, loadedCfg.UI.DefaultViewMode)
	}
}

func TestConfig_ConfigPath(t *testing.T) {
	// Test with XDG_CONFIG_HOME set
	t.Run("with XDG_CONFIG_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", tmpDir)

		path := config.ConfigPath()
		expected := filepath.Join(tmpDir, "memtui", "config.yaml")
		if path != expected {
			t.Errorf("expected path '%s', got '%s'", expected, path)
		}
	})

	// Test without XDG_CONFIG_HOME (should use HOME/.config)
	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		tmpDir := t.TempDir()
		t.Setenv("HOME", tmpDir)
		t.Setenv("XDG_CONFIG_HOME", "")

		path := config.ConfigPath()
		expected := filepath.Join(tmpDir, ".config", "memtui", "config.yaml")
		if path != expected {
			t.Errorf("expected path '%s', got '%s'", expected, path)
		}
	})
}

func TestConfig_MergeWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory and file with partial config
	configDir := filepath.Join(tmpDir, "memtui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	// Only specify some values, others should get defaults
	yamlContent := `connection:
  default_address: "customserver:11211"
ui:
  theme: "light"
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify custom values are loaded
	if cfg.Connection.DefaultAddress != "customserver:11211" {
		t.Errorf("expected custom address, got '%s'", cfg.Connection.DefaultAddress)
	}
	if cfg.UI.Theme != "light" {
		t.Errorf("expected custom theme 'light', got '%s'", cfg.UI.Theme)
	}

	// Verify default values are applied for missing fields
	if cfg.Connection.Timeout != 5*time.Second {
		t.Errorf("expected default timeout 5s, got '%v'", cfg.Connection.Timeout)
	}
	if cfg.UI.KeyDelimiter != ":" {
		t.Errorf("expected default key delimiter ':', got '%s'", cfg.UI.KeyDelimiter)
	}
	if cfg.UI.DefaultViewMode != "auto" {
		t.Errorf("expected default view mode 'auto', got '%s'", cfg.UI.DefaultViewMode)
	}
}

func TestConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create config directory and file with invalid YAML
	configDir := filepath.Join(tmpDir, "memtui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatalf("failed to create config dir: %v", err)
	}

	invalidYAML := `connection:
  default_address: [invalid yaml
`
	configPath := filepath.Join(configDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}

	_, err := config.Load()
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestConfig_ConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	dir := config.ConfigDir()
	expected := filepath.Join(tmpDir, "memtui")
	if dir != expected {
		t.Errorf("expected dir '%s', got '%s'", expected, dir)
	}
}

func TestConfig_Validate(t *testing.T) {
	// Helper to create a valid base config
	validBase := func() *config.Config {
		return &config.Config{
			Connection: config.ConnectionConfig{
				DefaultAddress: "localhost:11211",
				Timeout:        5 * time.Second,
			},
			Timeouts: config.TimeoutConfig{
				Connection:     3 * time.Second,
				KeyEnumeration: 30 * time.Second,
				Capability:     5 * time.Second,
			},
			Layout: config.LayoutConfig{
				KeyListWidthPercent: 30,
				ContentPadding:      4,
			},
			UI: config.UIConfig{
				Theme:           "dark",
				KeyDelimiter:    ":",
				DefaultViewMode: "auto",
			},
		}
	}

	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     config.DefaultConfig(),
			wantErr: false,
		},
		{
			name: "invalid theme",
			cfg: func() *config.Config {
				c := validBase()
				c.UI.Theme = "invalid"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid view mode",
			cfg: func() *config.Config {
				c := validBase()
				c.UI.DefaultViewMode = "invalid"
				return c
			}(),
			wantErr: true,
		},
		{
			name: "empty address",
			cfg: func() *config.Config {
				c := validBase()
				c.Connection.DefaultAddress = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "zero connection timeout",
			cfg: func() *config.Config {
				c := validBase()
				c.Connection.Timeout = 0
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid keylist width (too small)",
			cfg: func() *config.Config {
				c := validBase()
				c.Layout.KeyListWidthPercent = 5
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid keylist width (too large)",
			cfg: func() *config.Config {
				c := validBase()
				c.Layout.KeyListWidthPercent = 95
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConnectionConfig_TimeoutDuration(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.Connection.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", cfg.Connection.Timeout)
	}
}

func TestUIConfig_ViewModes(t *testing.T) {
	validModes := []string{"auto", "json", "hex", "text"}
	for _, mode := range validModes {
		cfg := &config.Config{
			Connection: config.ConnectionConfig{
				DefaultAddress: "localhost:11211",
				Timeout:        5 * time.Second,
			},
			Timeouts: config.TimeoutConfig{
				Connection:     3 * time.Second,
				KeyEnumeration: 30 * time.Second,
				Capability:     5 * time.Second,
			},
			Layout: config.LayoutConfig{
				KeyListWidthPercent: 30,
				ContentPadding:      4,
			},
			UI: config.UIConfig{
				Theme:           "dark",
				KeyDelimiter:    ":",
				DefaultViewMode: mode,
			},
		}
		if err := cfg.Validate(); err != nil {
			t.Errorf("expected valid view mode '%s', got error: %v", mode, err)
		}
	}
}
