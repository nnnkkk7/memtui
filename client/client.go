package client

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

// Client wraps gomemcache client with additional functionality
type Client struct {
	mc      *memcache.Client
	addr    string
	timeout time.Duration
}

// Option configures the client
type Option func(*Client)

// WithTimeout sets the connection and operation timeout
func WithTimeout(d time.Duration) Option {
	return func(c *Client) {
		c.timeout = d
	}
}

// WithMaxIdleConns sets the maximum number of idle connections
func WithMaxIdleConns(n int) Option {
	return func(c *Client) {
		c.mc.MaxIdleConns = n
	}
}

// New creates a new Memcached client
func New(addr string, opts ...Option) (*Client, error) {
	// Validate address format
	if !isValidAddress(addr) {
		return nil, errors.New("invalid address format: expected host:port")
	}

	mc := memcache.New(addr)

	c := &Client{
		mc:      mc,
		addr:    addr,
		timeout: 3 * time.Second, // default timeout
	}

	for _, opt := range opts {
		opt(c)
	}

	mc.Timeout = c.timeout

	return c, nil
}

// isValidAddress checks if the address is in host:port format
func isValidAddress(addr string) bool {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	return host != "" && port != ""
}

// Close closes the client connection
func (c *Client) Close() error {
	// gomemcache doesn't have a Close method, connections are pooled
	return nil
}

// Address returns the server address
func (c *Client) Address() string {
	return c.addr
}

// Ping checks if the server is reachable by executing a "version" command
func (c *Client) Ping(ctx context.Context) error {
	done := make(chan error, 1)

	go func() {
		// Use stats to verify connection
		_, err := c.Stats()
		done <- err
	}()

	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Get retrieves an item by key
func (c *Client) Get(key string) (*memcache.Item, error) {
	return c.mc.Get(key)
}

// Set stores an item
func (c *Client) Set(item *memcache.Item) error {
	return c.mc.Set(item)
}

// Delete removes an item by key
func (c *Client) Delete(key string) error {
	return c.mc.Delete(key)
}

// Stats returns server statistics
func (c *Client) Stats() (map[string]string, error) {
	items, err := c.mc.GetMulti([]string{"__stats_probe__"})
	if err != nil && !errors.Is(err, memcache.ErrCacheMiss) {
		return nil, err
	}
	_ = items

	// Actually run stats command - need raw connection
	stats := make(map[string]string)
	// For now, return empty stats as gomemcache doesn't expose raw stats
	// We'll implement this properly in capability.go using raw TCP
	return stats, nil
}

// RawCommand sends a raw command and returns the response lines
// This is used for commands not supported by gomemcache like lru_crawler metadump
func (c *Client) RawCommand(ctx context.Context, cmd string) ([]string, error) {
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set deadline from context
	if deadline, ok := ctx.Deadline(); ok {
		conn.SetDeadline(deadline)
	}

	// Send command
	_, err = conn.Write([]byte(cmd + "\r\n"))
	if err != nil {
		return nil, err
	}

	// Read response
	var lines []string
	buf := make([]byte, 4096)
	var data []byte

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		data = append(data, buf[:n]...)

		// Check if we've received END or ERROR
		str := string(data)
		if strings.Contains(str, "\r\nEND\r\n") || strings.Contains(str, "ERROR") {
			break
		}
	}

	// Parse lines
	rawLines := strings.Split(string(data), "\r\n")
	for _, line := range rawLines {
		line = strings.TrimSpace(line)
		if line != "" && line != "END" {
			lines = append(lines, line)
		}
	}

	return lines, nil
}
