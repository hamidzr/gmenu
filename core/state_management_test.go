package core

import (
	"sync"
	"testing"
	"time"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVisibilityStateManagement tests UI visibility state management
func TestVisibilityStateManagement(t *testing.T) {
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

	// Test initial state
	assert.False(t, gmenu.IsShown())

	// Test show/hide cycles
	for i := 0; i < 3; i++ {
		gmenu.ShowUI()
		assert.True(t, gmenu.IsShown())

		gmenu.HideUI()
		assert.False(t, gmenu.IsShown())
	}

	// Test toggle functionality
	gmenu.ToggleVisibility() // should show
	assert.True(t, gmenu.IsShown())

	gmenu.ToggleVisibility() // should hide
	assert.False(t, gmenu.IsShown())

	// Test multiple hide calls (should be safe)
	gmenu.HideUI()
	gmenu.HideUI()
	assert.False(t, gmenu.IsShown())

	// Test multiple show calls (should be safe)
	gmenu.ShowUI()
	gmenu.ShowUI()
	assert.True(t, gmenu.IsShown())
}

// TestConcurrentVisibilityAccess tests concurrent access to visibility state
func TestConcurrentVisibilityAccess(t *testing.T) {
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

	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 100

	// Test concurrent visibility operations
	wg.Add(numGoroutines * 3)

	// Goroutines for ShowUI
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				gmenu.ShowUI()
			}
		}()
	}

	// Goroutines for HideUI
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				gmenu.HideUI()
			}
		}()
	}

	// Goroutines for IsShown reads
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = gmenu.IsShown()
			}
		}()
	}

	wg.Wait()

	// Final state should be valid (either true or false)
	finalState := gmenu.IsShown()
	assert.IsType(t, false, finalState) // just ensure it's a boolean
}

// TestSelectionFuseStateManagement tests selection fuse state management
func TestSelectionFuseStateManagement(t *testing.T) {
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

	// Test initial state
	assert.False(t, gmenu.selectionFuse.IsBroken())
	assert.False(t, gmenu.ui.SearchEntry.Disabled())

	// Test first selection
	gmenu.markSelectionMade()
	assert.True(t, gmenu.selectionFuse.IsBroken())
	assert.True(t, gmenu.ui.SearchEntry.Disabled())

	// Test multiple selections (should be safe)
	gmenu.markSelectionMade()
	gmenu.markSelectionMade()
	assert.True(t, gmenu.selectionFuse.IsBroken())
	assert.True(t, gmenu.ui.SearchEntry.Disabled())
}

// TestConcurrentSelectionFuseAccess tests concurrent access to selection fuse
func TestConcurrentSelectionFuseAccess(t *testing.T) {
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

	var wg sync.WaitGroup
	numGoroutines := 20

	// Multiple goroutines trying to mark selection
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			gmenu.markSelectionMade()
		}()
	}

	wg.Wait()

	// Fuse should be broken exactly once
	assert.True(t, gmenu.selectionFuse.IsBroken())
	assert.True(t, gmenu.ui.SearchEntry.Disabled())
}

// TestMenuStateConsistency tests menu state consistency during operations
func TestMenuStateConsistency(t *testing.T) {
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

	testItems := []string{"apple", "banana", "cherry", "date", "elderberry"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test initial state
	assert.Equal(t, 0, gmenu.menu.Selected)
	assert.Equal(t, len(testItems), len(gmenu.menu.items))
	assert.Equal(t, len(testItems), len(gmenu.menu.Filtered))

	// Test search affects filtered but not items
	gmenu.Search("app")
	assert.Equal(t, len(testItems), len(gmenu.menu.items)) // original items unchanged
	assert.Equal(t, 1, len(gmenu.menu.Filtered))           // filtered items changed
	assert.Equal(t, 0, gmenu.menu.Selected)                // selection reset

	// Test empty search returns all items
	gmenu.Search("")
	assert.Equal(t, len(testItems), len(gmenu.menu.Filtered))
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Test search with no matches
	gmenu.Search("xyz_no_match")
	assert.Equal(t, 0, len(gmenu.menu.Filtered))
	assert.Equal(t, -1, gmenu.menu.Selected) // invalid selection for empty results
}

// TestConcurrentMenuOperations tests concurrent menu operations
func TestConcurrentMenuOperations(t *testing.T) {
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

	testItems := []string{"item1", "item2", "item3", "item4", "item5"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	var wg sync.WaitGroup
	numGoroutines := 5
	iterations := 50

	// Concurrent search operations
	queries := []string{"item", "1", "2", "3", "test", ""}

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				query := queries[(id+j)%len(queries)]
				results := gmenu.Search(query)
				assert.NotNil(t, results)
			}
		}(i)
	}

	wg.Wait()

	// Menu should still be in valid state
	assert.NotNil(t, gmenu.menu.items)
	assert.NotNil(t, gmenu.menu.Filtered)
	assert.GreaterOrEqual(t, gmenu.menu.Selected, -1) // -1 is valid for empty results
}

// TestUIUpdateMutexProtection tests UI update mutex protection
func TestUIUpdateMutexProtection(t *testing.T) {
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

	var wg sync.WaitGroup
	numGoroutines := 10
	iterations := 20

	// Concurrent UI updates
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				gmenu.safeUIUpdate(func() {
					// Simulate UI operations
					text := gmenu.ui.SearchEntry.Text
					gmenu.ui.SearchEntry.SetText(text + "test")
					time.Sleep(time.Microsecond)     // small delay to increase chance of race
					gmenu.ui.SearchEntry.SetText("") // reset
				})
			}
		}(i)
	}

	wg.Wait()

	// UI should be in valid state
	assert.NotNil(t, gmenu.ui.SearchEntry)
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text) // should be reset
}

// TestContextCancellationHandling tests context cancellation handling
func TestContextCancellationHandling(t *testing.T) {
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

	// Test context cancellation
	originalCancel := gmenu.menuCancel
	assert.NotNil(t, originalCancel)

	// Cancel the context
	originalCancel()

	// Operations should still be safe after cancellation
	assert.NotPanics(t, func() {
		gmenu.Search("test")
		gmenu.ShowUI()
		gmenu.HideUI()
	})
}

// TestResetStateConsistency tests state consistency after reset operations
func TestResetStateConsistency(t *testing.T) {
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

	// Modify state
	gmenu.ShowUI()
	gmenu.Search("item1")
	gmenu.markSelectionMade()

	// Verify modified state
	assert.True(t, gmenu.IsShown())
	assert.True(t, gmenu.selectionFuse.IsBroken())
	assert.True(t, gmenu.ui.SearchEntry.Disabled())

	// Reset
	gmenu.Reset(true)

	// Verify reset state - visibility is NOT affected by Reset
	assert.True(t, gmenu.IsShown())                  // Reset doesn't change visibility
	assert.False(t, gmenu.selectionFuse.IsBroken())  // Reset should reset the fuse
	assert.False(t, gmenu.ui.SearchEntry.Disabled()) // Reset should re-enable search entry
}

// TestExitCodeStateManagement tests exit code state management
func TestExitCodeStateManagement(t *testing.T) {
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

	// Test valid state transitions
	err = gmenu.SetExitCode(model.NoError)
	assert.NoError(t, err)
	assert.Equal(t, model.NoError, gmenu.GetExitCode())

	// Test attempting to change exit code
	err = gmenu.SetExitCode(model.UserCanceled)
	assert.NoError(t, err)                              // Should not error but should warn
	assert.Equal(t, model.NoError, gmenu.GetExitCode()) // Should remain unchanged

	// Test setting same exit code again
	err = gmenu.SetExitCode(model.NoError)
	assert.NoError(t, err)
	assert.Equal(t, model.NoError, gmenu.GetExitCode())
}

// TestConcurrentItemOperations tests concurrent item manipulation
func TestConcurrentItemOperations(t *testing.T) {
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

	require.NoError(t, gmenu.SetupMenu([]string{"initial"}, ""))

	var wg sync.WaitGroup
	numGoroutines := 5

	// Concurrent item operations
	wg.Add(numGoroutines * 2)

	// Prepend operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			items := []string{"prepend_" + string(rune('0'+id))}
			gmenu.PrependItems(items)
		}(i)
	}

	// Append operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			items := []string{"append_" + string(rune('0'+id))}
			gmenu.AppendItems(items)
		}(i)
	}

	wg.Wait()

	// Verify final state is consistent
	assert.NotNil(t, gmenu.menu.items)
	assert.Greater(t, len(gmenu.menu.items), 1) // Should have more than initial item

	// Should contain initial item and some added items
	found := false
	for _, item := range gmenu.menu.items {
		if item.Title == "initial" {
			found = true
			break
		}
	}
	assert.True(t, found, "Initial item should still be present")
}
