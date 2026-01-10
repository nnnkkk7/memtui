package app

import "github.com/nnnkkk7/memtui/models"

// ConnectedMsg is sent when connection is established
type ConnectedMsg struct {
	Version          string
	SupportsMetadump bool
}

// KeysLoadedMsg is sent when keys are loaded
type KeysLoadedMsg struct {
	Keys []models.KeyInfo
}

// ErrorMsg is sent when an error occurs
type ErrorMsg struct {
	Err string
}

// KeySelectedMsg is sent when a key is selected
type KeySelectedMsg struct {
	Key models.KeyInfo
}

// ValueLoadedMsg is sent when a value is loaded
type ValueLoadedMsg struct {
	Key        string
	Value      []byte
	Flags      uint32
	Expiration int32
	CAS        uint64
}

// DeleteKeyMsg is sent when user requests to delete a key (before confirmation)
type DeleteKeyMsg struct {
	Key string
}

// DeleteConfirmMsg is sent when user confirms deletion
type DeleteConfirmMsg struct {
	Key string
}

// KeyDeletedMsg is sent when a key is successfully deleted
type KeyDeletedMsg struct {
	Key string
}

// DeleteErrorMsg is sent when key deletion fails
type DeleteErrorMsg struct {
	Key string
	Err error
}

