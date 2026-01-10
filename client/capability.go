// Package client provides Memcached client functionality including connection handling,
// key enumeration, CAS operations, and server capability detection.
package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

// ErrUnsupportedVersion is returned when the Memcached version is too old
var ErrUnsupportedVersion = errors.New(
	"memcached version 1.4.31 or later is required for lru_crawler metadump support")

// MinRequiredVersion is the minimum supported Memcached version
const MinRequiredVersion = "1.4.31"

// Capability represents the detected server capabilities
type Capability struct {
	Version          string
	SupportsMetadump bool
}

// CapabilityDetector detects server capabilities
type CapabilityDetector struct {
	timeout time.Duration
}

// NewCapabilityDetector creates a new capability detector
func NewCapabilityDetector() *CapabilityDetector {
	return &CapabilityDetector{
		timeout: 5 * time.Second,
	}
}

// Detect connects to the server and detects its capabilities
func (d *CapabilityDetector) Detect(addr string) (*Capability, error) {
	conn, err := net.DialTimeout("tcp", addr, d.timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	_ = conn.SetDeadline(time.Now().Add(d.timeout))

	// Send stats command
	_, err = fmt.Fprintf(conn, "stats\r\n")
	if err != nil {
		return nil, fmt.Errorf("failed to send stats command: %w", err)
	}

	// Read response
	scanner := bufio.NewScanner(conn)
	var version string

	for scanner.Scan() {
		line := strings.TrimSuffix(scanner.Text(), "\r")
		if line == "END" {
			break
		}

		parts := strings.Fields(line)
		if len(parts) >= 3 && parts[0] == "STAT" && parts[1] == "version" {
			version = parts[2]
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read stats: %w", err)
	}

	if version == "" {
		return nil, errors.New("failed to detect server version")
	}

	return &Capability{
		Version:          version,
		SupportsMetadump: IsVersionSupported(version),
	}, nil
}

// Verify checks if the server meets the minimum requirements
func (d *CapabilityDetector) Verify(addr string) (*Capability, error) {
	caps, err := d.Detect(addr)
	if err != nil {
		return nil, err
	}

	if !caps.SupportsMetadump {
		return nil, fmt.Errorf("%w (detected: %s)", ErrUnsupportedVersion, caps.Version)
	}

	return caps, nil
}

// ParseVersion parses a version string like "1.6.22" into major, minor, patch
func ParseVersion(version string) (major, minor, patch int, err error) {
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", version)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %w", err)
	}

	minor, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid minor version: %w", err)
	}

	if len(parts) >= 3 {
		// Handle versions like "1.4.31-beta"
		patchStr := strings.Split(parts[2], "-")[0]
		patch, err = strconv.Atoi(patchStr)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid patch version: %w", err)
		}
	}

	return major, minor, patch, nil
}

// IsVersionSupported checks if the version meets minimum requirements (1.4.31+)
func IsVersionSupported(version string) bool {
	major, minor, patch, err := ParseVersion(version)
	if err != nil {
		return false
	}

	// Compare with minimum required version 1.4.31
	if major > 1 {
		return true
	}
	if major < 1 {
		return false
	}
	// major == 1
	if minor > 4 {
		return true
	}
	if minor < 4 {
		return false
	}
	// minor == 4
	return patch >= 31
}
