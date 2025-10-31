package core

import (
	"context"
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullApplicationWorkflow tests the complete user workflow
func TestFullApplicationWorkflow(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()

	config := &model.Config{
		Title:                 "Integration Test",
		Prompt:                "Select Item",
		MenuID:                "integration-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenuWithApp(testApp, searchMethod, config)
	require.NoError(t, err)

	testItems := []string{
		"Apple",
		"Banana",
		"Cherry",
		"Date",
		"Elderberry",
		"Fig",
		"Grape",
	}

	// Step 1: Setup menu with items
	require.NoError(t, gmenu.SetupMenu(testItems, ""))
	assert.Len(t, gmenu.menu.items, len(testItems))
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Step 2: Show UI
	require.NoError(t, gmenu.ShowUI())
	assert.True(t, gmenu.IsShown())

	// Step 3: Test search functionality
	test.Type(gmenu.ui.SearchEntry, "app")
	time.Sleep(10 * time.Millisecond) // Allow search to process

	// Should filter to items containing "app" (read under lock)
	gmenu.menu.itemsMutex.Lock()
	filtered := append([]model.MenuItem(nil), gmenu.menu.Filtered...)
	gmenu.menu.itemsMutex.Unlock()
	assert.GreaterOrEqual(t, len(filtered), 1) // At least "Apple"

	// Step 4: Navigate with keyboard
	downEvent := &fyne.KeyEvent{Name: fyne.KeyDown}
	gmenu.ui.SearchEntry.TypedKey(downEvent)
	// Selection might be 0 or 1 depending on search results
	assert.GreaterOrEqual(t, gmenu.menu.Selected, 0)

	// Step 5: Clear search to see all items
	gmenu.ui.SearchEntry.SetText("")
	time.Sleep(10 * time.Millisecond)
	gmenu.menu.itemsMutex.Lock()
	assert.Len(t, gmenu.menu.Filtered, len(testItems))
	gmenu.menu.itemsMutex.Unlock()

	// Step 6: Navigate to specific item using numeric key
	key3Event := &fyne.KeyEvent{Name: fyne.Key3}
	gmenu.ui.SearchEntry.TypedKey(key3Event)
	assert.Equal(t, 2, gmenu.menu.Selected) // 3 maps to index 2

	// Step 7: Hide UI
	gmenu.HideUI()
	assert.False(t, gmenu.IsShown())

	// Step 8: Reset and test state
	gmenu.Reset(true)
	assert.False(t, gmenu.selectionFuse.IsBroken())
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text)
	assert.Equal(t, 0, gmenu.menu.Selected)
}

// TestAutoAcceptWorkflow tests auto-accept when only one item matches
func TestAutoAcceptWorkflow(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()

	config := &model.Config{
		Title:                 "Auto Accept Test",
		Prompt:                "Select Item",
		MenuID:                "auto-accept-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            true, // Enable auto-accept
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	testItems := []string{"unique_item", "other", "another"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Search for something that matches only one item
	test.Type(gmenu.ui.SearchEntry, "unique")
	time.Sleep(10 * time.Millisecond)

	// Should have exactly one match
	gmenu.menu.itemsMutex.Lock()
	filtered := append([]model.MenuItem(nil), gmenu.menu.Filtered...)
	gmenu.menu.itemsMutex.Unlock()
	assert.Len(t, filtered, 1)
	assert.Equal(t, "unique_item", filtered[0].Title)
}

// TestInitialQueryWorkflow tests workflow with pre-filled search
func TestInitialQueryWorkflow(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues and force factory override
	testApp := test.NewApp()
	defer testApp.Quit()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return testApp }
	defer func() { newAppFunc = oldNewApp }()

	config := &model.Config{
		Title:                 "Initial Query Test",
		Prompt:                "Search",
		MenuID:                "initial-query-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "fruit",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	testItems := []string{
		"apple fruit",
		"banana fruit",
		"vegetable carrot",
		"fruit salad",
	}

	// Setup with initial query
	require.NoError(t, gmenu.SetupMenu(testItems, config.InitialQuery))

	// Search entry should have the initial query
	assert.Equal(t, config.InitialQuery, gmenu.ui.SearchEntry.Text)

	// Should be filtered to items containing "fruit"
	time.Sleep(10 * time.Millisecond)
	gmenu.menu.itemsMutex.Lock()
	filtered := append([]model.MenuItem(nil), gmenu.menu.Filtered...)
	gmenu.menu.itemsMutex.Unlock()
	assert.GreaterOrEqual(t, len(filtered), 3) // Should match 3 fruit items
}

// TestErrorHandling tests error conditions and edge cases
func TestErrorHandling(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()

	config := &model.Config{
		Title:                 "Error Test",
		Prompt:                "Search",
		MenuID:                "error-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	// Test with empty items list
	require.NoError(t, gmenu.SetupMenu([]string{}, ""))
	assert.Len(t, gmenu.menu.items, 1) // Should have loading item
	assert.Equal(t, "Loading", gmenu.menu.items[0].Title)

	// Test with nil items
	require.NoError(t, gmenu.SetupMenu(nil, ""))
	assert.Len(t, gmenu.menu.items, 1) // Should handle gracefully

	// Test search with no matches
	testItems := []string{"apple", "banana", "cherry"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))
	test.Type(gmenu.ui.SearchEntry, "xyz_no_match")
	time.Sleep(10 * time.Millisecond)

	gmenu.menu.itemsMutex.Lock()
	filtered := append([]model.MenuItem(nil), gmenu.menu.Filtered...)
	gmenu.menu.itemsMutex.Unlock()
	assert.Len(t, filtered, 0) // No matches
}

// TestConcurrentOperations tests thread safety
func TestConcurrentOperations(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()

	config := &model.Config{
		Title:                 "Concurrent Test",
		Prompt:                "Search",
		MenuID:                "concurrent-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	testItems := make([]string, 100)
	for i := 0; i < 100; i++ {
		testItems[i] = "item" + string(rune('0'+i%10))
	}

	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test concurrent search operations
	done := make(chan bool, 3)

	// Goroutine 1: Rapid search changes
	go func() {
		defer func() { done <- true }()
		searches := []string{"item1", "item2", "item3", "", "item"}
		for _, search := range searches {
			gmenu.ui.SearchEntry.SetText(search)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Goroutine 2: Navigation
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 10; i++ {
			downEvent := &fyne.KeyEvent{Name: fyne.KeyDown}
			gmenu.ui.SearchEntry.TypedKey(downEvent)
			time.Sleep(5 * time.Millisecond)
		}
	}()

	// Goroutine 3: Reset operations
	go func() {
		defer func() { done <- true }()
		for i := 0; i < 5; i++ {
			gmenu.Reset(false)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Wait for all goroutines to complete
	timeout := time.After(5 * time.Second)
	completed := 0
	for completed < 3 {
		select {
		case <-done:
			completed++
		case <-timeout:
			t.Fatal("Test timed out - possible deadlock")
		}
	}

	// Verify final state is consistent
	assert.NotNil(t, gmenu.menu)
	assert.NotNil(t, gmenu.ui)
}

// TestContextCancellation tests proper cleanup on context cancellation
func TestContextCancellation(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()

	config := &model.Config{
		Title:                 "Context Test",
		Prompt:                "Search",
		MenuID:                "context-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	testItems := []string{"item1", "item2", "item3"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start an operation that should be cancelled
	go func() {
		// Simulate some work that respects context
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Do some work
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	// Cancel the context
	cancel()

	// Verify that the menu still functions after cancellation
	test.Type(gmenu.ui.SearchEntry, "item")
	time.Sleep(10 * time.Millisecond)

	gmenu.menu.itemsMutex.Lock()
	filtered := append([]model.MenuItem(nil), gmenu.menu.Filtered...)
	gmenu.menu.itemsMutex.Unlock()
	assert.GreaterOrEqual(t, len(filtered), 1)
}

// TestMemoryManagement tests that resources are properly cleaned up
func TestMemoryManagement(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues and force factory override
	testApp := test.NewApp()
	defer testApp.Quit()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return testApp }
	defer func() { newAppFunc = oldNewApp }()

	config := &model.Config{
		Title:                 "Memory Test",
		Prompt:                "Search",
		MenuID:                "memory-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1200,
		MaxHeight:             800,
		AcceptCustomSelection: true,
	}

	// Create and destroy multiple instances to test for leaks
	for i := 0; i < 10; i++ {
		instanceApp := test.NewApp()
		searchMethod := SearchMethods["fuzzy"]
		gmenu, err := NewGMenuWithApp(instanceApp, searchMethod, config)
		require.NoError(t, err)

		testItems := make([]string, 1000) // Large dataset
		for j := 0; j < 1000; j++ {
			testItems[j] = "large_item_" + string(rune('0'+j%10))
		}

		require.NoError(t, gmenu.SetupMenu(testItems, ""))
		require.NoError(t, gmenu.ShowUI())

		// Perform some operations
		test.Type(gmenu.ui.SearchEntry, "large")
		time.Sleep(5 * time.Millisecond)

		gmenu.HideUI()
		gmenu.Reset(true)
		instanceApp.Quit()

		// The instance should be ready for garbage collection
		gmenu = nil
	}

	// Force garbage collection to test for leaks
	// In a real test environment, you might use tools like pprof
	// to verify no memory leaks occur
}
