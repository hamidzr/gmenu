package core

import (
	"os"
	"testing"

	"github.com/hamidzr/gmenu/model"
	"github.com/hamidzr/gmenu/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheOperationsWithMenuID tests cache operations with different menu IDs
func TestCacheOperationsWithMenuID(t *testing.T) {
	// Create temporary directory for test
	tmpDir, err := os.MkdirTemp("", "gmenu-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	tests := []struct {
		name     string
		menuID   string
		hasCache bool
	}{
		{
			name:     "with menu ID",
			menuID:   "test_menu",
			hasCache: true,
		},
		{
			name:     "empty menu ID",
			menuID:   "",
			hasCache: false, // should skip caching
		},
		{
			name:     "menu ID with special characters",
			menuID:   "menu-with_special.chars",
			hasCache: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &model.Config{
				MenuID:    tt.menuID,
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

			// Test cache operations
			err = gmenu.cacheState("test_value")
			assert.NoError(t, err)

			err = gmenu.clearCache()
			assert.NoError(t, err)

			// All operations should succeed regardless of whether caching is active
		})
	}
}

// TestInitValueWithCache tests initial value computation with cache
func TestInitValueWithCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-initval-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config := &model.Config{
		MenuID:    "cache_test",
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
		initialQuery string
		cacheValue   string
		expected     string
	}{
		{
			name:         "empty query with empty cache",
			initialQuery: "",
			cacheValue:   "",
			expected:     "",
		},
		{
			name:         "empty query with cache value",
			initialQuery: "",
			cacheValue:   "cachedvalue", // no underscore to pass canBeHighlighted
			expected:     "cachedvalue",
		},
		{
			name:         "query overrides cache",
			initialQuery: "queryvalue",
			cacheValue:   "cachedvalue",
			expected:     "queryvalue",
		},
		{
			name:         "cache with non-alphanumeric chars",
			initialQuery: "",
			cacheValue:   "cache@value",
			expected:     "", // should be filtered out by canBeHighlighted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up cache if needed
			if tt.cacheValue != "" {
				err := gmenu.withCache(func(cache *store.Cache) error {
					cache.SetLastInput(tt.cacheValue)
					return nil
				})
				require.NoError(t, err)
			}

			// Get initial value
			initVal, err := gmenu.initValue(tt.initialQuery)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, initVal)

			// Clear cache for next test
			err = gmenu.clearCache()
			require.NoError(t, err)
		})
	}
}

// TestWithCacheErrorHandling tests withCache error handling
func TestWithCacheErrorHandling(t *testing.T) {
	config := &model.Config{
		MenuID:    "error_test",
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

	// Test error in operation function
	err = gmenu.withCache(func(cache *store.Cache) error {
		return assert.AnError
	})
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)

	// Test successful operation
	err = gmenu.withCache(func(cache *store.Cache) error {
		cache.SetLastInput("test")
		return nil
	})
	assert.NoError(t, err)
}

// TestCacheStateWithDifferentValues tests caching different state values
func TestCacheStateWithDifferentValues(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-cache-state-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config := &model.Config{
		MenuID:    "state_test",
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

	testValues := []string{
		"",
		"simple_value",
		"value with spaces",
		"value_with_underscores",
		"value-with-hyphens",
		"value.with.dots",
		"ValueWithCaps",
		"123numeric456",
		"🚀emoji_value",
		"very_long_" + string(make([]rune, 100)), // long value
	}

	for _, value := range testValues {
		t.Run("value_"+value[:min(10, len(value))], func(t *testing.T) {
			// Set query and cache state
			gmenu.menu.query = value
			err := gmenu.cacheState(value)
			assert.NoError(t, err)

			// Verify cache was set (by trying to read it back)
			cache, err := gmenu.store.LoadCache()
			assert.NoError(t, err)
			assert.Equal(t, value, cache.LastInput)
			assert.Equal(t, value, cache.LastEntry)
		})
	}
}

// TestConfigurationDimensions tests configuration dimension handling
func TestConfigurationDimensions(t *testing.T) {
	tests := []struct {
		name      string
		config    *model.Config
		expectErr bool
	}{
		{
			name: "valid dimensions",
			config: &model.Config{
				MenuID:    "test",
				Title:     "Test",
				Prompt:    "test>",
				MinWidth:  300,
				MinHeight: 200,
				MaxWidth:  800,
				MaxHeight: 600,
			},
			expectErr: false,
		},
		{
			name: "zero dimensions",
			config: &model.Config{
				MenuID:    "test",
				Title:     "Test",
				Prompt:    "test>",
				MinWidth:  0,
				MinHeight: 0,
				MaxWidth:  0,
				MaxHeight: 0,
			},
			expectErr: false, // should handle gracefully
		},
		{
			name: "negative dimensions",
			config: &model.Config{
				MenuID:    "test",
				Title:     "Test",
				Prompt:    "test>",
				MinWidth:  -100,
				MinHeight: -100,
				MaxWidth:  -50,
				MaxHeight: -50,
			},
			expectErr: false, // should handle gracefully
		},
		{
			name: "min larger than max",
			config: &model.Config{
				MenuID:    "test",
				Title:     "Test",
				Prompt:    "test>",
				MinWidth:  800,
				MinHeight: 600,
				MaxWidth:  300,
				MaxHeight: 200,
			},
			expectErr: false, // should handle gracefully
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gmenu, err := NewGMenu(DirectSearch, tt.config)
			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, gmenu)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, gmenu)
				if gmenu != nil {
					defer func() {
						if gmenu.menuCancel != nil {
							gmenu.menuCancel()
						}
					}()

					// Verify dimensions are stored
					assert.Equal(t, tt.config.MinWidth, gmenu.dims.MinWidth)
					assert.Equal(t, tt.config.MinHeight, gmenu.dims.MinHeight)
					assert.Equal(t, tt.config.MaxWidth, gmenu.dims.MaxWidth)
					assert.Equal(t, tt.config.MaxHeight, gmenu.dims.MaxHeight)
				}
			}
		})
	}
}

// TestConfigurationFlags tests various configuration flags
func TestConfigurationFlags(t *testing.T) {
	baseConfig := &model.Config{
		MenuID:    "test",
		Title:     "Test Menu",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	tests := []struct {
		name   string
		modify func(*model.Config)
		test   func(*testing.T, *GMenu)
	}{
		{
			name: "auto accept enabled",
			modify: func(c *model.Config) {
				c.AutoAccept = true
			},
			test: func(t *testing.T, g *GMenu) {
				assert.True(t, g.config.AutoAccept)
			},
		},
		{
			name: "auto accept disabled",
			modify: func(c *model.Config) {
				c.AutoAccept = false
			},
			test: func(t *testing.T, g *GMenu) {
				assert.False(t, g.config.AutoAccept)
			},
		},
		{
			name: "numeric selection disabled",
			modify: func(c *model.Config) {
				c.NoNumericSelection = true
			},
			test: func(t *testing.T, g *GMenu) {
				assert.True(t, g.config.NoNumericSelection)
			},
		},
		{
			name: "custom selection enabled",
			modify: func(c *model.Config) {
				c.AcceptCustomSelection = true
			},
			test: func(t *testing.T, g *GMenu) {
				assert.True(t, g.config.AcceptCustomSelection)
			},
		},
		{
			name: "preserve order enabled",
			modify: func(c *model.Config) {
				c.PreserveOrder = true
			},
			test: func(t *testing.T, g *GMenu) {
				assert.True(t, g.preserveOrder)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create config copy
			config := *baseConfig
			tt.modify(&config)

			gmenu, err := NewGMenu(DirectSearch, &config)
			require.NoError(t, err)
			defer func() {
				if gmenu.menuCancel != nil {
					gmenu.menuCancel()
				}
			}()

			tt.test(t, gmenu)
		})
	}
}

// TestClearCacheFunction tests the clear cache functionality
func TestClearCacheFunction(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-clear-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	config := &model.Config{
		MenuID:    "clear_test",
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

	// Set some cache values
	gmenu.menu.query = "test_query"
	err = gmenu.cacheState("test_selection")
	require.NoError(t, err)

	// Verify cache has values
	cache, err := gmenu.store.LoadCache()
	require.NoError(t, err)
	assert.Equal(t, "test_query", cache.LastInput)
	assert.Equal(t, "test_selection", cache.LastEntry)

	// Clear cache
	err = gmenu.clearCache()
	assert.NoError(t, err)

	// Verify cache is cleared
	cache, err = gmenu.store.LoadCache()
	require.NoError(t, err)
	assert.Equal(t, "", cache.LastInput)
	assert.Equal(t, "", cache.LastEntry)
}

// TestCanBeHighlightedWithCacheValues tests canBeHighlighted with realistic cache values
func TestCanBeHighlightedWithCacheValues(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"simple alphanumeric", "abc123", true},
		{"with spaces", "abc 123", false},
		{"with underscore", "abc_123", false},
		{"with hyphen", "abc-123", false},
		{"with dot", "abc.123", false},
		{"with slash", "abc/123", false},
		{"with special chars", "abc@123", false},
		{"empty string", "", true},
		{"single char", "a", true},
		{"single number", "1", true},
		{"all caps", "ABC", true},
		{"mixed case", "AbC123", true},
		{"with unicode", "café", false},
		{"with emoji", "test🚀", false},
		{"typical file path", "/home/user/file.txt", false},
		{"typical command", "ls -la", false},
		{"url", "https://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canBeHighlighted(tt.value)
			assert.Equal(t, tt.expected, result,
				"canBeHighlighted(%q) = %v, want %v", tt.value, result, tt.expected)
		})
	}
}

// TestIsAlphaNumericHelper tests the isAlphaNumeric helper function
func TestIsAlphaNumericHelper(t *testing.T) {
	tests := []struct {
		name     string
		char     rune
		expected bool
	}{
		{"lowercase a", 'a', true},
		{"lowercase z", 'z', true},
		{"uppercase A", 'A', true},
		{"uppercase Z", 'Z', true},
		{"digit 0", '0', true},
		{"digit 9", '9', true},
		{"space", ' ', false},
		{"underscore", '_', false},
		{"hyphen", '-', false},
		{"dot", '.', false},
		{"at symbol", '@', false},
		{"unicode é", 'é', false},
		{"emoji", '🚀', false},
		{"newline", '\n', false},
		{"tab", '\t', false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAlphaNumeric(tt.char)
			assert.Equal(t, tt.expected, result,
				"isAlphaNumeric(%q) = %v, want %v", tt.char, result, tt.expected)
		})
	}
}
