package core

import (
	"context"
	"sync"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWaitForSelectionTimeout ensures WaitForSelection doesn't hang indefinitely
func TestWaitForSelectionTimeout(t *testing.T) {
	config := &model.Config{
		MenuID:    "timeout-test",
		Title:     "Timeout Test",
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

	// Test WaitForSelection with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)

		// Start waiting for selection
		go func() {
			gmenu.WaitForSelection()
		}()

		// Simulate selection after delay
		time.Sleep(100 * time.Millisecond)
		gmenu.markSelectionMade()
	}()

	// Ensure operation completes within timeout
	select {
	case <-done:
		// Test passed - operation completed
	case <-time.After(2 * time.Second):
		t.Fatal("WaitForSelection timed out - possible hang")
	}

	assert.True(t, gmenu.selectionFuse.IsBroken())
}

// TestRunAppForeverDoesNotHang tests that RunAppForever can be properly terminated
func TestRunAppForeverDoesNotHang(t *testing.T) {
	app := test.NewApp()

	config := &model.Config{
		MenuID:    "app-forever-test",
		Title:     "App Forever Test",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenuWithApp(app, DirectSearch, config)
	require.NoError(t, err)

	require.NoError(t, gmenu.SetupMenu([]string{"item1", "item2"}, ""))

	// Test RunAppForever with proper termination
	done := make(chan error, 1)
	go func() {
		done <- gmenu.RunAppForever()
	}()

	// Allow some time for the app to start
	time.Sleep(50 * time.Millisecond)

	// Quit the app to terminate RunAppForever
	app.Quit()

	// Ensure operation completes within timeout
	select {
	case err := <-done:
		// Should complete without error when app is quit
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("RunAppForever timed out - possible hang")
	}
}

// TestHideShowSelectionCycle tests the specific scenario where menu goes hidden,
// comes up with new items, selection is made, then goes hidden again
func TestHideShowSelectionCycle(t *testing.T) {
	config := &model.Config{
		MenuID:    "cycle-test",
		Title:     "Cycle Test",
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

	// Initial setup
	require.NoError(t, gmenu.SetupMenu([]string{"initial1", "initial2"}, ""))

	cycleComplete := make(chan struct{})

	go func() {
		defer close(cycleComplete)

		for cycle := 0; cycle < 5; cycle++ {
			// Step 1: Show menu
			gmenu.ShowUI()
			assert.True(t, gmenu.IsShown(), "Menu should be shown in cycle %d", cycle)

			// Step 2: Update with new items
			newItems := []string{
				"cycle" + string(rune('0'+cycle)) + "_item1",
				"cycle" + string(rune('0'+cycle)) + "_item2",
				"cycle" + string(rune('0'+cycle)) + "_item3",
			}
			require.NoError(t, gmenu.SetupMenu(newItems, ""))

			// Step 3: Verify items were updated
			assert.Equal(t, len(newItems), len(gmenu.menu.items))

			// Step 4: Search and make selection
			gmenu.Search("cycle" + string(rune('0'+cycle)))

			// Simulate user selection
			gmenu.markSelectionMade()
			assert.True(t, gmenu.selectionFuse.IsBroken())

			// Step 5: Hide menu
			gmenu.HideUI()
			assert.False(t, gmenu.IsShown(), "Menu should be hidden in cycle %d", cycle)

			// Step 6: Reset for next cycle
			gmenu.Reset(true)
			assert.False(t, gmenu.selectionFuse.IsBroken())

			// Small delay between cycles
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Ensure the entire cycle completes within timeout
	select {
	case <-cycleComplete:
		// Test passed - all cycles completed
	case <-time.After(5 * time.Second):
		t.Fatal("Hide/show/selection cycle timed out - possible hang")
	}

	// Verify final state
	assert.False(t, gmenu.IsShown())
	assert.False(t, gmenu.selectionFuse.IsBroken())
}

// TestConcurrentHideShowSelection tests concurrent hide/show/selection operations
func TestConcurrentHideShowSelection(t *testing.T) {
	config := &model.Config{
		MenuID:    "concurrent-cycle-test",
		Title:     "Concurrent Cycle Test",
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

	var wg sync.WaitGroup
	numGoroutines := 5
	iterations := 20

	completed := make(chan struct{})

	go func() {
		defer close(completed)

		// Multiple goroutines performing hide/show cycles
		wg.Add(numGoroutines * 3)

		// Goroutines for show/hide operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					gmenu.ShowUI()
					time.Sleep(time.Microsecond)
					gmenu.HideUI()
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Goroutines for selection operations
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					gmenu.markSelectionMade()
					time.Sleep(time.Microsecond)
					gmenu.Reset(false) // Reset without affecting visibility
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		// Goroutines for item updates
		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer wg.Done()
				for j := 0; j < iterations; j++ {
					newItems := []string{
						"concurrent_item_" + string(rune('0'+id)) + "_" + string(rune('0'+j)),
					}
					gmenu.SetupMenu(newItems, "")
					time.Sleep(time.Microsecond)
				}
			}(i)
		}

		wg.Wait()
	}()

	// Ensure all concurrent operations complete within timeout
	select {
	case <-completed:
		// Test passed - all operations completed
	case <-time.After(10 * time.Second):
		t.Fatal("Concurrent hide/show/selection operations timed out - possible deadlock")
	}

	// Verify final state is consistent
	assert.NotNil(t, gmenu.menu.items)
	assert.GreaterOrEqual(t, len(gmenu.menu.items), 1)
}

// TestLongRunningOperationsWithTimeout tests operations that might take a long time
func TestLongRunningOperationsWithTimeout(t *testing.T) {
	config := &model.Config{
		MenuID:    "long-running-test",
		Title:     "Long Running Test",
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

	// Create a large dataset
	largeItems := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		largeItems[i] = "large_dataset_item_" + string(rune('0'+i%10)) + "_" + string(rune('a'+i%26))
	}

	done := make(chan struct{})

	go func() {
		defer close(done)

		// Setup large dataset
		require.NoError(t, gmenu.SetupMenu(largeItems, ""))

		// Perform multiple operations that might be slow
		for i := 0; i < 10; i++ {
			// Search operations
			gmenu.Search("large_dataset")
			gmenu.Search("item_5")
			gmenu.Search("")

			// UI operations
			gmenu.ShowUI()
			gmenu.HideUI()

			// State changes
			gmenu.markSelectionMade()
			gmenu.Reset(true)
		}
	}()

	// Ensure operations complete within reasonable time
	select {
	case <-done:
		// Test passed - operations completed
	case <-time.After(15 * time.Second):
		t.Fatal("Long running operations timed out - possible performance issue")
	}
}

// TestResourceCleanupPreventsHangs tests that resources are properly cleaned up
func TestResourceCleanupPreventsHangs(t *testing.T) {
	// Test multiple instances to check for resource leaks that might cause hangs
	for instance := 0; instance < 5; instance++ {
		func() {
			config := &model.Config{
				MenuID:    "cleanup-test-" + string(rune('0'+instance)),
				Title:     "Cleanup Test",
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

			// Perform operations that allocate resources
			gmenu.ShowUI()
			gmenu.Search("item")
			gmenu.markSelectionMade()
			gmenu.HideUI()
			gmenu.Reset(true)

			// Resources should be cleaned up when function exits
		}()
	}
}

// TestContextCancellationPreventsHangs tests that context cancellation prevents hangs
func TestContextCancellationPreventsHangs(t *testing.T) {
	config := &model.Config{
		MenuID:    "cancellation-test",
		Title:     "Cancellation Test",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)

	require.NoError(t, gmenu.SetupMenu([]string{"item1", "item2"}, ""))

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan struct{})

	go func() {
		defer close(done)

		// Start operations that should be cancelled
		for {
			select {
			case <-ctx.Done():
				return
			default:
				gmenu.ShowUI()
				gmenu.Search("test")
				gmenu.HideUI()
				time.Sleep(time.Microsecond)
			}
		}
	}()

	// Cancel original menu context to test cleanup
	if gmenu.menuCancel != nil {
		gmenu.menuCancel()
	}

	// Wait for context timeout or completion
	select {
	case <-done:
		// Operations stopped due to context cancellation
	case <-time.After(2 * time.Second):
		t.Fatal("Context cancellation did not prevent hang")
	}
}

// TestMutexDeadlockPrevention tests that mutex usage doesn't cause deadlocks
func TestMutexDeadlockPrevention(t *testing.T) {
	config := &model.Config{
		MenuID:    "deadlock-test",
		Title:     "Deadlock Test",
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

	completed := make(chan struct{})

	go func() {
		defer close(completed)

		// Multiple goroutines accessing different mutexes in different orders
		wg.Add(numGoroutines * 2)

		// Goroutines accessing visibility mutex then UI mutex
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					gmenu.ShowUI()              // Uses visibility mutex
					gmenu.safeUIUpdate(func() { // Uses UI mutex
						// UI operation
					})
					gmenu.HideUI() // Uses visibility mutex
				}
			}()
		}

		// Goroutines accessing UI mutex then visibility mutex
		for i := 0; i < numGoroutines; i++ {
			go func() {
				defer wg.Done()
				for j := 0; j < 50; j++ {
					gmenu.safeUIUpdate(func() { // Uses UI mutex
						// UI operation
					})
					_ = gmenu.IsShown() // Uses visibility mutex (read lock)
				}
			}()
		}

		wg.Wait()
	}()

	// Ensure no deadlock occurs
	select {
	case <-completed:
		// Test passed - no deadlock
	case <-time.After(5 * time.Second):
		t.Fatal("Mutex operations timed out - possible deadlock")
	}
}

// TestRapidOperationsCycle tests rapid cycles that might cause race conditions
func TestRapidOperationsCycle(t *testing.T) {
	config := &model.Config{
		MenuID:    "rapid-test",
		Title:     "Rapid Test",
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

	done := make(chan struct{})

	go func() {
		defer close(done)

		// Perform rapid operations that might cause issues
		for i := 0; i < 100; i++ {
			// Rapid hide/show/selection cycle
			gmenu.ShowUI()
			gmenu.Search("item" + string(rune('1'+i%3)))
			gmenu.markSelectionMade()
			gmenu.HideUI()
			gmenu.Reset(true)

			// Update items
			newItems := []string{
				"rapid_item_" + string(rune('0'+i%10)),
				"rapid_item_" + string(rune('1'+i%10)),
			}
			gmenu.SetupMenu(newItems, "")
		}
	}()

	// Ensure rapid operations complete within timeout
	select {
	case <-done:
		// Test passed - rapid operations completed
	case <-time.After(10 * time.Second):
		t.Fatal("Rapid operations cycle timed out - possible hang or race condition")
	}

	// Verify final state
	assert.NotNil(t, gmenu.menu.items)
	assert.False(t, gmenu.selectionFuse.IsBroken())
}
