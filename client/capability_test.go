package client_test

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/nnnkkk7/memtui/client"
)

// mockMemcachedServer creates a mock Memcached server for testing
type mockMemcachedServer struct {
	listener net.Listener
	version  string
	closed   bool
}

func newMockMemcachedServer(t *testing.T, version string) *mockMemcachedServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start mock server: %v", err)
	}

	server := &mockMemcachedServer{
		listener: listener,
		version:  version,
	}

	go server.serve(t)

	return server
}

func (s *mockMemcachedServer) serve(t *testing.T) {
	for {
		if s.closed {
			return
		}
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed {
				return
			}
			t.Logf("accept error: %v", err)
			continue
		}
		go s.handleConn(conn)
	}
}

func (s *mockMemcachedServer) handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)

	for scanner.Scan() {
		cmd := scanner.Text()
		switch {
		case cmd == "stats":
			// Return version in stats response
			fmt.Fprintf(conn, "STAT version %s\r\n", s.version)
			fmt.Fprintf(conn, "STAT pid 12345\r\n")
			fmt.Fprintf(conn, "STAT uptime 86400\r\n")
			fmt.Fprintf(conn, "END\r\n")
		case cmd == "version":
			fmt.Fprintf(conn, "VERSION %s\r\n", s.version)
		case strings.HasPrefix(cmd, "lru_crawler metadump"):
			if s.version >= "1.4.31" {
				fmt.Fprintf(conn, "key=test exp=0 la=123 cas=456 fetch=yes cls=1 size=100\r\n")
				fmt.Fprintf(conn, "END\r\n")
			} else {
				fmt.Fprintf(conn, "ERROR\r\n")
			}
		default:
			fmt.Fprintf(conn, "ERROR\r\n")
		}
	}
}

func (s *mockMemcachedServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *mockMemcachedServer) Close() {
	s.closed = true
	s.listener.Close()
}

func TestCapabilityDetector_Detect_SupportedVersion(t *testing.T) {
	server := newMockMemcachedServer(t, "1.6.22")
	defer server.Close()

	detector := client.NewCapabilityDetector()
	cap, err := detector.Detect(server.Addr())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cap.Version != "1.6.22" {
		t.Errorf("expected version '1.6.22', got '%s'", cap.Version)
	}
	if !cap.SupportsMetadump {
		t.Error("expected SupportsMetadump to be true")
	}
}

func TestCapabilityDetector_Detect_MinimumVersion(t *testing.T) {
	server := newMockMemcachedServer(t, "1.4.31")
	defer server.Close()

	detector := client.NewCapabilityDetector()
	cap, err := detector.Detect(server.Addr())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cap.SupportsMetadump {
		t.Error("expected SupportsMetadump to be true for 1.4.31")
	}
}

func TestCapabilityDetector_Detect_UnsupportedVersion(t *testing.T) {
	server := newMockMemcachedServer(t, "1.4.30")
	defer server.Close()

	detector := client.NewCapabilityDetector()
	cap, err := detector.Detect(server.Addr())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cap.SupportsMetadump {
		t.Error("expected SupportsMetadump to be false for 1.4.30")
	}
}

func TestCapabilityDetector_Verify_Success(t *testing.T) {
	server := newMockMemcachedServer(t, "1.6.22")
	defer server.Close()

	detector := client.NewCapabilityDetector()
	cap, err := detector.Verify(server.Addr())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cap == nil {
		t.Fatal("expected non-nil capability")
	}
}

func TestCapabilityDetector_Verify_UnsupportedVersion(t *testing.T) {
	server := newMockMemcachedServer(t, "1.4.24")
	defer server.Close()

	detector := client.NewCapabilityDetector()
	_, err := detector.Verify(server.Addr())
	if err == nil {
		t.Fatal("expected error for unsupported version")
	}
	if !errors.Is(err, client.ErrUnsupportedVersion) {
		t.Errorf("expected ErrUnsupportedVersion, got: %v", err)
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		version string
		major   int
		minor   int
		patch   int
	}{
		{"1.6.22", 1, 6, 22},
		{"1.4.31", 1, 4, 31},
		{"1.4.30", 1, 4, 30},
		{"1.5.0", 1, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			major, minor, patch, err := client.ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if major != tt.major || minor != tt.minor || patch != tt.patch {
				t.Errorf("expected %d.%d.%d, got %d.%d.%d",
					tt.major, tt.minor, tt.patch, major, minor, patch)
			}
		})
	}
}

func TestIsVersionSupported(t *testing.T) {
	tests := []struct {
		version   string
		supported bool
	}{
		{"1.6.22", true},
		{"1.4.31", true},
		{"1.5.0", true},
		{"1.4.30", false},
		{"1.4.24", false},
		{"1.3.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			result := client.IsVersionSupported(tt.version)
			if result != tt.supported {
				t.Errorf("IsVersionSupported(%s) = %v, want %v", tt.version, result, tt.supported)
			}
		})
	}
}
