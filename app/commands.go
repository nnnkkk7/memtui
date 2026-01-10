package app

import (
	"context"
	"fmt"
	"time"

	"github.com/atotto/clipboard"
	"github.com/bradfitz/gomemcache/memcache"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/nnnkkk7/memtui/client"
	"github.com/nnnkkk7/memtui/models"
)

// ClipboardCopyMsg is sent when clipboard copy is successful
type ClipboardCopyMsg struct{}

// ClipboardErrorMsg is sent when clipboard copy fails
type ClipboardErrorMsg struct {
	Err error
}

func (m *Model) connectCmd() tea.Cmd {
	return func() tea.Msg {
		detector := client.NewCapabilityDetector()
		cap, err := detector.Detect(m.addr)
		if err != nil {
			return ErrorMsg{Err: err.Error()}
		}
		return ConnectedMsg{
			Version:          cap.Version,
			SupportsMetadump: cap.SupportsMetadump,
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

func (m *Model) saveValueCmd(key string, value []byte) tea.Cmd {
	return func() tea.Msg {
		if m.mcClient == nil {
			return ErrorMsg{Err: "client not connected"}
		}

		err := m.mcClient.Set(&memcache.Item{
			Key:   key,
			Value: value,
		})
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("failed to save value: %v", err)}
		}

		return KeyCreatedMsg{Key: key}
	}
}

func (m *Model) saveValueWithCASCmd(key string, value []byte, originalCASItem *client.CASItem) tea.Cmd {
	return func() tea.Msg {
		if m.mcClient == nil {
			return ErrorMsg{Err: "client not connected"}
		}

		// Re-fetch the item to get a valid mcItem for CompareAndSwap
		casItem, err := m.mcClient.GetWithCAS(key)
		if err != nil {
			return ErrorMsg{Err: fmt.Sprintf("failed to get current value for CAS: %v", err)}
		}

		// Check if the CAS value has changed (item was modified)
		if casItem.CAS != originalCASItem.CAS {
			return ErrorMsg{Err: "CAS conflict: value was modified by another client. Please reload and try again."}
		}

		// Update the CAS item with new value, preserving flags and expiration
		casItem.Value = value
		casItem.Flags = originalCASItem.Flags
		casItem.Expiration = originalCASItem.Expiration

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
