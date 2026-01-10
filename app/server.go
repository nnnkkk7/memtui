// Package app provides server switching functionality for memtui.
// This file contains messages, commands, and state management for
// connecting to and switching between multiple Memcached servers.
package app

import (
	"context"
	"errors"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// ServerConnector defines the interface for connecting to Memcached servers.
// This interface allows for mocking in tests and decouples the server switching
// logic from the actual client implementation.
type ServerConnector interface {
	// Connect establishes a connection to the specified server address
	Connect(ctx context.Context, address string) error

	// Disconnect closes the current connection
	Disconnect() error

	// Ping checks if the current connection is alive
	Ping(ctx context.Context) error

	// Version returns the Memcached server version
	Version(ctx context.Context) (string, error)

	// IsConnected returns true if currently connected
	IsConnected() bool

	// Address returns the current server address
	Address() string
}

// Messages

// SwitchServerMsg is sent to initiate a server switch
type SwitchServerMsg struct {
	Address string // Server address to connect to
	Name    string // Human-readable server name
}

// ServerConnectedMsg is sent when a server connection is established successfully
type ServerConnectedMsg struct {
	Address string // Server address
	Name    string // Human-readable server name
	Version string // Memcached server version
}

// ServerErrorMsg is sent when a server operation fails
type ServerErrorMsg struct {
	Address string // Server address that was being accessed
	Name    string // Human-readable server name
	Err     error  // The error that occurred
}

// Error returns the error message
func (m ServerErrorMsg) Error() string {
	if m.Err == nil {
		return ""
	}
	return m.Err.Error()
}

// ServerDisconnectedMsg is sent when a server is disconnected
type ServerDisconnectedMsg struct {
	Address string // Server address that was disconnected
	Name    string // Human-readable server name
}

// ServerPingResultMsg is sent with the result of a server ping operation
type ServerPingResultMsg struct {
	Address string        // Server address
	Name    string        // Human-readable server name
	Success bool          // Whether the ping was successful
	Latency time.Duration // Round-trip latency
	Err     error         // Error if ping failed
}

// Commands

// SwitchServerCmd creates a command that switches to a different server.
// It connects to the new server and returns the appropriate message.
func SwitchServerCmd(connector ServerConnector, name, address string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		if connector == nil {
			return ServerErrorMsg{
				Address: address,
				Name:    name,
				Err:     errors.New("connector is nil"),
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Connect to the new server
		if err := connector.Connect(ctx, address); err != nil {
			return ServerErrorMsg{
				Address: address,
				Name:    name,
				Err:     err,
			}
		}

		// Get server version
		version, err := connector.Version(ctx)
		if err != nil {
			// Connection succeeded but couldn't get version
			// This is not a critical error
			version = "unknown"
		}

		return ServerConnectedMsg{
			Address: address,
			Name:    name,
			Version: version,
		}
	}
}

// DisconnectServerCmd creates a command that disconnects from the current server.
func DisconnectServerCmd(connector ServerConnector, name, address string) tea.Cmd {
	return func() tea.Msg {
		if connector == nil {
			return ServerErrorMsg{
				Address: address,
				Name:    name,
				Err:     errors.New("connector is nil"),
			}
		}

		if err := connector.Disconnect(); err != nil {
			return ServerErrorMsg{
				Address: address,
				Name:    name,
				Err:     err,
			}
		}

		return ServerDisconnectedMsg{
			Address: address,
			Name:    name,
		}
	}
}

// PingServerCmd creates a command that pings the current server to check connectivity.
func PingServerCmd(connector ServerConnector, name, address string, timeout time.Duration) tea.Cmd {
	return func() tea.Msg {
		if connector == nil {
			return ServerPingResultMsg{
				Address: address,
				Name:    name,
				Success: false,
				Err:     errors.New("connector is nil"),
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		start := time.Now()
		err := connector.Ping(ctx)
		latency := time.Since(start)

		if err != nil {
			return ServerPingResultMsg{
				Address: address,
				Name:    name,
				Success: false,
				Latency: latency,
				Err:     err,
			}
		}

		return ServerPingResultMsg{
			Address: address,
			Name:    name,
			Success: true,
			Latency: latency,
		}
	}
}

// HandleServerSwitch processes server-related messages and returns the appropriate command.
// This function is designed to be called from the main application Update function.
func HandleServerSwitch(msg tea.Msg, connector ServerConnector, timeout time.Duration) tea.Cmd {
	switch m := msg.(type) {
	case SwitchServerMsg:
		return SwitchServerCmd(connector, m.Name, m.Address, timeout)
	default:
		return nil
	}
}

// ServerConnectionState tracks the current server connection state.
// This is thread-safe and can be used to track connection status from
// multiple goroutines if needed.
type ServerConnectionState struct {
	mu            sync.RWMutex
	connected     bool
	name          string
	address       string
	version       string
	lastError     error
	switching     bool
	switchingName string
	switchingAddr string
}

// NewServerConnectionState creates a new connection state tracker
func NewServerConnectionState() *ServerConnectionState {
	return &ServerConnectionState{}
}

// IsConnected returns true if currently connected to a server
func (s *ServerConnectionState) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// CurrentAddress returns the address of the currently connected server
func (s *ServerConnectionState) CurrentAddress() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.address
}

// CurrentName returns the name of the currently connected server
func (s *ServerConnectionState) CurrentName() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.name
}

// CurrentVersion returns the version of the currently connected server
func (s *ServerConnectionState) CurrentVersion() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.version
}

// LastError returns the last error that occurred
func (s *ServerConnectionState) LastError() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastError
}

// IsSwitching returns true if currently switching servers
func (s *ServerConnectionState) IsSwitching() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.switching
}

// SwitchingTo returns the address being switched to
func (s *ServerConnectionState) SwitchingTo() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.switchingAddr
}

// SetConnected updates the state to indicate a successful connection
func (s *ServerConnectionState) SetConnected(name, address, version string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = true
	s.name = name
	s.address = address
	s.version = version
	s.lastError = nil
	s.switching = false
	s.switchingName = ""
	s.switchingAddr = ""
}

// SetDisconnected updates the state to indicate disconnection
func (s *ServerConnectionState) SetDisconnected() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
}

// SetError updates the state to indicate an error
func (s *ServerConnectionState) SetError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.connected = false
	s.lastError = err
	s.switching = false
	s.switchingName = ""
	s.switchingAddr = ""
}

// ClearError clears the last error
func (s *ServerConnectionState) ClearError() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lastError = nil
}

// SetSwitching updates the state to indicate a server switch is in progress
func (s *ServerConnectionState) SetSwitching(name, address string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.switching = true
	s.switchingName = name
	s.switchingAddr = address
}

// ConnectionInfo returns a snapshot of the current connection information
type ConnectionInfo struct {
	Connected bool
	Name      string
	Address   string
	Version   string
	Switching bool
	Error     error
}

// Info returns a snapshot of the current connection state
func (s *ServerConnectionState) Info() ConnectionInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ConnectionInfo{
		Connected: s.connected,
		Name:      s.name,
		Address:   s.address,
		Version:   s.version,
		Switching: s.switching,
		Error:     s.lastError,
	}
}
