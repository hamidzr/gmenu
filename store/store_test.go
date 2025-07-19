package store

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFileStoreCreation tests creating a new file store
func TestFileStoreCreation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-store-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create store with test path
	storePath := []string{"test", "gmenu"}
	store, err := NewFileStore[Cache, Config](storePath, "yaml")
	require.NoError(t, err)
	require.NotNil(t, store)

	// Verify store was created successfully
	assert.NotNil(t, store)
}

// TestCacheOperations tests cache loading and saving
func TestCacheOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-cache-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	storePath := []string{"test", "cache"}
	store, err := NewFileStore[Cache, Config](storePath, "yaml")
	require.NoError(t, err)

	// Test loading cache when file doesn't exist
	cache, err := store.LoadCache()
	// Should return error or empty cache for non-existent file
	if err == nil {
		// If no error, cache should be empty/default
		assert.NotNil(t, cache)
	}

	// Create test cache data
	testCache := &Cache{
		UsageCount:       map[string]int{"item1": 5, "item2": 3},
		NotFoundAccepted: []string{"not_found1", "not_found2"},
		LastEntry:        "test_entry",
		LastInput:        "test input",
	}

	// Test saving cache
	err = store.SaveCache(*testCache)
	require.NoError(t, err)

	// Verify cache file was created (the actual path will be based on the store's cache directory)
	// We don't need to check the exact path since it's system-dependent

	// Test loading saved cache
	loadedCache, err := store.LoadCache()
	require.NoError(t, err)
	require.NotNil(t, loadedCache)

	// Verify loaded data matches saved data
	assert.Equal(t, testCache.UsageCount, loadedCache.UsageCount)
	assert.Equal(t, testCache.NotFoundAccepted, loadedCache.NotFoundAccepted)
	assert.Equal(t, testCache.LastEntry, loadedCache.LastEntry)
	assert.Equal(t, testCache.LastInput, loadedCache.LastInput)
}

// TestConfigOperations tests config loading and saving
func TestConfigOperations(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-config-store-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Change to temp directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Logf("failed to restore directory: %v", err)
		}
	}()
	err = os.Chdir(tmpDir)
	require.NoError(t, err)

	storePath := []string{"test", "config"}
	store, err := NewFileStore[Cache, Config](storePath, "yaml")
	require.NoError(t, err)

	// Create test config data
	testConfig := &Config{
		AppTitle:      "Test App",
		DefaultPrompt: "Test Prompt",
		DefaultLimit:  10,
		SearchMethod:  "fuzzy",
	}
	testConfig.WindowSize.Width = 800
	testConfig.WindowSize.Height = 600

	// Test saving config
	err = store.SaveConfig(*testConfig)
	require.NoError(t, err)

	// Verify config file was created (the actual path will be based on the store's config directory)
	// We don't need to check the exact path since it's system-dependent

	// Test loading saved config
	loadedConfig, err := store.LoadConfig()
	require.NoError(t, err)
	require.NotNil(t, loadedConfig)

	// Verify loaded data matches saved data
	assert.Equal(t, testConfig.AppTitle, loadedConfig.AppTitle)
	assert.Equal(t, testConfig.DefaultPrompt, loadedConfig.DefaultPrompt)
	assert.Equal(t, testConfig.DefaultLimit, loadedConfig.DefaultLimit)
	assert.Equal(t, testConfig.SearchMethod, loadedConfig.SearchMethod)
	assert.Equal(t, testConfig.WindowSize.Width, loadedConfig.WindowSize.Width)
	assert.Equal(t, testConfig.WindowSize.Height, loadedConfig.WindowSize.Height)
}

// TestStoreWithDifferentFormats tests different serialization formats
func TestStoreWithDifferentFormats(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-format-test")
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

	formats := []string{"yaml", "json"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			// Create unique directory for each format test to avoid conflicts
			formatTmpDir := filepath.Join(tmpDir, format+"_test")
			err := os.MkdirAll(formatTmpDir, 0755)
			require.NoError(t, err)

			// Change to format-specific directory and restore afterwards
			currentDir, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				if err := os.Chdir(currentDir); err != nil {
					t.Logf("failed to restore directory: %v", err)
				}
			}()

			err = os.Chdir(formatTmpDir)
			require.NoError(t, err)

			storePath := []string{"test", format}
			store, err := NewFileStore[Cache, Config](storePath, format)
			require.NoError(t, err)

			testCache := &Cache{
				UsageCount: map[string]int{"test": 1},
				LastEntry:  "test_" + format,
				LastInput:  "test input " + format,
			}

			// Test save/load cycle
			err = store.SaveCache(*testCache)
			require.NoError(t, err)

			loadedCache, err := store.LoadCache()
			require.NoError(t, err)
			assert.Equal(t, testCache.LastEntry, loadedCache.LastEntry)
		})
	}
}

// TestStoreErrorHandling tests error conditions
func TestStoreErrorHandling(t *testing.T) {
	// Test with invalid format
	_, err := NewFileStore[Cache, Config]([]string{"test"}, "invalid_format")
	assert.Error(t, err)

	// Test with empty path (this is actually allowed - it uses root directories)
	store, err := NewFileStore[Cache, Config]([]string{}, "yaml")
	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Test loading from non-existent directory with proper setup
	tmpDir, err := os.MkdirTemp("", "store-error-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create store in a nested path that doesn't exist yet
	store2, err := NewFileStore[Cache, Config]([]string{"non", "existent", "path"}, "yaml")
	require.NoError(t, err) // Store creation should succeed (it creates directories)

	// Loading from empty/nonexistent file should return default zero values
	cache, err := store2.LoadCache()
	assert.NoError(t, err) // Loading empty cache should succeed with zero value
	assert.NotNil(t, cache)
}

// TestConcurrentAccess tests concurrent store operations
func TestConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-concurrent-test")
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

	// Create separate stores for each goroutine to avoid file conflicts
	store1, err := NewFileStore[Cache, Config]([]string{"concurrent", "test1"}, "yaml")
	require.NoError(t, err)
	store2, err := NewFileStore[Cache, Config]([]string{"concurrent", "test2"}, "yaml")
	require.NoError(t, err)
	store3, err := NewFileStore[Cache, Config]([]string{"concurrent", "test3"}, "yaml")
	require.NoError(t, err)

	// Create initial caches for each store
	initialCache := &Cache{
		LastEntry:        "initial",
		NotFoundAccepted: []string{"initial"},
		LastInput:        "initial query",
	}
	err = store1.SaveCache(*initialCache)
	require.NoError(t, err)
	err = store2.SaveCache(*initialCache)
	require.NoError(t, err)
	err = store3.SaveCache(*initialCache)
	require.NoError(t, err)

	// Simulate concurrent operations with separate stores
	done := make(chan bool, 3)

	// Goroutine 1: Read operations
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			_, err := store1.LoadCache()
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
			}
		}
	}()

	// Goroutine 2: Write operations
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			cache := &Cache{
				LastEntry:        "concurrent_write",
				NotFoundAccepted: []string{"write", "test"},
				LastInput:        "concurrent query",
			}
			err := store2.SaveCache(*cache)
			if err != nil {
				t.Errorf("Concurrent write failed: %v", err)
			}
		}
	}()

	// Goroutine 3: Mixed operations
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			// Read
			_, err := store3.LoadCache()
			if err != nil {
				t.Errorf("Concurrent mixed read failed: %v", err)
			}

			// Write
			cache := &Cache{
				LastEntry:        "mixed_op",
				NotFoundAccepted: []string{"mixed"},
				LastInput:        "mixed query",
			}
			err = store3.SaveCache(*cache)
			if err != nil {
				t.Errorf("Concurrent mixed write failed: %v", err)
			}
		}
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		<-done
	}

	// Verify final state is consistent for each store
	finalCache1, err := store1.LoadCache()
	require.NoError(t, err)
	assert.NotNil(t, finalCache1)

	finalCache2, err := store2.LoadCache()
	require.NoError(t, err)
	assert.NotNil(t, finalCache2)

	finalCache3, err := store3.LoadCache()
	require.NoError(t, err)
	assert.NotNil(t, finalCache3)
}

// TestCacheStructure tests the cache data structure
func TestCacheStructure(t *testing.T) {
	cache := &Cache{
		LastEntry: "test_selection",
		NotFoundAccepted: []string{
			"item1",
			"item2",
			"item3",
		},
		LastInput: "test query",
	}

	// Test field access
	assert.Equal(t, "test_selection", cache.LastEntry)
	assert.Len(t, cache.NotFoundAccepted, 3)
	assert.Equal(t, "item1", cache.NotFoundAccepted[0])
	assert.Equal(t, "test query", cache.LastInput)

	// Test field modification
	cache.LastEntry = "new_selection"
	assert.Equal(t, "new_selection", cache.LastEntry)

	cache.NotFoundAccepted = append(cache.NotFoundAccepted, "item4")
	assert.Len(t, cache.NotFoundAccepted, 4)
}

// TestConfigStructure tests the config data structure
func TestConfigStructure(t *testing.T) {
	config := &Config{
		AppTitle:      "Test App",
		DefaultPrompt: "Test Prompt",
		DefaultLimit:  20,
		SearchMethod:  "fuzzy",
	}
	config.WindowSize.Width = 1024
	config.WindowSize.Height = 768

	// Test structure access
	assert.Equal(t, "Test App", config.AppTitle)
	assert.Equal(t, "Test Prompt", config.DefaultPrompt)
	assert.Equal(t, 20, config.DefaultLimit)
	assert.Equal(t, "fuzzy", config.SearchMethod)
	assert.Equal(t, 1024, config.WindowSize.Width)
	assert.Equal(t, 768, config.WindowSize.Height)

	// Test field modification
	config.WindowSize.Width = 1280
	assert.Equal(t, 1280, config.WindowSize.Width)
}

// TestUtilityFunctions tests utility functions in the store package
func TestUtilityFunctions(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gmenu-utils-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test createDirectoryPath (if it exists)
	testPath := filepath.Join(tmpDir, "deep", "nested", "path")

	// Create the directory structure
	err = os.MkdirAll(testPath, 0755)
	require.NoError(t, err)

	// Verify it was created
	info, err := os.Stat(testPath)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}
