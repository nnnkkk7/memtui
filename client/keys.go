package client

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/nnnkkk7/memtui/models"
)

// memcachedEnd is the terminator for Memcached responses
const memcachedEnd = "END"

// KeyEnumerator enumerates keys from Memcached using lru_crawler metadump
type KeyEnumerator struct {
	addr    string
	timeout time.Duration
}

// NewKeyEnumerator creates a new key enumerator
func NewKeyEnumerator(addr string) *KeyEnumerator {
	return &KeyEnumerator{
		addr:    addr,
		timeout: 30 * time.Second,
	}
}

// WithTimeout sets the timeout for enumeration
func (e *KeyEnumerator) WithTimeout(d time.Duration) *KeyEnumerator {
	e.timeout = d
	return e
}

// EnumerateAll collects all keys and returns them as a slice
func (e *KeyEnumerator) EnumerateAll(ctx context.Context) ([]models.KeyInfo, error) {
	keyChan, errChan := e.EnumerateStream(ctx)

	var keys []models.KeyInfo
	for {
		select {
		case ki, ok := <-keyChan:
			if !ok {
				return keys, nil
			}
			keys = append(keys, ki)
		case err := <-errChan:
			if err != nil {
				return keys, err
			}
		case <-ctx.Done():
			return keys, ctx.Err()
		}
	}
}

// EnumerateStream returns channels for streaming key enumeration
func (e *KeyEnumerator) EnumerateStream(ctx context.Context) (<-chan models.KeyInfo, <-chan error) {
	keyChan := make(chan models.KeyInfo, 100)
	errChan := make(chan error, 1)

	go func() {
		defer close(keyChan)
		defer close(errChan)

		if err := e.enumerate(ctx, keyChan); err != nil {
			select {
			case errChan <- err:
			default:
			}
		}
	}()

	return keyChan, errChan
}

// enumerate performs the actual enumeration
func (e *KeyEnumerator) enumerate(ctx context.Context, keyChan chan<- models.KeyInfo) error {
	conn, err := net.DialTimeout("tcp", e.addr, e.timeout)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Set deadline from context or timeout
	deadline := time.Now().Add(e.timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	_ = conn.SetDeadline(deadline)

	// Send lru_crawler metadump command
	_, err = fmt.Fprintf(conn, "lru_crawler metadump all\r\n")
	if err != nil {
		return fmt.Errorf("failed to send command: %w", err)
	}

	// Read and parse response
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		// Check for cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := scanner.Text()
		// Handle \r from Memcached protocol
		line = strings.TrimSuffix(line, "\r")

		if line == memcachedEnd {
			break
		}

		if strings.HasPrefix(line, "ERROR") {
			return fmt.Errorf("server error: %s", line)
		}

		ki, err := models.ParseMetadumpLine(line)
		if err != nil {
			// Skip invalid lines
			continue
		}

		select {
		case keyChan <- ki:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	return nil
}
