package client_test

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/models"
)

// mockMetadumpServer creates a mock server that responds to lru_crawler metadump
type mockMetadumpServer struct {
	listener net.Listener
	keys     []string
	closed   bool
}

func newMockMetadumpServer(t *testing.T, keys []string) *mockMetadumpServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}

	server := &mockMetadumpServer{
		listener: listener,
		keys:     keys,
	}

	go server.serve(t)

	return server
}

func (s *mockMetadumpServer) serve(t *testing.T) {
	for {
		if s.closed {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed {
				return
			}
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *mockMetadumpServer) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		cmd := scanner.Text()
		switch {
		case cmd == "stats":
			fmt.Fprintf(conn, "STAT version 1.6.22\r\n")
			fmt.Fprintf(conn, "END\r\n")
		case strings.HasPrefix(cmd, "lru_crawler metadump"):
			// Send keys
			for i, key := range s.keys {
				fmt.Fprintf(conn, "key=%s exp=0 la=%d cas=%d fetch=yes cls=1 size=100\r\n",
					key, time.Now().Unix()-int64(i), i+1)
			}
			fmt.Fprintf(conn, "END\r\n")
		default:
			fmt.Fprintf(conn, "ERROR\r\n")
		}
	}
}

func (s *mockMetadumpServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *mockMetadumpServer) Close() {
	s.closed = true
	s.listener.Close()
}

func TestKeyEnumerator_EnumerateAll(t *testing.T) {
	expectedKeys := []string{"user:1", "user:2", "session:abc", "cache:data"}
	server := newMockMetadumpServer(t, expectedKeys)
	defer server.Close()

	enum := client.NewKeyEnumerator(server.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keys, err := enum.EnumerateAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(keys))
	}

	// Verify all keys are present
	keySet := make(map[string]bool)
	for _, ki := range keys {
		keySet[ki.Key] = true
	}

	for _, expected := range expectedKeys {
		if !keySet[expected] {
			t.Errorf("missing key: %s", expected)
		}
	}
}

func TestKeyEnumerator_EnumerateStream(t *testing.T) {
	expectedKeys := []string{"key1", "key2", "key3"}
	server := newMockMetadumpServer(t, expectedKeys)
	defer server.Close()

	enum := client.NewKeyEnumerator(server.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keyChan, errChan := enum.EnumerateStream(ctx)

	var receivedKeys []models.KeyInfo
	done := false

	for !done {
		select {
		case ki, ok := <-keyChan:
			if !ok {
				done = true
				break
			}
			receivedKeys = append(receivedKeys, ki)
		case err := <-errChan:
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		case <-ctx.Done():
			t.Fatal("context timeout")
		}
	}

	if len(receivedKeys) != len(expectedKeys) {
		t.Errorf("expected %d keys, got %d", len(expectedKeys), len(receivedKeys))
	}
}

func TestKeyEnumerator_ContextCancel(t *testing.T) {
	// Create server with many keys to ensure cancellation happens during enumeration
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key:%d", i)
	}
	server := newMockMetadumpServer(t, keys)
	defer server.Close()

	enum := client.NewKeyEnumerator(server.Addr())
	ctx, cancel := context.WithCancel(context.Background())

	keyChan, errChan := enum.EnumerateStream(ctx)

	// Cancel after receiving a few keys
	received := 0
	for received < 10 {
		select {
		case _, ok := <-keyChan:
			if !ok {
				t.Fatal("channel closed before cancel")
			}
			received++
		case err := <-errChan:
			if err != nil {
				t.Fatalf("unexpected error before cancel: %v", err)
			}
		}
	}

	cancel()

	// Drain channel after cancel
	timeout := time.After(2 * time.Second)
	for {
		select {
		case _, ok := <-keyChan:
			if !ok {
				return // Channel closed - expected after cancel
			}
		case <-timeout:
			t.Fatal("timeout waiting for channel close after cancel")
		}
	}
}

func TestKeyEnumerator_LargeKeySet(t *testing.T) {
	// Test with 1000 keys
	keys := make([]string, 1000)
	for i := range keys {
		keys[i] = fmt.Sprintf("key:%d", i)
	}
	server := newMockMetadumpServer(t, keys)
	defer server.Close()

	enum := client.NewKeyEnumerator(server.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := enum.EnumerateAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1000 {
		t.Errorf("expected 1000 keys, got %d", len(result))
	}
}

func TestKeyEnumerator_EmptyResult(t *testing.T) {
	server := newMockMetadumpServer(t, []string{})
	defer server.Close()

	enum := client.NewKeyEnumerator(server.Addr())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	keys, err := enum.EnumerateAll(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(keys) != 0 {
		t.Errorf("expected 0 keys, got %d", len(keys))
	}
}
