// Package command provides a command palette component for the memtui application.
package command

import (
	"testing"
)

// TestFuzzyMatch_Empty verifies empty query matches everything.
func TestFuzzyMatch_Empty(t *testing.T) {
	match, score := FuzzyMatch("", "Hello World")
	if !match {
		t.Error("expected empty query to match")
	}
	if score != 0 {
		t.Errorf("expected score 0 for empty query, got %d", score)
	}
}

// TestFuzzyMatch_Substring verifies substring matching.
func TestFuzzyMatch_Substring(t *testing.T) {
	tests := []struct {
		query    string
		text     string
		expected bool
	}{
		{"hello", "Hello World", true},
		{"WORLD", "Hello World", true},
		{"ello", "Hello World", true},
		{"xyz", "Hello World", false},
		{"ref", "Refresh keys", true},
		{"key", "Delete key", true},
	}

	for _, tc := range tests {
		match, _ := FuzzyMatch(tc.query, tc.text)
		if match != tc.expected {
			t.Errorf("FuzzyMatch(%q, %q) = %v, expected %v",
				tc.query, tc.text, match, tc.expected)
		}
	}
}

// TestFuzzyMatch_CaseInsensitive verifies case-insensitive matching.
func TestFuzzyMatch_CaseInsensitive(t *testing.T) {
	tests := []struct {
		query string
		text  string
	}{
		{"REFRESH", "Refresh keys"},
		{"refresh", "Refresh keys"},
		{"ReFrEsH", "Refresh keys"},
		{"KEYS", "Refresh keys"},
	}

	for _, tc := range tests {
		match, _ := FuzzyMatch(tc.query, tc.text)
		if !match {
			t.Errorf("expected %q to match %q (case-insensitive)", tc.query, tc.text)
		}
	}
}

// TestFuzzyMatch_Score verifies that scores are calculated properly.
func TestFuzzyMatch_Score(t *testing.T) {
	// Beginning matches should score higher
	_, scoreBegin := FuzzyMatch("ref", "Refresh keys")
	_, scoreMiddle := FuzzyMatch("esh", "Refresh keys")

	if scoreBegin <= scoreMiddle {
		t.Errorf("expected beginning match score (%d) > middle match score (%d)",
			scoreBegin, scoreMiddle)
	}

	// Exact matches should score highest
	_, scoreExact := FuzzyMatch("Refresh", "Refresh")
	_, scorePartial := FuzzyMatch("Refresh", "Refresh keys")

	if scoreExact <= scorePartial {
		t.Errorf("expected exact match score (%d) > partial match score (%d)",
			scoreExact, scorePartial)
	}
}

// TestFuzzyMatch_FuzzyCharacters verifies fuzzy character-by-character matching.
func TestFuzzyMatch_FuzzyCharacters(t *testing.T) {
	// Characters in order but not consecutive
	match, _ := FuzzyMatch("rk", "Refresh keys")
	if !match {
		t.Error("expected 'rk' to fuzzy match 'Refresh keys'")
	}

	match, _ = FuzzyMatch("dk", "Delete key")
	if !match {
		t.Error("expected 'dk' to fuzzy match 'Delete key'")
	}

	// Characters out of order should not match
	match, _ = FuzzyMatch("kr", "Refresh keys")
	if match {
		t.Error("expected 'kr' NOT to match 'Refresh keys' (out of order)")
	}
}

// TestRankCommands_Empty verifies empty query returns all commands.
func TestRankCommands_Empty(t *testing.T) {
	commands := []Command{
		{Name: "Alpha"},
		{Name: "Beta"},
		{Name: "Gamma"},
	}

	ranked := RankCommands(commands, "")

	if len(ranked) != len(commands) {
		t.Errorf("expected %d commands, got %d", len(commands), len(ranked))
	}
}

// TestRankCommands_Filter verifies that commands are filtered.
func TestRankCommands_Filter(t *testing.T) {
	commands := []Command{
		{Name: "Refresh keys"},
		{Name: "Delete key"},
		{Name: "Show stats"},
	}

	ranked := RankCommands(commands, "key")

	if len(ranked) != 2 {
		t.Errorf("expected 2 matching commands, got %d", len(ranked))
	}

	// Verify only key-related commands are returned
	for _, cmd := range ranked {
		if cmd.Name != "Refresh keys" && cmd.Name != "Delete key" {
			t.Errorf("unexpected command in results: %q", cmd.Name)
		}
	}
}

// TestRankCommands_Sorting verifies that commands are sorted by score.
func TestRankCommands_Sorting(t *testing.T) {
	commands := []Command{
		{Name: "Show statistics"},   // "stat" in middle
		{Name: "Statistics viewer"}, // "stat" at beginning
		{Name: "View stats"},        // "stat" in middle
	}

	ranked := RankCommands(commands, "stat")

	if len(ranked) != 3 {
		t.Fatalf("expected 3 matching commands, got %d", len(ranked))
	}

	// First result should have "stat" at the beginning
	if ranked[0].Name != "Statistics viewer" {
		t.Errorf("expected 'Statistics viewer' to be first, got %q", ranked[0].Name)
	}
}

// TestRankCommands_NoMatch verifies empty result for no matches.
func TestRankCommands_NoMatch(t *testing.T) {
	commands := []Command{
		{Name: "Refresh keys"},
		{Name: "Delete key"},
	}

	ranked := RankCommands(commands, "xyz")

	if len(ranked) != 0 {
		t.Errorf("expected 0 commands, got %d", len(ranked))
	}
}

// TestRankCommands_DescriptionMatch verifies that description is also considered.
func TestRankCommands_DescriptionMatch(t *testing.T) {
	commands := []Command{
		{Name: "Action A", Description: "Reload the data"},
		{Name: "Action B", Description: "Nothing here"},
	}

	ranked := RankCommands(commands, "reload")

	if len(ranked) != 1 {
		t.Errorf("expected 1 matching command, got %d", len(ranked))
	}

	if len(ranked) > 0 && ranked[0].Name != "Action A" {
		t.Errorf("expected 'Action A' to match via description, got %q", ranked[0].Name)
	}
}
