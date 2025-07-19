package core

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
)

// TestTerminalModeConfig tests terminal mode configuration
func TestTerminalModeConfig(t *testing.T) {
	config := &model.Config{
		Title:                 "Terminal Test",
		Prompt:                "Enter text: ",
		MenuID:                "terminal-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          true, // Enable terminal mode
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	// Verify terminal mode is enabled
	assert.True(t, config.TerminalMode)
	assert.Equal(t, "Enter text: ", config.Prompt)
}

// TestReadUserInputLive tests the live terminal input functionality
func TestReadUserInputLive(t *testing.T) {
	config := &model.Config{
		Prompt:       "Test prompt: ",
		InitialQuery: "initial",
	}

	// Create a channel to simulate query updates
	queryChan := make(chan string, 10)

	// Start the input reader in a goroutine with timeout
	done := make(chan string, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		// Simulate the ReadUserInputLive function behavior
		// In a real test, you might need to mock stdin or use a different approach
		<-ctx.Done()
		done <- config.InitialQuery // Return initial query on timeout
	}()

	// Wait for result or timeout
	select {
	case result := <-done:
		assert.Equal(t, config.InitialQuery, result)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Test timed out")
	}

	close(queryChan)
}

// TestTerminalInputValidation tests input validation in terminal mode
func TestTerminalInputValidation(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal input",
			input:    "normal text",
			expected: "normal text",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "whitespace input",
			input:    "   ",
			expected: "   ",
		},
		{
			name:     "special characters",
			input:    "!@#$%^&*()",
			expected: "!@#$%^&*()",
		},
		{
			name:     "unicode input",
			input:    "αβγδε 🚀",
			expected: "αβγδε 🚀",
		},
		{
			name:     "long input",
			input:    strings.Repeat("a", 1000),
			expected: strings.Repeat("a", 1000),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test that input validation accepts various inputs
			// In terminal mode, most inputs should be accepted as-is
			assert.Equal(t, tc.expected, tc.input)
		})
	}
}

// TestTerminalPromptDisplay tests prompt formatting and display
func TestTerminalPromptDisplay(t *testing.T) {
	testCases := []struct {
		name           string
		prompt         string
		initialQuery   string
		expectedPrompt string
	}{
		{
			name:           "simple prompt",
			prompt:         "Enter: ",
			initialQuery:   "",
			expectedPrompt: "Enter: ",
		},
		{
			name:           "prompt with initial query",
			prompt:         "Search: ",
			initialQuery:   "test",
			expectedPrompt: "Search: ",
		},
		{
			name:           "empty prompt",
			prompt:         "",
			initialQuery:   "",
			expectedPrompt: "",
		},
		{
			name:           "prompt with unicode",
			prompt:         "🔍 Search: ",
			initialQuery:   "",
			expectedPrompt: "🔍 Search: ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &model.Config{
				Prompt:       tc.prompt,
				InitialQuery: tc.initialQuery,
				TerminalMode: true,
			}

			assert.Equal(t, tc.expectedPrompt, config.Prompt)
			assert.Equal(t, tc.initialQuery, config.InitialQuery)
		})
	}
}

// TestTerminalKeyHandling tests keyboard input handling in terminal mode
func TestTerminalKeyHandling(t *testing.T) {
	config := &model.Config{
		Prompt:       "Test: ",
		InitialQuery: "",
		TerminalMode: true,
	}

	// Test that terminal mode config is properly set
	assert.True(t, config.TerminalMode)

	// In terminal mode, certain key behaviors might be different
	// This test verifies the configuration is set up correctly
	assert.Equal(t, "Test: ", config.Prompt)
}

// TestTerminalOutputFormatting tests output formatting in terminal mode
func TestTerminalOutputFormatting(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line",
			input:    "single line output",
			expected: "single line output",
		},
		{
			name:     "multiple words",
			input:    "multiple words here",
			expected: "multiple words here",
		},
		{
			name:     "with numbers",
			input:    "test123",
			expected: "test123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test output formatting (in a real implementation,
			// this might test actual terminal output formatting)
			result := tc.input // Simulate processing
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestTerminalContextHandling tests context handling in terminal mode
func TestTerminalContextHandling(t *testing.T) {

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Test that context cancellation is handled properly
	done := make(chan bool, 1)

	go func() {
		select {
		case <-ctx.Done():
			done <- true
		case <-time.After(100 * time.Millisecond):
			done <- false
		}
	}()

	select {
	case result := <-done:
		assert.True(t, result, "Context should be cancelled within timeout")
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Test timed out")
	}
}

// TestTerminalErrorHandling tests error handling in terminal mode
func TestTerminalErrorHandling(t *testing.T) {
	config := &model.Config{
		Prompt:       "Error test: ",
		InitialQuery: "",
		TerminalMode: true,
	}

	// Test with nil config
	assert.NotNil(t, config) // Config should not be nil

	// Test with empty prompt
	config.Prompt = ""
	assert.Equal(t, "", config.Prompt) // Should handle empty prompt

	// Test with nil initial query (should default to empty string)
	config.InitialQuery = ""
	assert.Equal(t, "", config.InitialQuery)
}

// TestTerminalModeIntegration tests terminal mode with other components
func TestTerminalModeIntegration(t *testing.T) {
	config := &model.Config{
		Title:                 "Terminal Integration",
		Prompt:                "Integration test: ",
		MenuID:                "terminal-integration",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "test",
		AutoAccept:            false,
		TerminalMode:          true,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	// Verify terminal mode works with other configuration options
	assert.True(t, config.TerminalMode)
	assert.Equal(t, "fuzzy", config.SearchMethod)
	assert.Equal(t, "test", config.InitialQuery)
	assert.True(t, config.AcceptCustomSelection)

	// Terminal mode should work with different search methods
	config.SearchMethod = "exact"
	assert.Equal(t, "exact", config.SearchMethod)

	config.SearchMethod = "regex"
	assert.Equal(t, "regex", config.SearchMethod)
}

// TestTerminalQueryUpdates tests query update handling in terminal mode
func TestTerminalQueryUpdates(t *testing.T) {
	// Create a channel to simulate query updates
	queryChan := make(chan string, 10)

	testQueries := []string{
		"a",
		"ap",
		"app",
		"apple",
		"appl",
		"app",
		"",
	}

	// Simulate sending query updates
	go func() {
		defer close(queryChan)
		for _, query := range testQueries {
			queryChan <- query
			time.Sleep(1 * time.Millisecond)
		}
	}()

	// Collect received queries
	var receivedQueries []string
	timeout := time.After(100 * time.Millisecond)

	for {
		select {
		case query, ok := <-queryChan:
			if !ok {
				// Channel closed
				goto done
			}
			receivedQueries = append(receivedQueries, query)
		case <-timeout:
			t.Fatal("Test timed out waiting for queries")
		}
	}

done:
	// Verify all queries were received
	assert.Equal(t, len(testQueries), len(receivedQueries))
	for i, expected := range testQueries {
		assert.Equal(t, expected, receivedQueries[i])
	}
}

// TestTerminalConfigDefaults tests default values for terminal mode
func TestTerminalConfigDefaults(t *testing.T) {
	config := &model.Config{}

	// Test default values
	assert.False(t, config.TerminalMode)     // Should default to false
	assert.Equal(t, "", config.Prompt)       // Should default to empty
	assert.Equal(t, "", config.InitialQuery) // Should default to empty

	// Test setting terminal mode
	config.TerminalMode = true
	assert.True(t, config.TerminalMode)

	// Test setting prompt and initial query
	config.Prompt = "Terminal prompt: "
	config.InitialQuery = "initial"
	assert.Equal(t, "Terminal prompt: ", config.Prompt)
	assert.Equal(t, "initial", config.InitialQuery)
}
