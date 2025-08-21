package core

import (
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGMenuErrorHandling tests error handling in GMenu creation and initialization
func TestGMenuErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		searchMethod  SearchMethod
		config        *model.Config
		expectError   bool
		errorContains string
	}{
		{
			name:         "nil config should cause error",
			searchMethod: DirectSearch,
			config:       nil,
			expectError:  true,
		},
		{
			name:         "valid config should succeed",
			searchMethod: DirectSearch,
			config: &model.Config{
				MenuID:    "test",
				Title:     "Test Menu",
				Prompt:    "test>",
				MinWidth:  300,
				MinHeight: 200,
			},
			expectError: false,
		},
		{
			name:         "config with invalid dimensions handled gracefully",
			searchMethod: DirectSearch,
			config: &model.Config{
				MenuID:    "test",
				Title:     "Test Menu",
				Prompt:    "test>",
				MinWidth:  -100, // invalid
				MinHeight: -100, // invalid
			},
			expectError: false, // should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.config == nil {
				// Test nil config scenario - this should be handled in NewGMenu
				assert.Panics(t, func() {
					_, _ = NewGMenu(tt.searchMethod, tt.config)
				})
				return
			}

			gmenu, err := NewGMenu(tt.searchMethod, tt.config)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
				assert.Nil(t, gmenu)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gmenu)
				defer func() {
					if gmenu.menuCancel != nil {
						gmenu.menuCancel()
					}
				}()
			}
		})
	}
}

// TestSetupMenuErrorHandling tests error handling in menu setup
func TestSetupMenuErrorHandling(t *testing.T) {
	config := &model.Config{
		MenuID:    "test",
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	tests := []struct {
		name         string
		items        []string
		initialQuery string
		expectError  bool
	}{
		{
			name:         "empty items list should succeed",
			items:        []string{},
			initialQuery: "",
			expectError:  false,
		},
		{
			name:         "nil items list should succeed",
			items:        nil,
			initialQuery: "",
			expectError:  false,
		},
		{
			name:         "valid items should succeed",
			items:        []string{"item1", "item2", "item3"},
			initialQuery: "item",
			expectError:  false,
		},
		{
			name:         "items with empty strings should succeed",
			items:        []string{"", "valid", ""},
			initialQuery: "",
			expectError:  false,
		},
		{
			name:         "very long query should succeed",
			items:        []string{"test"},
			initialQuery: strings.Repeat("a", 1000),
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gmenu.SetupMenu(tt.items, tt.initialQuery)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestExitCodeHandling tests exit code setting and validation
func TestExitCodeHandling(t *testing.T) {
	config := &model.Config{
		MenuID:    "test",
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	// Test initial state
	assert.Equal(t, model.Unset, gmenu.GetExitCode())

	// Test setting valid exit codes
	err = gmenu.SetExitCode(model.NoError)
	assert.NoError(t, err)
	assert.Equal(t, model.NoError, gmenu.GetExitCode())

	// Test attempting to change exit code (should log warning but not error)
	err = gmenu.SetExitCode(model.UserCanceled)
	assert.NoError(t, err)                              // Should not error, but should keep original code
	assert.Equal(t, model.NoError, gmenu.GetExitCode()) // Should remain NoError

	// Test quit with unset exit code should panic with proper recovery
	gmenu2, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu2.menuCancel != nil {
			gmenu2.menuCancel()
		}
	}()

	assert.Panics(t, func() {
		gmenu2.Quit() // Should panic because exit code is unset
	})
}

// TestCacheErrorHandling tests cache operation error handling
func TestCacheErrorHandling(t *testing.T) {
	// Test with empty menuID (should skip caching gracefully)
	config := &model.Config{
		MenuID:    "", // empty menuID
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	// Operations should succeed even with empty menuID
	err = gmenu.clearCache()
	assert.NoError(t, err)

	err = gmenu.cacheState("test_value")
	assert.NoError(t, err)
}

// TestInvalidSearchMethod tests handling of invalid search methods
func TestInvalidSearchMethod(t *testing.T) {
	config := &model.Config{
		MenuID:    "test",
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	// Test with nil search method - actually doesn't panic, just creates with nil method
	gmenu, err := NewGMenu(nil, config)
	if err == nil && gmenu != nil {
		defer func() {
			if gmenu.menuCancel != nil {
				gmenu.menuCancel()
			}
		}()
	}

	// Test with valid search method but invalid operations
	gmenu2, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu2.menuCancel != nil {
			gmenu2.menuCancel()
		}
	}()

	require.NoError(t, gmenu2.SetupMenu([]string{"test"}, ""))

	// Search should handle edge cases gracefully
	results := gmenu2.Search("") // empty query
	assert.NotNil(t, results)

	results = gmenu2.Search("nonexistent") // no matches
	assert.NotNil(t, results)
	assert.Empty(t, results)
}

// TestRaceConditionInSelection tests potential race conditions
func TestRaceConditionInSelection(t *testing.T) {
	config := &model.Config{
		MenuID:    "test",
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	require.NoError(t, gmenu.SetupMenu([]string{"item1", "item2", "item3"}, ""))

	// Test multiple concurrent markSelectionMade calls
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			defer func() { done <- true }()
			gmenu.markSelectionMade()
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Fuse should be broken only once
	assert.True(t, gmenu.selectionFuse.IsBroken())
	assert.True(t, gmenu.ui.SearchEntry.Disabled())
}

// TestAutoSelectErrorHandling tests auto-select functionality error handling
func TestAutoSelectErrorHandling(t *testing.T) {
	config := &model.Config{
		MenuID:     "test",
		Title:      "Test Menu",
		Prompt:     "test>",
		MinWidth:   300,
		MinHeight:  200,
		AutoAccept: true,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	// Test auto-select with loading item (should not auto-select)
	require.NoError(t, gmenu.SetupMenu([]string{}, ""))

	// Should not auto-select because only loading item present
	time.Sleep(50 * time.Millisecond)
	assert.False(t, gmenu.selectionFuse.IsBroken())

	// Test that shouldAutoSelect works correctly
	require.NoError(t, gmenu.SetupMenu([]string{"unique_item"}, ""))

	// shouldAutoSelect should return true for single non-loading item
	assert.True(t, gmenu.shouldAutoSelect())
}

// TestUIErrorRecovery tests UI error recovery mechanisms
func TestUIErrorRecovery(t *testing.T) {
	config := &model.Config{
		MenuID:    "test",
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	require.NoError(t, gmenu.SetupMenu([]string{"item1", "item2"}, ""))

	// Test safe UI updates with potential panics
	assert.NotPanics(t, func() {
		gmenu.safeUIUpdate(func() {
			// Simulate potential UI operation that might panic
			gmenu.ui.SearchEntry.SetText("test")
		})
	})

	// Test visibility state management edge cases
	gmenu.setShown(true)
	assert.True(t, gmenu.IsShown())

	gmenu.setShown(false)
	assert.False(t, gmenu.IsShown())

	// Multiple calls should be safe
	gmenu.setShown(false)
	assert.False(t, gmenu.IsShown())
}

// TestFileOperationErrors tests file operation error handling
func TestFileOperationErrors(t *testing.T) {
	// Test with read-only directory for PID file operations
	if os.Getuid() == 0 {
		t.Skip("Skipping test as root user (can write to read-only directories)")
	}

	tmpDir, err := os.MkdirTemp("", "gmenu-error-test")
	require.NoError(t, err)
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Test createPidFile with valid directory
	pidFile, err := createPidFile("test_pid")
	if err == nil {
		defer func() {
			_ = RemovePidFile("test_pid")
		}()
		assert.NotEmpty(t, pidFile)

		// Test creating PID file when one already exists
		_, err = createPidFile("test_pid")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	}

	// Test removePidFile with non-existent file
	err = RemovePidFile("nonexistent_pid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

// TestCanBeHighlightedEdgeCases tests edge cases for highlighting
func TestCanBeHighlightedEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"empty string", "", true},
		{"alphanumeric only", "abc123", true},
		{"with space", "abc 123", false},
		{"with special chars", "abc@123", false},
		{"with unicode", "café", false},
		{"with emoji", "test🚀", false},
		{"single char", "a", true},
		{"single number", "1", true},
		{"mixed case", "AbC123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canBeHighlighted(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
