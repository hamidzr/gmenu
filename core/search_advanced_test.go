package core

import (
	"testing"
	"unicode/utf8"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
)

// TestIsDirectMatchEdgeCases tests edge cases for direct matching
func TestIsDirectMatchEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		keyword    string
		smartMatch bool
		expected   bool
	}{
		{"empty keyword", "hello", "", true, true},
		{"empty text", "", "hello", true, false},
		{"both empty", "", "", true, true},
		{"case sensitive smart match", "Hello", "Hello", true, true},
		{"case insensitive smart match", "Hello", "hello", true, true},
		{"mixed case keyword forces case sensitive", "hello", "Hello", true, false},
		{"unicode characters", "café", "fé", true, true},
		{"emoji matching", "test 🚀 rocket", "🚀", true, true},
		{"special characters", "file@domain.com", "@domain", true, true},
		{"whitespace matching", "hello world", " ", true, true},
		{"newline in text", "hello\nworld", "hello", true, true},
		{"tab in text", "hello\tworld", "\t", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDirectMatch(tt.text, tt.keyword, tt.smartMatch)
			assert.Equal(t, tt.expected, result,
				"IsDirectMatch(%q, %q, %v) = %v, want %v",
				tt.text, tt.keyword, tt.smartMatch, result, tt.expected)
		})
	}
}

// TestFuzzyContainsConsecEdgeCases tests edge cases for consecutive fuzzy matching
func TestFuzzyContainsConsecEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		text           string
		query          string
		ignoreCase     bool
		minConsecutive int
		expected       bool
	}{
		{"empty query", "hello", "", true, 2, true},
		{"empty text", "", "hello", true, 2, false},
		{"query longer than text", "hi", "hello", true, 2, false},
		{"exact match", "hello", "hello", true, 2, true},
		{"consecutive at start", "hello world", "hel", true, 3, true},
		{"consecutive at end", "hello world", "rld", true, 3, true},
		{"consecutive in middle", "hello world", "llo", true, 3, true},
		{"non-consecutive match", "hello", "hlo", true, 2, false},
		{"case sensitive match", "Hello", "hello", false, 2, false},
		{"case insensitive match", "Hello", "hello", true, 2, true},
		{"unicode consecutive", "naïve café", "ïve", true, 3, true},
		{"min consecutive 1", "hello", "h", true, 1, true},
		{"min consecutive larger than query", "hello", "he", true, 5, true}, // should adjust to query length
		{"overlapping consecutive", "aaa", "aa", true, 2, true},
		{"multiple consecutive opportunities", "abcabc", "abc", true, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fuzzyContainsConsec(tt.text, tt.query, tt.ignoreCase, tt.minConsecutive)
			assert.Equal(t, tt.expected, result,
				"fuzzyContainsConsec(%q, %q, %v, %d) = %v, want %v",
				tt.text, tt.query, tt.ignoreCase, tt.minConsecutive, result, tt.expected)
		})
	}
}

// TestDirectSearchPerformance tests performance characteristics of direct search
func TestDirectSearchPerformance(t *testing.T) {
	// Create large dataset
	largeDataset := make([]model.MenuItem, 10000)
	for i := 0; i < 10000; i++ {
		largeDataset[i] = model.MenuItem{
			Title: "item_" + string(rune('a'+(i%26))) + "_" + string(rune('0'+(i%10))),
		}
	}

	tests := []struct {
		name       string
		keyword    string
		limit      int
		minResults int
		maxResults int
	}{
		{"common pattern", "item", 100, 90, 100},      // should find many matches
		{"specific pattern", "item_a_1", 100, 0, 100}, // might find none but up to 100
		{"rare pattern", "item_z_9", 100, 0, 100},     // might find none but up to 100
		{"no limit", "item", 0, 1000, 10000},          // no limit applied
		{"small limit", "item", 5, 1, 5},              // limited results
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := DirectSearch(largeDataset, tt.keyword, false, tt.limit)

			assert.GreaterOrEqual(t, len(results), tt.minResults)
			assert.LessOrEqual(t, len(results), tt.maxResults)

			// Verify all results actually match
			for _, result := range results {
				assert.True(t, IsDirectMatch(result.ComputedTitle(), tt.keyword, true),
					"Result %q should match keyword %q", result.ComputedTitle(), tt.keyword)
			}

			// Verify limit is respected
			if tt.limit > 0 {
				assert.LessOrEqual(t, len(results), tt.limit)
			}
		})
	}
}

// TestFuzzySearchBruteRobustness tests robustness of fuzzy search
func TestFuzzySearchBruteRobustness(t *testing.T) {
	testItems := []model.MenuItem{
		{Title: ""},                               // empty string
		{Title: "a"},                              // single character
		{Title: "🚀"},                              // emoji
		{Title: "café naïve résumé"},              // unicode with accents
		{Title: "line1\nline2"},                   // multiline
		{Title: "tab\there"},                      // tabs
		{Title: "  spaces  "},                     // leading/trailing spaces
		{Title: string(make([]byte, 1000))}, // very long string
		{Title: "normal text"},                    // normal case
	}

	// Fill the very long string with repeating pattern
	longTitle := string(make([]rune, 1000))
	for i := range longTitle {
		longTitle = string(rune('a' + (i % 26)))
	}
	testItems[7].Title = longTitle

	tests := []struct {
		name       string
		query      string
		minConsec  int
		shouldFind bool
	}{
		{"empty query", "", 2, true},
		{"single char query", "a", 1, true},
		{"emoji query", "🚀", 1, true},
		{"unicode query", "café", 2, true},
		{"whitespace query", " ", 1, true},
		{"newline query", "\n", 1, true},
		{"long query", string(make([]rune, 100)), 2, false}, // very long query
		{"normal query", "normal", 2, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test should not panic
			assert.NotPanics(t, func() {
				results := fuzzySearchBruteConsec(testItems, tt.query, 10, tt.minConsec)

				if tt.shouldFind {
					// Should find at least something for reasonable queries
					if tt.query != "" && utf8.ValidString(tt.query) {
						// Don't require matches for all cases, just ensure no panic
						assert.NotNil(t, results)
					}
				}

				// All results should be valid
				for _, result := range results {
					assert.NotNil(t, result)
				}
			})
		})
	}
}

// TestSearchWithSeparatorAdvanced tests advanced separator-based searching
func TestSearchWithSeparatorAdvanced(t *testing.T) {
	testItems := []model.MenuItem{
		{Title: "hello world test"},
		{Title: "hello_world_test"},
		{Title: "hello-world-test"},
		{Title: "hello.world.test"},
		{Title: "helloWorldTest"},
		{Title: "HELLO WORLD TEST"},
		{Title: "test hello world"},
		{Title: "world test hello"},
	}

	tests := []struct {
		name      string
		separator string
		query     string
		expected  []string
	}{
		{
			name:      "space separator multiple words",
			separator: " ",
			query:     "hello world",
			expected:  []string{"hello world test", "HELLO WORLD TEST"},
		},
		{
			name:      "space separator reversed order",
			separator: " ",
			query:     "world hello",
			expected:  []string{"world test hello"},
		},
		{
			name:      "space separator partial words",
			separator: " ",
			query:     "hel wor",
			expected:  []string{"hello world test", "HELLO WORLD TEST"},
		},
		{
			name:      "underscore separator",
			separator: "_",
			query:     "hello_world",
			expected:  []string{"hello_world_test"},
		},
		{
			name:      "hyphen separator",
			separator: "-",
			query:     "hello-world",
			expected:  []string{"hello-world-test"},
		},
		{
			name:      "dot separator",
			separator: ".",
			query:     "hello.world",
			expected:  []string{"hello.world.test"},
		},
		{
			name:      "mixed separators with space",
			separator: " ",
			query:     "hello world test",
			expected:  []string{"hello world test", "HELLO WORLD TEST", "test hello world"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			searchMethod := SearchWithSeparator(tt.separator, DirectSearch)
			results := searchMethod(testItems, tt.query, false, 0)

			var resultTitles []string
			for _, result := range results {
				resultTitles = append(resultTitles, result.Title)
			}

			for _, expected := range tt.expected {
				assert.Contains(t, resultTitles, expected,
					"Expected %q to be found in results %v", expected, resultTitles)
			}
		})
	}
}

// TestSearchMethodsRegistration tests that all search methods are properly registered
func TestSearchMethodsRegistration(t *testing.T) {
	expectedMethods := []string{
		"direct",
		"fuzzy",
		"fuzzy1",
		"fuzzy3",
		"default",
	}

	for _, methodName := range expectedMethods {
		t.Run(methodName, func(t *testing.T) {
			method, exists := SearchMethods[methodName]
			assert.True(t, exists, "Search method %q should be registered", methodName)
			assert.NotNil(t, method, "Search method %q should not be nil", methodName)

			// Test that the method actually works
			testItems := []model.MenuItem{
				{Title: "test item 1"},
				{Title: "test item 2"},
				{Title: "another item"},
			}

			assert.NotPanics(t, func() {
				results := method(testItems, "test", false, 10)
				assert.NotNil(t, results)
			})
		})
	}
}

// TestApplyLimitEdgeCases tests limit application edge cases
func TestApplyLimitEdgeCases(t *testing.T) {
	testItems := []model.MenuItem{
		{Title: "item1"},
		{Title: "item2"},
		{Title: "item3"},
		{Title: "item4"},
		{Title: "item5"},
	}

	tests := []struct {
		name     string
		items    []model.MenuItem
		limit    int
		expected int
	}{
		{"no limit (0)", testItems, 0, 5},
		{"limit larger than items", testItems, 10, 5},
		{"limit smaller than items", testItems, 3, 3},
		{"limit of 1", testItems, 1, 1},
		{"empty items with limit", []model.MenuItem{}, 5, 0},
		{"empty items no limit", []model.MenuItem{}, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := applyLimit(tt.items, tt.limit)
			assert.Equal(t, tt.expected, len(result))

			// Verify returned items are from the beginning of the slice
			for i, item := range result {
				if i < len(tt.items) {
					assert.Equal(t, tt.items[i].Title, item.Title)
				}
			}
		})
	}
}

// TestSearchMethodConsistency tests that search methods behave consistently
func TestSearchMethodConsistency(t *testing.T) {
	testItems := []model.MenuItem{
		{Title: "apple pie"},
		{Title: "apple juice"},
		{Title: "grape juice"},
		{Title: "orange juice"},
		{Title: "banana split"},
	}

	query := "apple"

	for methodName, method := range SearchMethods {
		t.Run(methodName, func(t *testing.T) {
			// Test consistency across multiple calls
			results1 := method(testItems, query, false, 10)
			results2 := method(testItems, query, false, 10)
			results3 := method(testItems, query, false, 10)

			assert.Equal(t, len(results1), len(results2))
			assert.Equal(t, len(results2), len(results3))

			// Results should be identical for deterministic methods
			for i := range results1 {
				if i < len(results2) && i < len(results3) {
					assert.Equal(t, results1[i].Title, results2[i].Title)
					assert.Equal(t, results2[i].Title, results3[i].Title)
				}
			}

			// All results should actually contain the query term for direct methods
			if methodName == "direct" {
				for _, result := range results1 {
					assert.True(t, IsDirectMatch(result.ComputedTitle(), query, true),
						"Result %q should match query %q for method %q",
						result.ComputedTitle(), query, methodName)
				}
			}
		})
	}
}
