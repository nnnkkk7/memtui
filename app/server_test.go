package app

import (
	"context"
	"errors"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// MockConnector implements the ServerConnector interface for testing
type MockConnector struct {
	ConnectFunc    func(ctx context.Context, address string) error
	DisconnectFunc func() error
	PingFunc       func(ctx context.Context) error
	VersionFunc    func(ctx context.Context) (string, error)
	connected      bool
	address        string
}

func (m *MockConnector) Connect(ctx context.Context, address string) error {
	if m.ConnectFunc != nil {
		err := m.ConnectFunc(ctx, address)
		if err == nil {
			m.connected = true
			m.address = address
		}
		return err
	}
	m.connected = true
	m.address = address
	return nil
}

func (m *MockConnector) Disconnect() error {
	if m.DisconnectFunc != nil {
		err := m.DisconnectFunc()
		if err == nil {
			m.connected = false
		}
		return err
	}
	m.connected = false
	return nil
}

func (m *MockConnector) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	if m.connected {
		return nil
	}
	return errors.New("not connected")
}

func (m *MockConnector) Version(ctx context.Context) (string, error) {
	if m.VersionFunc != nil {
		return m.VersionFunc(ctx)
	}
	return "1.6.0", nil
}

func (m *MockConnector) IsConnected() bool {
	return m.connected
}

func (m *MockConnector) Address() string {
	return m.address
}

func TestSwitchServerMsg(t *testing.T) {
	t.Run("SwitchServerMsg contains address", func(t *testing.T) {
		msg := SwitchServerMsg{
			Address: "localhost:11211",
			Name:    "local",
		}

		if msg.Address != "localhost:11211" {
			t.Errorf("expected address 'localhost:11211', got %q", msg.Address)
		}
		if msg.Name != "local" {
			t.Errorf("expected name 'local', got %q", msg.Name)
		}
	})
}

func TestServerConnectedMsg(t *testing.T) {
	t.Run("ServerConnectedMsg contains connection info", func(t *testing.T) {
		msg := ServerConnectedMsg{
			Address: "prod.example.com:11211",
			Name:    "production",
			Version: "1.6.0",
		}

		if msg.Address != "prod.example.com:11211" {
			t.Errorf("expected address 'prod.example.com:11211', got %q", msg.Address)
		}
		if msg.Name != "production" {
			t.Errorf("expected name 'production', got %q", msg.Name)
		}
		if msg.Version != "1.6.0" {
			t.Errorf("expected version '1.6.0', got %q", msg.Version)
		}
	})
}

func TestServerErrorMsg(t *testing.T) {
	t.Run("ServerErrorMsg contains error info", func(t *testing.T) {
		err := errors.New("connection refused")
		msg := ServerErrorMsg{
			Address: "fail.example.com:11211",
			Name:    "failing",
			Err:     err,
		}

		if msg.Address != "fail.example.com:11211" {
			t.Errorf("expected address 'fail.example.com:11211', got %q", msg.Address)
		}
		if msg.Name != "failing" {
			t.Errorf("expected name 'failing', got %q", msg.Name)
		}
		if msg.Err != err {
			t.Errorf("expected error %v, got %v", err, msg.Err)
		}
	})

	t.Run("ServerErrorMsg Error method", func(t *testing.T) {
		err := errors.New("timeout")
		msg := ServerErrorMsg{
			Address: "slow.example.com:11211",
			Err:     err,
		}

		if msg.Error() != "timeout" {
			t.Errorf("expected 'timeout', got %q", msg.Error())
		}
	})

	t.Run("ServerErrorMsg Error method with nil error", func(t *testing.T) {
		msg := ServerErrorMsg{
			Address: "slow.example.com:11211",
			Err:     nil,
		}

		if msg.Error() != "" {
			t.Errorf("expected empty string, got %q", msg.Error())
		}
	})
}

func TestServerDisconnectedMsg(t *testing.T) {
	t.Run("ServerDisconnectedMsg contains address", func(t *testing.T) {
		msg := ServerDisconnectedMsg{
			Address: "old.example.com:11211",
			Name:    "old-server",
		}

		if msg.Address != "old.example.com:11211" {
			t.Errorf("expected address 'old.example.com:11211', got %q", msg.Address)
		}
		if msg.Name != "old-server" {
			t.Errorf("expected name 'old-server', got %q", msg.Name)
		}
	})
}

func TestSwitchServerCmd(t *testing.T) {
	t.Run("successful connection", func(t *testing.T) {
		connector := &MockConnector{
			VersionFunc: func(ctx context.Context) (string, error) {
				return "1.6.5", nil
			},
		}

		cmd := SwitchServerCmd(connector, "test", "localhost:11211", 5*time.Second)
		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		msg := cmd()
		connectedMsg, ok := msg.(ServerConnectedMsg)
		if !ok {
			t.Fatalf("expected ServerConnectedMsg, got %T", msg)
		}

		if connectedMsg.Address != "localhost:11211" {
			t.Errorf("expected address 'localhost:11211', got %q", connectedMsg.Address)
		}
		if connectedMsg.Name != "test" {
			t.Errorf("expected name 'test', got %q", connectedMsg.Name)
		}
		if connectedMsg.Version != "1.6.5" {
			t.Errorf("expected version '1.6.5', got %q", connectedMsg.Version)
		}
	})

	t.Run("connection error", func(t *testing.T) {
		connector := &MockConnector{
			ConnectFunc: func(ctx context.Context, address string) error {
				return errors.New("connection refused")
			},
		}

		cmd := SwitchServerCmd(connector, "failing", "fail:11211", 5*time.Second)
		msg := cmd()

		errorMsg, ok := msg.(ServerErrorMsg)
		if !ok {
			t.Fatalf("expected ServerErrorMsg, got %T", msg)
		}

		if errorMsg.Address != "fail:11211" {
			t.Errorf("expected address 'fail:11211', got %q", errorMsg.Address)
		}
		if errorMsg.Error() != "connection refused" {
			t.Errorf("expected 'connection refused', got %q", errorMsg.Error())
		}
	})

	t.Run("version error returns connection with unknown version", func(t *testing.T) {
		connector := &MockConnector{
			VersionFunc: func(ctx context.Context) (string, error) {
				return "", errors.New("version command failed")
			},
		}

		cmd := SwitchServerCmd(connector, "test", "localhost:11211", 5*time.Second)
		msg := cmd()

		// Should still succeed but with empty version
		connectedMsg, ok := msg.(ServerConnectedMsg)
		if !ok {
			t.Fatalf("expected ServerConnectedMsg, got %T", msg)
		}

		if connectedMsg.Version != "unknown" {
			t.Errorf("expected version 'unknown', got %q", connectedMsg.Version)
		}
	})

	t.Run("nil connector returns error", func(t *testing.T) {
		cmd := SwitchServerCmd(nil, "test", "localhost:11211", 5*time.Second)
		msg := cmd()

		errorMsg, ok := msg.(ServerErrorMsg)
		if !ok {
			t.Fatalf("expected ServerErrorMsg, got %T", msg)
		}

		if errorMsg.Err == nil {
			t.Error("expected non-nil error")
		}
	})
}

func TestDisconnectServerCmd(t *testing.T) {
	t.Run("successful disconnection", func(t *testing.T) {
		connector := &MockConnector{
			connected: true,
			address:   "localhost:11211",
		}

		cmd := DisconnectServerCmd(connector, "test", "localhost:11211")
		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		msg := cmd()
		disconnectedMsg, ok := msg.(ServerDisconnectedMsg)
		if !ok {
			t.Fatalf("expected ServerDisconnectedMsg, got %T", msg)
		}

		if disconnectedMsg.Address != "localhost:11211" {
			t.Errorf("expected address 'localhost:11211', got %q", disconnectedMsg.Address)
		}
		if disconnectedMsg.Name != "test" {
			t.Errorf("expected name 'test', got %q", disconnectedMsg.Name)
		}
	})

	t.Run("disconnection error", func(t *testing.T) {
		connector := &MockConnector{
			connected: true,
			DisconnectFunc: func() error {
				return errors.New("disconnect failed")
			},
		}

		cmd := DisconnectServerCmd(connector, "test", "localhost:11211")
		msg := cmd()

		errorMsg, ok := msg.(ServerErrorMsg)
		if !ok {
			t.Fatalf("expected ServerErrorMsg, got %T", msg)
		}

		if errorMsg.Error() != "disconnect failed" {
			t.Errorf("expected 'disconnect failed', got %q", errorMsg.Error())
		}
	})

	t.Run("nil connector returns error", func(t *testing.T) {
		cmd := DisconnectServerCmd(nil, "test", "localhost:11211")
		msg := cmd()

		errorMsg, ok := msg.(ServerErrorMsg)
		if !ok {
			t.Fatalf("expected ServerErrorMsg, got %T", msg)
		}

		if errorMsg.Err == nil {
			t.Error("expected non-nil error")
		}
	})
}

func TestHandleServerSwitch(t *testing.T) {
	t.Run("handles SwitchServerMsg", func(t *testing.T) {
		connector := &MockConnector{}
		msg := SwitchServerMsg{
			Address: "new.example.com:11211",
			Name:    "new-server",
		}

		cmd := HandleServerSwitch(msg, connector, 5*time.Second)
		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		// Execute the command
		result := cmd()
		_, ok := result.(ServerConnectedMsg)
		if !ok {
			t.Fatalf("expected ServerConnectedMsg, got %T", result)
		}
	})

	t.Run("returns nil for unrelated messages", func(t *testing.T) {
		connector := &MockConnector{}
		msg := tea.KeyMsg{Type: tea.KeyEnter}

		cmd := HandleServerSwitch(msg, connector, 5*time.Second)
		if cmd != nil {
			t.Errorf("expected nil for unrelated message, got %T", cmd)
		}
	})

	t.Run("handles ServerConnectedMsg", func(t *testing.T) {
		connector := &MockConnector{}
		msg := ServerConnectedMsg{
			Address: "connected.example.com:11211",
			Name:    "connected-server",
			Version: "1.6.0",
		}

		// This should return nil as it's a result message, not an action
		cmd := HandleServerSwitch(msg, connector, 5*time.Second)
		if cmd != nil {
			t.Error("expected nil for result message")
		}
	})
}

func TestServerConnectionState(t *testing.T) {
	t.Run("initial state", func(t *testing.T) {
		state := NewServerConnectionState()

		if state.IsConnected() {
			t.Error("expected not connected initially")
		}
		if state.CurrentAddress() != "" {
			t.Errorf("expected empty address, got %q", state.CurrentAddress())
		}
		if state.CurrentName() != "" {
			t.Errorf("expected empty name, got %q", state.CurrentName())
		}
	})

	t.Run("set connected", func(t *testing.T) {
		state := NewServerConnectionState()

		state.SetConnected("test", "localhost:11211", "1.6.0")

		if !state.IsConnected() {
			t.Error("expected connected")
		}
		if state.CurrentAddress() != "localhost:11211" {
			t.Errorf("expected 'localhost:11211', got %q", state.CurrentAddress())
		}
		if state.CurrentName() != "test" {
			t.Errorf("expected 'test', got %q", state.CurrentName())
		}
		if state.CurrentVersion() != "1.6.0" {
			t.Errorf("expected '1.6.0', got %q", state.CurrentVersion())
		}
	})

	t.Run("set disconnected", func(t *testing.T) {
		state := NewServerConnectionState()
		state.SetConnected("test", "localhost:11211", "1.6.0")

		state.SetDisconnected()

		if state.IsConnected() {
			t.Error("expected not connected after disconnect")
		}
	})

	t.Run("set error", func(t *testing.T) {
		state := NewServerConnectionState()
		state.SetConnected("test", "localhost:11211", "1.6.0")

		testErr := errors.New("connection lost")
		state.SetError(testErr)

		if state.IsConnected() {
			t.Error("expected not connected after error")
		}
		if state.LastError() != testErr {
			t.Errorf("expected error %v, got %v", testErr, state.LastError())
		}
	})

	t.Run("clear error", func(t *testing.T) {
		state := NewServerConnectionState()
		state.SetError(errors.New("some error"))

		state.ClearError()

		if state.LastError() != nil {
			t.Error("expected nil error after clear")
		}
	})
}

func TestServerSwitchingState(t *testing.T) {
	t.Run("initial state is not switching", func(t *testing.T) {
		state := NewServerConnectionState()

		if state.IsSwitching() {
			t.Error("expected not switching initially")
		}
	})

	t.Run("set switching", func(t *testing.T) {
		state := NewServerConnectionState()

		state.SetSwitching("new-server", "new:11211")

		if !state.IsSwitching() {
			t.Error("expected switching")
		}
		if state.SwitchingTo() != "new:11211" {
			t.Errorf("expected 'new:11211', got %q", state.SwitchingTo())
		}
	})

	t.Run("switching completes on connect", func(t *testing.T) {
		state := NewServerConnectionState()
		state.SetSwitching("new-server", "new:11211")

		state.SetConnected("new-server", "new:11211", "1.6.0")

		if state.IsSwitching() {
			t.Error("expected not switching after connection")
		}
	})

	t.Run("switching completes on error", func(t *testing.T) {
		state := NewServerConnectionState()
		state.SetSwitching("new-server", "new:11211")

		state.SetError(errors.New("connection failed"))

		if state.IsSwitching() {
			t.Error("expected not switching after error")
		}
	})
}

func TestPingServerCmd(t *testing.T) {
	t.Run("successful ping", func(t *testing.T) {
		connector := &MockConnector{
			connected: true,
		}

		cmd := PingServerCmd(connector, "test", "localhost:11211", 5*time.Second)
		if cmd == nil {
			t.Fatal("expected command, got nil")
		}

		msg := cmd()
		pingMsg, ok := msg.(ServerPingResultMsg)
		if !ok {
			t.Fatalf("expected ServerPingResultMsg, got %T", msg)
		}

		if !pingMsg.Success {
			t.Error("expected successful ping")
		}
		if pingMsg.Address != "localhost:11211" {
			t.Errorf("expected 'localhost:11211', got %q", pingMsg.Address)
		}
	})

	t.Run("failed ping", func(t *testing.T) {
		connector := &MockConnector{
			PingFunc: func(ctx context.Context) error {
				return errors.New("ping timeout")
			},
		}

		cmd := PingServerCmd(connector, "test", "localhost:11211", 5*time.Second)
		msg := cmd()

		pingMsg, ok := msg.(ServerPingResultMsg)
		if !ok {
			t.Fatalf("expected ServerPingResultMsg, got %T", msg)
		}

		if pingMsg.Success {
			t.Error("expected failed ping")
		}
		if pingMsg.Err == nil {
			t.Error("expected non-nil error")
		}
	})

	t.Run("nil connector returns error", func(t *testing.T) {
		cmd := PingServerCmd(nil, "test", "localhost:11211", 5*time.Second)
		msg := cmd()

		pingMsg, ok := msg.(ServerPingResultMsg)
		if !ok {
			t.Fatalf("expected ServerPingResultMsg, got %T", msg)
		}

		if pingMsg.Success {
			t.Error("expected failed ping with nil connector")
		}
	})
}

func TestServerPingResultMsg(t *testing.T) {
	t.Run("successful ping result", func(t *testing.T) {
		msg := ServerPingResultMsg{
			Address: "localhost:11211",
			Name:    "test",
			Success: true,
			Latency: 5 * time.Millisecond,
		}

		if !msg.Success {
			t.Error("expected Success to be true")
		}
		if msg.Latency != 5*time.Millisecond {
			t.Errorf("expected 5ms latency, got %v", msg.Latency)
		}
	})

	t.Run("failed ping result", func(t *testing.T) {
		err := errors.New("timeout")
		msg := ServerPingResultMsg{
			Address: "localhost:11211",
			Name:    "test",
			Success: false,
			Err:     err,
		}

		if msg.Success {
			t.Error("expected Success to be false")
		}
		if msg.Err != err {
			t.Errorf("expected error %v, got %v", err, msg.Err)
		}
	})
}
