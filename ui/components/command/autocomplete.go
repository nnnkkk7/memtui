// Package command provides a command palette component for the memtui application.
package command

import (
	"strings"
	"unicode"
)

// FuzzyMatch performs a fuzzy match of query against text.
// It returns whether there is a match and a score (higher is better).
// The matching is case-insensitive and supports substring matching.
func FuzzyMatch(query, text string) (bool, int) {
	if query == "" {
		return true, 0
	}

	queryLower := strings.ToLower(query)
	textLower := strings.ToLower(text)

	// Check for substring match
	if strings.Contains(textLower, queryLower) {
		// Calculate score based on match position and length
		index := strings.Index(textLower, queryLower)
		score := 100

		// Prefer matches at the beginning
		if index == 0 {
			score += 50
		} else if index > 0 && (textLower[index-1] == ' ' || textLower[index-1] == '_' || textLower[index-1] == '-') {
			// Word boundary match
			score += 25
		}

		// Prefer exact length matches
		if len(query) == len(text) {
			score += 50
		}

		// Bonus for case-sensitive match
		if strings.Contains(text, query) {
			score += 10
		}

		return true, score
	}

	// Try fuzzy character-by-character matching
	// Characters must appear in order but not necessarily consecutively
	queryIdx := 0
	score := 0
	lastMatchIdx := -1
	consecutiveBonus := 0

	for i, char := range textLower {
		if queryIdx < len(queryLower) && char == rune(queryLower[queryIdx]) {
			queryIdx++
			score += 10

			// Bonus for consecutive matches
			if lastMatchIdx == i-1 {
				consecutiveBonus += 5
			}

			// Bonus for matching at word boundaries
			if i == 0 || !unicode.IsLetter(rune(textLower[i-1])) {
				score += 15
			}

			lastMatchIdx = i
		}
	}

	// All query characters must match
	if queryIdx == len(queryLower) {
		return true, score + consecutiveBonus
	}

	return false, 0
}

// SortByScore sorts commands by their fuzzy match score (descending).
// This is a simple bubble sort for small lists.
type scoredCommand struct {
	command Command
	score   int
}

// RankCommands ranks commands by their fuzzy match score against the query.
// It returns commands sorted by score (highest first).
func RankCommands(commands []Command, query string) []Command {
	if query == "" {
		return commands
	}

	scored := make([]scoredCommand, 0, len(commands))
	for _, cmd := range commands {
		nameMatch, nameScore := FuzzyMatch(query, cmd.Name)
		descMatch, descScore := FuzzyMatch(query, cmd.Description)

		if nameMatch || descMatch {
			score := 0
			if nameMatch {
				score = nameScore
			}
			// Also add description score (weighted less)
			if descMatch {
				score += descScore / 2
			}
			scored = append(scored, scoredCommand{command: cmd, score: score})
		}
	}

	// Sort by score (descending) using simple bubble sort
	for i := 0; i < len(scored)-1; i++ {
		for j := 0; j < len(scored)-i-1; j++ {
			if scored[j].score < scored[j+1].score {
				scored[j], scored[j+1] = scored[j+1], scored[j]
			}
		}
	}

	result := make([]Command, len(scored))
	for i, sc := range scored {
		result[i] = sc.command
	}

	return result
}
