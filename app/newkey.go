package app

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/bradfitz/gomemcache/memcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/ui/components/dialog"
)

// MaxKeyLength is the maximum length of a Memcached key (250 bytes)
const MaxKeyLength = 250

// Setter is an interface for setting keys in Memcached.
// This allows for easy mocking in tests.
type Setter interface {
	Set(item *memcache.Item) error
}

// NewKeyRequest holds the parameters for creating a new key.
type NewKeyRequest struct {
	Key   string // The key name
	Value string // The initial value
	TTL   int32  // Time to live in seconds (0 = no expiration)
	Flags uint32 // Optional flags for the item
}

// NewKeyMsg is sent when user requests to create a new key (triggers dialog).
type NewKeyMsg struct{}

// KeyCreatedMsg is sent when a key is successfully created.
type KeyCreatedMsg struct {
	Key string
}

// NewKeyErrorMsg is sent when key creation fails.
type NewKeyErrorMsg struct {
	Key string
	Err error
}

// NewKeyResult holds the result of a new key operation for UI handling.
type NewKeyResult struct {
	ShouldRefresh bool   // Whether the key list should be refreshed
	Error         string // Error message if creation failed
	CreatedKey    string // Key that was successfully created
	FailedKey     string // Key that failed to create
}

// NewKeyContext holds contextual information for a new key operation.
// Used to pass data between dialog confirmation and key creation.
type NewKeyContext struct {
	Key string
}

// NewKeyCmd creates a tea.Cmd that creates a new key in Memcached.
// Returns KeyCreatedMsg on success or NewKeyErrorMsg on failure.
func NewKeyCmd(client Setter, req NewKeyRequest) tea.Cmd {
	return func() tea.Msg {
		if client == nil {
			return NewKeyErrorMsg{
				Key: req.Key,
				Err: errors.New("client is nil"),
			}
		}

		item := &memcache.Item{
			Key:        req.Key,
			Value:      []byte(req.Value),
			Expiration: req.TTL,
			Flags:      req.Flags,
		}

		err := client.Set(item)
		if err != nil {
			return NewKeyErrorMsg{
				Key: req.Key,
				Err: err,
			}
		}

		return KeyCreatedMsg{Key: req.Key}
	}
}

// HandleNewKeyResult processes the result of a new key operation.
// Returns a NewKeyResult struct with information about what happened
// and what the UI should do next.
func HandleNewKeyResult(msg tea.Msg) NewKeyResult {
	switch m := msg.(type) {
	case KeyCreatedMsg:
		return NewKeyResult{
			ShouldRefresh: true,
			CreatedKey:    m.Key,
		}
	case NewKeyErrorMsg:
		errStr := ""
		if m.Err != nil {
			errStr = m.Err.Error()
		}
		return NewKeyResult{
			ShouldRefresh: false,
			Error:         errStr,
			FailedKey:     m.Key,
		}
	default:
		return NewKeyResult{}
	}
}

// ValidateKeyName validates a Memcached key name.
// Returns an error if the key is invalid according to Memcached rules:
// - Key cannot be empty
// - Key cannot contain spaces, newlines, or control characters
// - Key cannot exceed 250 bytes
func ValidateKeyName(key string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}

	if len(key) > MaxKeyLength {
		return fmt.Errorf("key cannot exceed %d bytes (got %d)", MaxKeyLength, len(key))
	}

	// Check for invalid characters
	for i, r := range key {
		// Space is not allowed
		if r == ' ' {
			return errors.New("key cannot contain spaces")
		}

		// Newlines are not allowed
		if r == '\n' || r == '\r' {
			return errors.New("key cannot contain newlines")
		}

		// Control characters are not allowed (including null byte)
		if unicode.IsControl(r) {
			return fmt.Errorf("key cannot contain control characters (found at position %d)", i)
		}
	}

	// Check for whitespace other than space (tabs, etc.)
	if strings.ContainsAny(key, "\t\v\f") {
		return errors.New("key cannot contain whitespace characters")
	}

	return nil
}

// CreateNewKeyDialog creates an input dialog for entering a new key name.
// The dialog includes validation for Memcached key requirements.
func CreateNewKeyDialog() *dialog.InputDialog {
	return dialog.NewInput("New Key").
		WithPlaceholder("Enter key name...").
		WithValidator(ValidateKeyName)
}

// CreateValueInputDialog creates an input dialog for entering a value for a key.
// The key name is stored in the context for retrieval after input is submitted.
func CreateValueInputDialog(key string) *dialog.InputDialog {
	title := fmt.Sprintf("Value for: %s", key)
	return dialog.NewInput(title).
		WithPlaceholder("Enter value...").
		WithContext(NewKeyContext{Key: key})
}

// ExtractNewKeyContext extracts the key from an input result context.
// Returns the key string and a boolean indicating success.
func ExtractNewKeyContext(ctx interface{}) (string, bool) {
	if ctx == nil {
		return "", false
	}

	// Check if context is a string (key directly)
	if key, ok := ctx.(string); ok {
		return key, true
	}

	// Check if context is a NewKeyContext struct
	if nc, ok := ctx.(NewKeyContext); ok {
		return nc.Key, true
	}

	return "", false
}

// ProcessNewKeyInputResult creates a NewKeyRequest from user input.
// This is called after the user has entered both key name and value.
func ProcessNewKeyInputResult(key, value string, ttl int32) *NewKeyRequest {
	return &NewKeyRequest{
		Key:   key,
		Value: value,
		TTL:   ttl,
	}
}

// HandleNewKeyConfirm processes a new key request and returns the appropriate command.
// This is called after the user has entered both key name and value.
func HandleNewKeyConfirm(client Setter, req NewKeyRequest) tea.Cmd {
	return NewKeyCmd(client, req)
}

// ProcessKeyNameInput processes the result of the key name input dialog.
// Returns the key name if valid, or an error if cancelled or invalid.
func ProcessKeyNameInput(result dialog.InputResultMsg) (string, error) {
	if result.Cancelled {
		return "", errors.New("cancelled")
	}

	key := result.Value
	if err := ValidateKeyName(key); err != nil {
		return "", err
	}

	return key, nil
}

// ProcessValueInput processes the result of the value input dialog.
// Returns the key name from context and the value, or an error if cancelled.
func ProcessValueInput(result dialog.InputResultMsg) (string, string, error) {
	if result.Cancelled {
		return "", "", errors.New("cancelled")
	}

	key, ok := ExtractNewKeyContext(result.Context)
	if !ok {
		return "", "", errors.New("key context not found")
	}

	return key, result.Value, nil
}
