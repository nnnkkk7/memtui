package app

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/atotto/clipboard"
	"github.com/bradfitz/gomemcache/memcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/models"
)

// calculateRemainingTTL calculates the remaining TTL from an absolute expiration timestamp.
// Returns 0 if the key has no expiration or has already expired.
func calculateRemainingTTL(expiration int64) int32 {
	if expiration == 0 {
		return 0 // No expiration
	}
	remaining := expiration - time.Now().Unix()
	if remaining <= 0 {
		return 0 // Already expired, set to no expiration
	}
	// Memcached accepts int32 for expiration, cap at max int32
	if remaining > math.MaxInt32 {
		return math.MaxInt32
	}
	return int32(remaining)
}

// ClipboardCopyMsg is sent when clipboard copy is successful
type ClipboardCopyMsg struct{}

// ClipboardErrorMsg is sent when clipboard copy fails
type ClipboardErrorMsg struct {
	Err error
}

func (m *Model) connectCmd() tea.Cmd {
	return func() tea.Msg {
		detector := client.NewCapabilityDetector()
		caps, err := detector.Detect(m.addr)
		if err != nil {
			return ErrorMsg{Err: err.Error()}
		}
		return ConnectedMsg{
			Version:          caps.Version,
			SupportsMetadump: caps.SupportsMetadump,
		}
	}
}

func (m *Model) loadKeysCmd() tea.Cmd {
	return func() tea.Msg {
		// Skip key enumeration if server doesn't support metadump
		if !m.supportsMetadump {
			return KeysLoadedMsg{Keys: []models.KeyInfo{}}
		}

		enum := client.NewKeyEnumerator(m.addr)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		keys, err := enum.EnumerateAll(ctx)
		if err != nil {
			return ErrorMsg{Err: err.Error()}
		}
		return KeysLoadedMsg{Keys: keys}
	}
}

func (m *Model) loadValueCmd(key string) tea.Cmd {
	return func() tea.Msg {
		if m.mcClient == nil {
			return ErrorMsg{Err: "client not connected"}
		}

		// Use GetWithCAS to get full metadata including CAS token
		casItem, err := m.mcClient.GetWithCAS(key)
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("failed to get value: %v", err)}
		}
		return ValueLoadedMsg{
			Key:        key,
			Value:      casItem.Value,
			Flags:      casItem.Flags,
			Expiration: casItem.Expiration,
			CAS:        casItem.CAS,
		}
	}
}

func (m *Model) saveValueCmd(key string, value []byte, keyInfo *models.KeyInfo, casItem *client.CASItem) tea.Cmd {
	return func() tea.Msg {
		if m.mcClient == nil {
			return ErrorMsg{Err: "client not connected"}
		}

		// Preserve Flags from CAS item and TTL from key info
		flags := uint32(0)
		expiration := int32(0)
		if casItem != nil {
			flags = casItem.Flags
		}
		if keyInfo != nil {
			expiration = calculateRemainingTTL(keyInfo.Expiration)
		}

		err := m.mcClient.Set(&memcache.Item{
			Key:        key,
			Value:      value,
			Flags:      flags,
			Expiration: expiration,
		})
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("failed to save value: %v", err)}
		}

		return KeyCreatedMsg{Key: key}
	}
}

func (m *Model) saveValueWithCASCmd(key string, value []byte, originalCASItem *client.CASItem, keyInfo *models.KeyInfo) tea.Cmd {
	return func() tea.Msg {
		if m.mcClient == nil {
			return ErrorMsg{Err: "client not connected"}
		}

		// Re-fetch the item to get a valid mcItem for CompareAndSwap
		casItem, err := m.mcClient.GetWithCAS(key)
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("failed to get current value for CAS: %v", err)}
		}

		// Note: CAS conflict detection relies on CompareAndSwap internally.
		// The CAS value returned by extractCASValue is a placeholder (always 1),
		// so pre-comparison here would be meaningless. The actual CAS token is
		// stored in the unexported casid field of memcache.Item.

		// Update the CAS item with new value, preserving flags from original CAS item
		// and calculating remaining TTL from key info (metadump expiration)
		casItem.Value = value
		if originalCASItem != nil {
			casItem.Flags = originalCASItem.Flags
		}
		if keyInfo != nil {
			casItem.Expiration = calculateRemainingTTL(keyInfo.Expiration)
		}

		err = m.mcClient.CompareAndSwap(casItem)
		if err != nil {
			if client.IsCASConflict(err) {
				return ErrorMsg{Err: "CAS conflict: value was modified by another client. Please reload and try again."}
			}
			return ErrorMsg{Err: fmt.Sprintf("failed to save value: %v", err)}
		}

		return KeyCreatedMsg{Key: key}
	}
}

func (m *Model) copyToClipboardCmd(value []byte) tea.Cmd {
	return func() tea.Msg {
		err := clipboard.WriteAll(string(value))
		if err != nil {
			return ClipboardErrorMsg{Err: err}
		}
		return ClipboardCopyMsg{}
	}
}
