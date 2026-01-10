package client_test

import (
	"context"
	"testing"
	"time"

	"github.com/nnnkkk7/memtui/client"
)

func TestNewClient(t *testing.T) {
	// Test with invalid address
	_, err := client.New("invalid:address:format")
	if err == nil {
		t.Error("expected error for invalid address format")
	}

	// Test with valid address format (connection will fail but format is OK)
	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Error("expected non-nil client")
	}
}

func TestClient_Options(t *testing.T) {
	c, err := client.New("localhost:11211",
		client.WithTimeout(5*time.Second),
		client.WithMaxIdleConns(10),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Error("expected non-nil client")
	}
}

func TestClient_Close(t *testing.T) {
	c, err := client.New("localhost:11211")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = c.Close()
	if err != nil {
		t.Errorf("unexpected error on close: %v", err)
	}
}

func TestClient_PingNoServer(t *testing.T) {
	// Use a port that's unlikely to have a server
	c, err := client.New("localhost:59999")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err = c.Ping(ctx)
	if err == nil {
		t.Error("expected error when pinging non-existent server")
	}
}

func TestClient_Address(t *testing.T) {
	addr := "localhost:11211"
	c, err := client.New(addr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer c.Close()

	if c.Address() != addr {
		t.Errorf("expected address '%s', got '%s'", addr, c.Address())
	}
}
