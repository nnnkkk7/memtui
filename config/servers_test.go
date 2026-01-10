package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestServerConfig(t *testing.T) {
	t.Run("ServerConfig struct fields", func(t *testing.T) {
		cfg := ServerConfig{
			Name:    "production",
			Address: "prod.example.com:11211",
			Default: true,
		}

		if cfg.Name != "production" {
			t.Errorf("expected Name 'production', got %q", cfg.Name)
		}
		if cfg.Address != "prod.example.com:11211" {
			t.Errorf("expected Address 'prod.example.com:11211', got %q", cfg.Address)
		}
		if !cfg.Default {
			t.Error("expected Default to be true")
		}
	})
}

func TestServersConfig(t *testing.T) {
	t.Run("ServersConfig struct fields", func(t *testing.T) {
		cfg := ServersConfig{
			Servers: []ServerConfig{
				{Name: "local", Address: "localhost:11211", Default: true},
				{Name: "staging", Address: "staging:11211", Default: false},
			},
			LastUsed: "local",
		}

		if len(cfg.Servers) != 2 {
			t.Errorf("expected 2 servers, got %d", len(cfg.Servers))
		}
		if cfg.LastUsed != "local" {
			t.Errorf("expected LastUsed 'local', got %q", cfg.LastUsed)
		}
	})
}

func TestDefaultServersConfig(t *testing.T) {
	cfg := DefaultServersConfig()

	if cfg == nil {
		t.Fatal("DefaultServersConfig returned nil")
	}

	if len(cfg.Servers) != 1 {
		t.Errorf("expected 1 default server, got %d", len(cfg.Servers))
	}

	if len(cfg.Servers) > 0 {
		defaultServer := cfg.Servers[0]
		if defaultServer.Name != "localhost" {
			t.Errorf("expected default server name 'localhost', got %q", defaultServer.Name)
		}
		if defaultServer.Address != "localhost:11211" {
			t.Errorf("expected default address 'localhost:11211', got %q", defaultServer.Address)
		}
		if !defaultServer.Default {
			t.Error("expected default server to have Default=true")
		}
	}
}

func TestServersFilePath(t *testing.T) {
	path := ServersFilePath()

	if path == "" {
		t.Error("ServersFilePath returned empty string")
	}

	// Should contain "servers.yaml"
	if filepath.Base(path) != "servers.yaml" {
		t.Errorf("expected filename 'servers.yaml', got %q", filepath.Base(path))
	}

	// Should be in config directory
	dir := filepath.Dir(path)
	expectedDir := ConfigDir()
	if dir != expectedDir {
		t.Errorf("expected directory %q, got %q", expectedDir, dir)
	}
}

func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "memtui-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Set XDG_CONFIG_HOME to temp directory
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	os.Setenv("XDG_CONFIG_HOME", tmpDir)

	cleanup := func() {
		os.Setenv("XDG_CONFIG_HOME", oldXDG)
		os.RemoveAll(tmpDir)
	}

	return tmpDir, cleanup
}

func TestLoadServers(t *testing.T) {
	t.Run("returns default when file does not exist", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg, err := LoadServers()
		if err != nil {
			t.Fatalf("LoadServers failed: %v", err)
		}

		defaultCfg := DefaultServersConfig()
		if len(cfg.Servers) != len(defaultCfg.Servers) {
			t.Errorf("expected %d servers, got %d", len(defaultCfg.Servers), len(cfg.Servers))
		}
	})

	t.Run("loads existing file", func(t *testing.T) {
		tmpDir, cleanup := setupTestDir(t)
		defer cleanup()

		// Create config directory and file
		configDir := filepath.Join(tmpDir, AppName)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}

		content := `servers:
  - name: test-server
    address: test:11211
    default: true
  - name: another
    address: another:11211
    default: false
last_used: test-server
`
		path := filepath.Join(configDir, "servers.yaml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		cfg, err := LoadServers()
		if err != nil {
			t.Fatalf("LoadServers failed: %v", err)
		}

		if len(cfg.Servers) != 2 {
			t.Errorf("expected 2 servers, got %d", len(cfg.Servers))
		}
		if cfg.LastUsed != "test-server" {
			t.Errorf("expected LastUsed 'test-server', got %q", cfg.LastUsed)
		}
	})

	t.Run("returns error on invalid yaml", func(t *testing.T) {
		tmpDir, cleanup := setupTestDir(t)
		defer cleanup()

		// Create config directory and invalid file
		configDir := filepath.Join(tmpDir, AppName)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			t.Fatalf("failed to create config dir: %v", err)
		}

		content := `invalid: [yaml: content`
		path := filepath.Join(configDir, "servers.yaml")
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write test file: %v", err)
		}

		_, err := LoadServers()
		if err == nil {
			t.Error("expected error for invalid yaml, got nil")
		}
	})
}

func TestSaveServers(t *testing.T) {
	t.Run("creates config directory and file", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "test", Address: "test:11211", Default: true},
			},
			LastUsed: "test",
		}

		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("SaveServers failed: %v", err)
		}

		// Verify file was created
		path := ServersFilePath()
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Error("servers.yaml was not created")
		}

		// Verify content by loading
		loaded, err := LoadServers()
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}

		if len(loaded.Servers) != 1 {
			t.Errorf("expected 1 server, got %d", len(loaded.Servers))
		}
		if loaded.LastUsed != "test" {
			t.Errorf("expected LastUsed 'test', got %q", loaded.LastUsed)
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		err := SaveServers(nil)
		if err == nil {
			t.Error("expected error for nil config, got nil")
		}
	})
}

func TestAddServer(t *testing.T) {
	t.Run("adds new server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		// Start with default config
		cfg := DefaultServersConfig()
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = AddServer("production", "prod.example.com:11211")
		if err != nil {
			t.Fatalf("AddServer failed: %v", err)
		}

		loaded, err := LoadServers()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if len(loaded.Servers) != 2 {
			t.Errorf("expected 2 servers, got %d", len(loaded.Servers))
		}

		// Find the added server
		found := false
		for _, s := range loaded.Servers {
			if s.Name == "production" && s.Address == "prod.example.com:11211" {
				found = true
				break
			}
		}
		if !found {
			t.Error("added server not found in config")
		}
	})

	t.Run("returns error for duplicate name", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := DefaultServersConfig()
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = AddServer("localhost", "other:11211")
		if err == nil {
			t.Error("expected error for duplicate name, got nil")
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		err := AddServer("", "localhost:11211")
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("returns error for empty address", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		err := AddServer("test", "")
		if err == nil {
			t.Error("expected error for empty address, got nil")
		}
	})

	t.Run("returns error for invalid address format", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		err := AddServer("test", "invalid-address")
		if err == nil {
			t.Error("expected error for invalid address format, got nil")
		}
	})
}

func TestRemoveServer(t *testing.T) {
	t.Run("removes existing server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: false},
				{Name: "server2", Address: "server2:11211", Default: true},
			},
			LastUsed: "server1",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = RemoveServer("server1")
		if err != nil {
			t.Fatalf("RemoveServer failed: %v", err)
		}

		loaded, err := LoadServers()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if len(loaded.Servers) != 1 {
			t.Errorf("expected 1 server, got %d", len(loaded.Servers))
		}

		for _, s := range loaded.Servers {
			if s.Name == "server1" {
				t.Error("server1 should have been removed")
			}
		}
	})

	t.Run("returns error for non-existent server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := DefaultServersConfig()
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = RemoveServer("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent server, got nil")
		}
	})

	t.Run("returns error when removing last server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "only-server", Address: "only:11211", Default: true},
			},
			LastUsed: "only-server",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = RemoveServer("only-server")
		if err == nil {
			t.Error("expected error when removing last server, got nil")
		}
	})

	t.Run("updates LastUsed when removing current server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: false},
				{Name: "server2", Address: "server2:11211", Default: true},
			},
			LastUsed: "server1",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = RemoveServer("server1")
		if err != nil {
			t.Fatalf("RemoveServer failed: %v", err)
		}

		loaded, err := LoadServers()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		// LastUsed should be updated to another server
		if loaded.LastUsed == "server1" {
			t.Error("LastUsed should not be the removed server")
		}
	})
}

func TestSetDefault(t *testing.T) {
	t.Run("sets server as default", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: true},
				{Name: "server2", Address: "server2:11211", Default: false},
			},
			LastUsed: "server1",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = SetDefault("server2")
		if err != nil {
			t.Fatalf("SetDefault failed: %v", err)
		}

		loaded, err := LoadServers()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		for _, s := range loaded.Servers {
			if s.Name == "server1" && s.Default {
				t.Error("server1 should no longer be default")
			}
			if s.Name == "server2" && !s.Default {
				t.Error("server2 should now be default")
			}
		}
	})

	t.Run("returns error for non-existent server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := DefaultServersConfig()
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = SetDefault("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent server, got nil")
		}
	})
}

func TestSetLastUsed(t *testing.T) {
	t.Run("sets last used server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: true},
				{Name: "server2", Address: "server2:11211", Default: false},
			},
			LastUsed: "server1",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = SetLastUsed("server2")
		if err != nil {
			t.Fatalf("SetLastUsed failed: %v", err)
		}

		loaded, err := LoadServers()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if loaded.LastUsed != "server2" {
			t.Errorf("expected LastUsed 'server2', got %q", loaded.LastUsed)
		}
	})

	t.Run("returns error for non-existent server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := DefaultServersConfig()
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		err = SetLastUsed("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent server, got nil")
		}
	})
}

func TestGetServer(t *testing.T) {
	t.Run("returns server by name", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: true},
				{Name: "server2", Address: "server2:11211", Default: false},
			},
			LastUsed: "server1",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		server, err := GetServer("server2")
		if err != nil {
			t.Fatalf("GetServer failed: %v", err)
		}

		if server.Name != "server2" {
			t.Errorf("expected name 'server2', got %q", server.Name)
		}
		if server.Address != "server2:11211" {
			t.Errorf("expected address 'server2:11211', got %q", server.Address)
		}
	})

	t.Run("returns error for non-existent server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := DefaultServersConfig()
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		_, err = GetServer("nonexistent")
		if err == nil {
			t.Error("expected error for non-existent server, got nil")
		}
	})
}

func TestGetDefaultServer(t *testing.T) {
	t.Run("returns default server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: false},
				{Name: "server2", Address: "server2:11211", Default: true},
			},
			LastUsed: "server1",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		server, err := GetDefaultServer()
		if err != nil {
			t.Fatalf("GetDefaultServer failed: %v", err)
		}

		if server.Name != "server2" {
			t.Errorf("expected default server 'server2', got %q", server.Name)
		}
	})

	t.Run("returns first server when no default set", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: false},
				{Name: "server2", Address: "server2:11211", Default: false},
			},
			LastUsed: "",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		server, err := GetDefaultServer()
		if err != nil {
			t.Fatalf("GetDefaultServer failed: %v", err)
		}

		// Should return the first server
		if server.Name != "server1" {
			t.Errorf("expected first server 'server1', got %q", server.Name)
		}
	})
}

func TestGetLastUsedServer(t *testing.T) {
	t.Run("returns last used server", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: true},
				{Name: "server2", Address: "server2:11211", Default: false},
			},
			LastUsed: "server2",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		server, err := GetLastUsedServer()
		if err != nil {
			t.Fatalf("GetLastUsedServer failed: %v", err)
		}

		if server.Name != "server2" {
			t.Errorf("expected last used server 'server2', got %q", server.Name)
		}
	})

	t.Run("returns default when no last used", func(t *testing.T) {
		_, cleanup := setupTestDir(t)
		defer cleanup()

		cfg := &ServersConfig{
			Servers: []ServerConfig{
				{Name: "server1", Address: "server1:11211", Default: true},
				{Name: "server2", Address: "server2:11211", Default: false},
			},
			LastUsed: "",
		}
		err := SaveServers(cfg)
		if err != nil {
			t.Fatalf("failed to save initial config: %v", err)
		}

		server, err := GetLastUsedServer()
		if err != nil {
			t.Fatalf("GetLastUsedServer failed: %v", err)
		}

		// Should return the default server
		if server.Name != "server1" {
			t.Errorf("expected default server 'server1', got %q", server.Name)
		}
	})
}

func TestServerConfigValidate(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := ServerConfig{
			Name:    "test",
			Address: "localhost:11211",
			Default: false,
		}

		err := cfg.Validate()
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		cfg := ServerConfig{
			Name:    "",
			Address: "localhost:11211",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for empty name, got nil")
		}
	})

	t.Run("empty address", func(t *testing.T) {
		cfg := ServerConfig{
			Name:    "test",
			Address: "",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for empty address, got nil")
		}
	})

	t.Run("invalid address format", func(t *testing.T) {
		cfg := ServerConfig{
			Name:    "test",
			Address: "invalid-address",
		}

		err := cfg.Validate()
		if err == nil {
			t.Error("expected error for invalid address format, got nil")
		}
	})
}
