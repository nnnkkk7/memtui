// Package config provides configuration file support for memtui.
// It supports YAML configuration files following the XDG Base Directory specification.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/goccy/go-yaml"
)

const (
	// AppName is the application name used for config directories
	AppName = "memtui"
	// ConfigFileName is the name of the configuration file
	ConfigFileName = "config.yaml"
)

// Config represents the complete configuration for memtui
type Config struct {
	Connection  ConnectionConfig  `yaml:"connection"`
	Timeouts    TimeoutConfig     `yaml:"timeouts"`
	Layout      LayoutConfig      `yaml:"layout"`
	UI          UIConfig          `yaml:"ui"`
	Keybindings KeybindingsConfig `yaml:"keybindings,omitempty"`
}

// ConnectionConfig holds connection-related settings
type ConnectionConfig struct {
	DefaultAddress string        `yaml:"default_address"`
	Timeout        time.Duration `yaml:"timeout"`
}

// TimeoutConfig holds timeout settings for various operations
type TimeoutConfig struct {
	Connection     time.Duration `yaml:"connection"`      // Client connection timeout (default: 3s)
	KeyEnumeration time.Duration `yaml:"key_enumeration"` // Key listing timeout (default: 30s)
	Capability     time.Duration `yaml:"capability"`      // Server capability detection timeout (default: 5s)
}

// LayoutConfig holds UI layout settings
type LayoutConfig struct {
	KeyListWidthPercent int `yaml:"keylist_width_percent"` // Width of key list as percentage (default: 30)
	ContentPadding      int `yaml:"content_padding"`       // Padding for content area in lines (default: 4)
}

// UIConfig holds UI-related settings
type UIConfig struct {
	Theme           string `yaml:"theme"`
	KeyDelimiter    string `yaml:"key_delimiter"`
	DefaultViewMode string `yaml:"default_view_mode"`
}

// KeybindingsConfig holds custom keybinding settings
type KeybindingsConfig struct {
	CommandPalette string `yaml:"command_palette"` // default: "ctrl+p"
	Help           string `yaml:"help"`            // default: "?"
	Refresh        string `yaml:"refresh"`         // default: "r"
	Delete         string `yaml:"delete"`          // default: "d"
	Edit           string `yaml:"edit"`            // default: "e"
	NewKey         string `yaml:"new_key"`         // default: "n"
	Filter         string `yaml:"filter"`          // default: "/"
	Quit           string `yaml:"quit"`            // default: "q"
	SwitchPane     string `yaml:"switch_pane"`     // default: "tab"
}

// Valid theme options
var validThemes = map[string]bool{
	"dark":  true,
	"light": true,
}

// Valid view mode options
var validViewModes = map[string]bool{
	"auto": true,
	"json": true,
	"hex":  true,
	"text": true,
}

// DefaultConfig returns a Config with sensible default values
func DefaultConfig() *Config {
	return &Config{
		Connection: ConnectionConfig{
			DefaultAddress: "localhost:11211",
			Timeout:        5 * time.Second,
		},
		Timeouts: TimeoutConfig{
			Connection:     3 * time.Second,
			KeyEnumeration: 30 * time.Second,
			Capability:     5 * time.Second,
		},
		Layout: LayoutConfig{
			KeyListWidthPercent: 30,
			ContentPadding:      4,
		},
		UI: UIConfig{
			Theme:           "dark",
			KeyDelimiter:    ":",
			DefaultViewMode: "auto",
		},
		Keybindings: KeybindingsConfig{
			CommandPalette: "ctrl+p",
			Help:           "?",
			Refresh:        "r",
			Delete:         "d",
			Edit:           "e",
			NewKey:         "n",
			Filter:         "/",
			Quit:           "q",
			SwitchPane:     "tab",
		},
	}
}

// ConfigDir returns the XDG-compliant configuration directory path
func ConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = os.Getenv("HOME")
		}
		configHome = filepath.Join(home, ".config")
	}
	return filepath.Join(configHome, AppName)
}

// ConfigPath returns the full path to the configuration file
func ConfigPath() string {
	return filepath.Join(ConfigDir(), ConfigFileName)
}

// Load reads the configuration from the config file.
// If the file doesn't exist, it returns the default configuration.
// Partial configurations are merged with defaults.
func Load() (*Config, error) {
	cfg := DefaultConfig()
	path := ConfigPath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file exists, return defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse the YAML into a fresh config (to detect what was actually set)
	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file config with defaults
	mergeConfig(cfg, &fileCfg)

	return cfg, nil
}

// mergeConfig merges the file configuration into the default configuration
// Only non-zero values from fileCfg override the defaults
func mergeConfig(defaults *Config, fileCfg *Config) {
	// Merge connection settings
	if fileCfg.Connection.DefaultAddress != "" {
		defaults.Connection.DefaultAddress = fileCfg.Connection.DefaultAddress
	}
	if fileCfg.Connection.Timeout != 0 {
		defaults.Connection.Timeout = fileCfg.Connection.Timeout
	}

	// Merge timeout settings
	if fileCfg.Timeouts.Connection != 0 {
		defaults.Timeouts.Connection = fileCfg.Timeouts.Connection
	}
	if fileCfg.Timeouts.KeyEnumeration != 0 {
		defaults.Timeouts.KeyEnumeration = fileCfg.Timeouts.KeyEnumeration
	}
	if fileCfg.Timeouts.Capability != 0 {
		defaults.Timeouts.Capability = fileCfg.Timeouts.Capability
	}

	// Merge layout settings
	if fileCfg.Layout.KeyListWidthPercent != 0 {
		defaults.Layout.KeyListWidthPercent = fileCfg.Layout.KeyListWidthPercent
	}
	if fileCfg.Layout.ContentPadding != 0 {
		defaults.Layout.ContentPadding = fileCfg.Layout.ContentPadding
	}

	// Merge UI settings
	if fileCfg.UI.Theme != "" {
		defaults.UI.Theme = fileCfg.UI.Theme
	}
	if fileCfg.UI.KeyDelimiter != "" {
		defaults.UI.KeyDelimiter = fileCfg.UI.KeyDelimiter
	}
	if fileCfg.UI.DefaultViewMode != "" {
		defaults.UI.DefaultViewMode = fileCfg.UI.DefaultViewMode
	}

	// Merge keybindings settings
	if fileCfg.Keybindings.CommandPalette != "" {
		defaults.Keybindings.CommandPalette = fileCfg.Keybindings.CommandPalette
	}
	if fileCfg.Keybindings.Help != "" {
		defaults.Keybindings.Help = fileCfg.Keybindings.Help
	}
	if fileCfg.Keybindings.Refresh != "" {
		defaults.Keybindings.Refresh = fileCfg.Keybindings.Refresh
	}
	if fileCfg.Keybindings.Delete != "" {
		defaults.Keybindings.Delete = fileCfg.Keybindings.Delete
	}
	if fileCfg.Keybindings.Edit != "" {
		defaults.Keybindings.Edit = fileCfg.Keybindings.Edit
	}
	if fileCfg.Keybindings.NewKey != "" {
		defaults.Keybindings.NewKey = fileCfg.Keybindings.NewKey
	}
	if fileCfg.Keybindings.Filter != "" {
		defaults.Keybindings.Filter = fileCfg.Keybindings.Filter
	}
	if fileCfg.Keybindings.Quit != "" {
		defaults.Keybindings.Quit = fileCfg.Keybindings.Quit
	}
	if fileCfg.Keybindings.SwitchPane != "" {
		defaults.Keybindings.SwitchPane = fileCfg.Keybindings.SwitchPane
	}
}

// Save writes the configuration to the config file.
// It creates the config directory if it doesn't exist.
func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	path := ConfigPath()
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration values are valid
func (c *Config) Validate() error {
	// Validate connection settings
	if c.Connection.DefaultAddress == "" {
		return errors.New("connection.default_address cannot be empty")
	}
	if c.Connection.Timeout <= 0 {
		return errors.New("connection.timeout must be positive")
	}

	// Validate timeout settings
	if c.Timeouts.Connection <= 0 {
		return errors.New("timeouts.connection must be positive")
	}
	if c.Timeouts.KeyEnumeration <= 0 {
		return errors.New("timeouts.key_enumeration must be positive")
	}
	if c.Timeouts.Capability <= 0 {
		return errors.New("timeouts.capability must be positive")
	}

	// Validate layout settings
	if c.Layout.KeyListWidthPercent < 10 || c.Layout.KeyListWidthPercent > 90 {
		return fmt.Errorf("layout.keylist_width_percent must be between 10 and 90 (got: %d)", c.Layout.KeyListWidthPercent)
	}
	if c.Layout.ContentPadding < 0 || c.Layout.ContentPadding > 20 {
		return fmt.Errorf("layout.content_padding must be between 0 and 20 (got: %d)", c.Layout.ContentPadding)
	}

	// Validate UI settings
	if !validThemes[c.UI.Theme] {
		return fmt.Errorf("ui.theme must be one of: dark, light (got: %s)", c.UI.Theme)
	}
	if !validViewModes[c.UI.DefaultViewMode] {
		return fmt.Errorf("ui.default_view_mode must be one of: auto, json, hex, text (got: %s)", c.UI.DefaultViewMode)
	}

	return nil
}

// LoadOrCreate loads the configuration from file, or creates a default config file
// if one doesn't exist. This is useful for first-run scenarios.
func LoadOrCreate() (*Config, error) {
	cfg, err := Load()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	path := ConfigPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create default config file
		if err := Save(cfg); err != nil {
			// Don't fail if we can't write, just return the defaults
			return cfg, nil
		}
	}

	return cfg, nil
}
