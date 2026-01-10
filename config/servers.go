// Package config provides server configuration management for memtui.
// It supports storing and managing multiple Memcached server configurations.
package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/goccy/go-yaml"
)

const (
	// ServersFileName is the name of the servers configuration file
	ServersFileName = "servers.yaml"
)

// ServerConfig represents a single Memcached server configuration
type ServerConfig struct {
	Name    string `yaml:"name"`    // Human-readable server name
	Address string `yaml:"address"` // Server address in host:port format
	Default bool   `yaml:"default"` // Whether this is the default server
}

// ServersConfig holds all server configurations
type ServersConfig struct {
	Servers  []ServerConfig `yaml:"servers"`   // List of configured servers
	LastUsed string         `yaml:"last_used"` // Name of the last used server
}

// Validate checks if the server configuration is valid
func (s *ServerConfig) Validate() error {
	if s.Name == "" {
		return errors.New("server name cannot be empty")
	}
	if s.Address == "" {
		return errors.New("server address cannot be empty")
	}

	// Validate address format (host:port)
	host, port, err := net.SplitHostPort(s.Address)
	if err != nil {
		return fmt.Errorf("invalid address format: %w", err)
	}
	if host == "" || port == "" {
		return errors.New("invalid address format: expected host:port")
	}

	return nil
}

// DefaultServersConfig returns a ServersConfig with sensible default values
func DefaultServersConfig() *ServersConfig {
	return &ServersConfig{
		Servers: []ServerConfig{
			{
				Name:    "localhost",
				Address: "localhost:11211",
				Default: true,
			},
		},
		LastUsed: "",
	}
}

// ServersFilePath returns the full path to the servers configuration file
func ServersFilePath() string {
	return filepath.Join(ConfigDir(), ServersFileName)
}

// LoadServers reads the servers configuration from the config file.
// If the file doesn't exist, it returns the default configuration.
func LoadServers() (*ServersConfig, error) {
	path := ServersFilePath()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file exists, return defaults
			return DefaultServersConfig(), nil
		}
		return nil, fmt.Errorf("failed to read servers config file: %w", err)
	}

	var cfg ServersConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse servers config file: %w", err)
	}

	return &cfg, nil
}

// SaveServers writes the servers configuration to the config file.
// It creates the config directory if it doesn't exist.
func SaveServers(cfg *ServersConfig) error {
	if cfg == nil {
		return errors.New("cannot save nil config")
	}

	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal servers config: %w", err)
	}

	path := ServersFilePath()
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write servers config file: %w", err)
	}

	return nil
}

// AddServer adds a new server to the configuration.
// Returns an error if a server with the same name already exists.
func AddServer(name, address string) error {
	// Validate inputs
	if name == "" {
		return errors.New("server name cannot be empty")
	}
	if address == "" {
		return errors.New("server address cannot be empty")
	}

	// Validate address format
	serverCfg := ServerConfig{Name: name, Address: address}
	if err := serverCfg.Validate(); err != nil {
		return err
	}

	cfg, err := LoadServers()
	if err != nil {
		return fmt.Errorf("failed to load servers config: %w", err)
	}

	// Check for duplicate name
	for _, s := range cfg.Servers {
		if s.Name == name {
			return fmt.Errorf("server with name %q already exists", name)
		}
	}

	// Add the new server
	cfg.Servers = append(cfg.Servers, ServerConfig{
		Name:    name,
		Address: address,
		Default: false,
	})

	return SaveServers(cfg)
}

// RemoveServer removes a server from the configuration by name.
// Returns an error if the server doesn't exist or if it's the last server.
func RemoveServer(name string) error {
	cfg, err := LoadServers()
	if err != nil {
		return fmt.Errorf("failed to load servers config: %w", err)
	}

	// Check that we have more than one server
	if len(cfg.Servers) <= 1 {
		return errors.New("cannot remove the last server")
	}

	// Find and remove the server
	found := false
	newServers := make([]ServerConfig, 0, len(cfg.Servers)-1)
	for _, s := range cfg.Servers {
		if s.Name == name {
			found = true
			continue
		}
		newServers = append(newServers, s)
	}

	if !found {
		return fmt.Errorf("server %q not found", name)
	}

	cfg.Servers = newServers

	// Update LastUsed if it was the removed server
	if cfg.LastUsed == name {
		// Set to first remaining server
		if len(cfg.Servers) > 0 {
			cfg.LastUsed = cfg.Servers[0].Name
		} else {
			cfg.LastUsed = ""
		}
	}

	return SaveServers(cfg)
}

// SetDefault sets a server as the default server.
// Only one server can be the default at a time.
func SetDefault(name string) error {
	cfg, err := LoadServers()
	if err != nil {
		return fmt.Errorf("failed to load servers config: %w", err)
	}

	// Find the server and update default status
	found := false
	for i := range cfg.Servers {
		if cfg.Servers[i].Name == name {
			found = true
			cfg.Servers[i].Default = true
		} else {
			cfg.Servers[i].Default = false
		}
	}

	if !found {
		return fmt.Errorf("server %q not found", name)
	}

	return SaveServers(cfg)
}

// SetLastUsed sets the last used server by name.
func SetLastUsed(name string) error {
	cfg, err := LoadServers()
	if err != nil {
		return fmt.Errorf("failed to load servers config: %w", err)
	}

	// Verify the server exists
	found := false
	for _, s := range cfg.Servers {
		if s.Name == name {
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("server %q not found", name)
	}

	cfg.LastUsed = name
	return SaveServers(cfg)
}

// GetServer returns a server configuration by name.
func GetServer(name string) (*ServerConfig, error) {
	cfg, err := LoadServers()
	if err != nil {
		return nil, fmt.Errorf("failed to load servers config: %w", err)
	}

	for _, s := range cfg.Servers {
		if s.Name == name {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("server %q not found", name)
}

// GetDefaultServer returns the default server configuration.
// If no server is marked as default, returns the first server.
func GetDefaultServer() (*ServerConfig, error) {
	cfg, err := LoadServers()
	if err != nil {
		return nil, fmt.Errorf("failed to load servers config: %w", err)
	}

	if len(cfg.Servers) == 0 {
		return nil, errors.New("no servers configured")
	}

	// Look for default server
	for _, s := range cfg.Servers {
		if s.Default {
			return &s, nil
		}
	}

	// Return first server if no default
	return &cfg.Servers[0], nil
}

// GetLastUsedServer returns the last used server configuration.
// If no last used server is set, returns the default server.
func GetLastUsedServer() (*ServerConfig, error) {
	cfg, err := LoadServers()
	if err != nil {
		return nil, fmt.Errorf("failed to load servers config: %w", err)
	}

	if len(cfg.Servers) == 0 {
		return nil, errors.New("no servers configured")
	}

	// If LastUsed is set, find that server
	if cfg.LastUsed != "" {
		for _, s := range cfg.Servers {
			if s.Name == cfg.LastUsed {
				return &s, nil
			}
		}
	}

	// Fall back to default server
	return GetDefaultServer()
}
