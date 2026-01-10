package models_test

import (
	"testing"

	"github.com/nnnkkk7/memtui/models"
)

func TestKeyInfo_Fields(t *testing.T) {
	ki := models.KeyInfo{
		Key:        "user:123",
		Expiration: 1704067200,
		LastAccess: 1704060000,
		CAS:        12345,
		Fetch:      true,
		SlabClass:  1,
		Size:       256,
	}

	if ki.Key != "user:123" {
		t.Errorf("expected key 'user:123', got '%s'", ki.Key)
	}
	if ki.Size != 256 {
		t.Errorf("expected size 256, got %d", ki.Size)
	}
}

func TestParseMetadumpLine(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantKey string
		wantExp int64
		wantErr bool
	}{
		{
			name:    "normal line",
			input:   "key=test exp=1704067200 la=1704060000 cas=12345 fetch=yes cls=1 size=100",
			wantKey: "test",
			wantExp: 1704067200,
			wantErr: false,
		},
		{
			name:    "line with trailing CR",
			input:   "key=test exp=0 la=123 cas=456 fetch=yes cls=1 size=100\r",
			wantKey: "test",
			wantExp: 0,
			wantErr: false,
		},
		{
			name:    "key with colon delimiter",
			input:   "key=user:profile:123 exp=3600 la=1000 cas=789 fetch=no cls=2 size=512",
			wantKey: "user:profile:123",
			wantExp: 3600,
			wantErr: false,
		},
		{
			name:    "empty line",
			input:   "",
			wantKey: "",
			wantErr: true,
		},
		{
			name:    "malformed line - no key",
			input:   "exp=0 la=123 cas=456",
			wantKey: "",
			wantErr: true,
		},
		{
			name:    "END marker",
			input:   "END",
			wantKey: "",
			wantErr: true,
		},
		{
			name:    "URL-encoded key with colon",
			input:   "key=user%3A123 exp=0 la=123 cas=456 fetch=yes cls=1 size=100",
			wantKey: "user:123",
			wantExp: 0,
			wantErr: false,
		},
		{
			name:    "URL-encoded key with multiple colons",
			input:   "key=cache%3Aproducts%3Alist exp=3600 la=1000 cas=789 fetch=no cls=2 size=512",
			wantKey: "cache:products:list",
			wantExp: 3600,
			wantErr: false,
		},
		{
			name:    "URL-encoded key with special chars",
			input:   "key=data%2Fpath%3Fquery%3D1 exp=0 la=0 cas=0 fetch=no cls=1 size=50",
			wantKey: "data/path?query=1",
			wantExp: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ki, err := models.ParseMetadumpLine(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if ki.Key != tt.wantKey {
				t.Errorf("expected key '%s', got '%s'", tt.wantKey, ki.Key)
			}
			if ki.Expiration != tt.wantExp {
				t.Errorf("expected expiration %d, got %d", tt.wantExp, ki.Expiration)
			}
		})
	}
}

func TestParseMetadumpLine_AllFields(t *testing.T) {
	input := "key=session:abc exp=1704067200 la=1704060000 cas=99999 fetch=yes cls=3 size=1024"
	ki, err := models.ParseMetadumpLine(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ki.Key != "session:abc" {
		t.Errorf("Key: expected 'session:abc', got '%s'", ki.Key)
	}
	if ki.Expiration != 1704067200 {
		t.Errorf("Expiration: expected 1704067200, got %d", ki.Expiration)
	}
	if ki.LastAccess != 1704060000 {
		t.Errorf("LastAccess: expected 1704060000, got %d", ki.LastAccess)
	}
	if ki.CAS != 99999 {
		t.Errorf("CAS: expected 99999, got %d", ki.CAS)
	}
	if !ki.Fetch {
		t.Error("Fetch: expected true, got false")
	}
	if ki.SlabClass != 3 {
		t.Errorf("SlabClass: expected 3, got %d", ki.SlabClass)
	}
	if ki.Size != 1024 {
		t.Errorf("Size: expected 1024, got %d", ki.Size)
	}
}

func TestKeyInfo_IsExpired(t *testing.T) {
	tests := []struct {
		name       string
		expiration int64
		now        int64
		want       bool
	}{
		{
			name:       "permanent (exp=0)",
			expiration: 0,
			now:        1704067200,
			want:       false,
		},
		{
			name:       "not expired",
			expiration: 1704067200,
			now:        1704060000,
			want:       false,
		},
		{
			name:       "expired",
			expiration: 1704060000,
			now:        1704067200,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ki := models.KeyInfo{Expiration: tt.expiration}
			if got := ki.IsExpiredAt(tt.now); got != tt.want {
				t.Errorf("IsExpiredAt(%d) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestSortKeyInfos(t *testing.T) {
	keys := []models.KeyInfo{
		{Key: "zebra", Size: 100},
		{Key: "alpha", Size: 300},
		{Key: "beta", Size: 200},
	}

	// Sort by key name
	sorted := models.SortKeyInfos(keys, models.SortByKey)
	if sorted[0].Key != "alpha" || sorted[1].Key != "beta" || sorted[2].Key != "zebra" {
		t.Errorf("SortByKey failed: got %v", sorted)
	}

	// Sort by size
	sorted = models.SortKeyInfos(keys, models.SortBySize)
	if sorted[0].Size != 100 || sorted[1].Size != 200 || sorted[2].Size != 300 {
		t.Errorf("SortBySize failed: got %v", sorted)
	}
}

func TestFilterKeyInfos(t *testing.T) {
	keys := []models.KeyInfo{
		{Key: "user:123"},
		{Key: "user:456"},
		{Key: "session:abc"},
		{Key: "cache:data"},
	}

	// Filter by prefix
	filtered := models.FilterKeyInfos(keys, "user:")
	if len(filtered) != 2 {
		t.Errorf("expected 2 results, got %d", len(filtered))
	}

	// Filter with no match
	filtered = models.FilterKeyInfos(keys, "nonexistent")
	if len(filtered) != 0 {
		t.Errorf("expected 0 results, got %d", len(filtered))
	}

	// Empty filter returns all
	filtered = models.FilterKeyInfos(keys, "")
	if len(filtered) != 4 {
		t.Errorf("expected 4 results, got %d", len(filtered))
	}
}
