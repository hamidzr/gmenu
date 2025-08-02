package core

import (
	"testing"
	"time"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestVisualGUIHideShowCycle tests the hide/show cycle with actual visible GUI
// Run this test manually to see the actual GUI behavior
// Usage: go test ./core -run TestVisualGUIHideShowCycle -v
func TestVisualGUIHideShowCycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual GUI test in short mode")
	}
	
	// Skip this test in automated test runs due to Fyne main goroutine requirement
	t.Skip("Skipping visual GUI test - requires manual execution with real GUI")

	config := &model.Config{
		MenuID:    "visual-test",
		Title:     "Visual GUI Test",
		Prompt:    "Select an item:",
		MinWidth:  400,
		MinHeight: 300,
		MaxWidth:  600,
		MaxHeight: 500,
	}

	// Create gmenu without test app - this will create a real visible GUI
	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	testItems := []string{
		"Apple 🍎",
		"Banana 🍌", 
		"Cherry 🍒",
		"Date 🌴",
		"Elderberry 🫐",
	}

	t.Log("Setting up menu with initial items...")
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Start the app in a goroutine so we can control it
	appDone := make(chan struct{})
	go func() {
		defer close(appDone)
		err := gmenu.RunAppForever()
		if err != nil {
			t.Errorf("RunAppForever failed: %v", err)
		}
	}()

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	// Test multiple hide/show cycles with user-visible pauses
	for cycle := 0; cycle < 3; cycle++ {
		t.Logf("=== Cycle %d ===", cycle+1)
		
		// Show the menu
		t.Log("Showing GUI...")
		gmenu.ShowUI()
		assert.True(t, gmenu.IsShown())
		
		// Give user time to see the GUI
		time.Sleep(2 * time.Second)
		
		// Update with new items for this cycle
		newItems := []string{
			"Cycle " + string(rune('1'+cycle)) + " - Item A",
			"Cycle " + string(rune('1'+cycle)) + " - Item B", 
			"Cycle " + string(rune('1'+cycle)) + " - Item C",
		}
		
		t.Log("Updating items...")
		require.NoError(t, gmenu.SetupMenu(newItems, ""))
		
		// Give user time to see the updated items
		time.Sleep(2 * time.Second)
		
		// Simulate a search
		t.Log("Performing search...")
		results := gmenu.Search("Cycle")
		assert.GreaterOrEqual(t, len(results), 1)
		
		// Wait a bit more
		time.Sleep(1 * time.Second)
		
		// Simulate selection
		t.Log("Simulating selection...")
		gmenu.markSelectionMade()
		assert.True(t, gmenu.selectionFuse.IsBroken())
		
		// Wait for user to see the disabled state
		time.Sleep(1 * time.Second)
		
		// Hide the menu
		t.Log("Hiding GUI...")
		gmenu.HideUI()
		assert.False(t, gmenu.IsShown())
		
		// Reset for next cycle
		t.Log("Resetting for next cycle...")
		gmenu.Reset(true)
		assert.False(t, gmenu.selectionFuse.IsBroken())
		
		// Pause between cycles
		time.Sleep(1 * time.Second)
	}
	
	t.Log("Test completed successfully! Stopping app...")
	
	// Stop the app
	gmenu.Quit()
	
	// Wait for app to stop
	select {
	case <-appDone:
		t.Log("App stopped successfully")
	case <-time.After(2 * time.Second):
		t.Log("App stop timed out, but test completed")
	}
}

// TestVisualGUILongRunning tests the GUI with a long-running visible session
// This test will show the GUI for 10 seconds and perform various operations
// Run with: go test ./core -run TestVisualGUILongRunning -v
func TestVisualGUILongRunning(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping long-running visual GUI test in short mode")
	}
	
	// Skip this test in automated test runs due to Fyne main goroutine requirement
	t.Skip("Skipping visual GUI test - requires manual execution with real GUI")

	config := &model.Config{
		MenuID:    "long-running-visual-test",
		Title:     "Long Running Visual Test",
		Prompt:    "Search items:",
		MinWidth:  500,
		MinHeight: 400,
		MaxWidth:  800,
		MaxHeight: 600,
	}

	// Create gmenu without test app - this will create a real visible GUI
	gmenu, err := NewGMenu(FuzzySearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	// Create a large list of items
	items := make([]string, 50)
	for i := 0; i < 50; i++ {
		items[i] = "Item " + string(rune('A'+i%26)) + string(rune('0'+i%10)) + " - Long description for testing"
	}

	t.Log("Setting up menu with 50 items...")
	require.NoError(t, gmenu.SetupMenu(items, ""))

	// Start the app in a goroutine
	appDone := make(chan struct{})
	go func() {
		defer close(appDone)
		err := gmenu.RunAppForever()
		if err != nil {
			t.Errorf("RunAppForever failed: %v", err)
		}
	}()

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	t.Log("Showing GUI for 10 seconds...")
	gmenu.ShowUI()
	assert.True(t, gmenu.IsShown())

	// Perform various operations over 10 seconds
	operations := []struct {
		delay  time.Duration
		action func()
		desc   string
	}{
		{1 * time.Second, func() { gmenu.Search("Item A") }, "Searching for 'Item A'"},
		{2 * time.Second, func() { gmenu.Search("") }, "Clearing search"},
		{1 * time.Second, func() { gmenu.Search("Long") }, "Searching for 'Long'"},
		{2 * time.Second, func() { gmenu.Search("Item B") }, "Searching for 'Item B'"},
		{1 * time.Second, func() { gmenu.Search("") }, "Clearing search again"},
		{2 * time.Second, func() { 
			// Add some new items
			newItems := []string{"🚀 New Item 1", "🎯 New Item 2", "⭐ New Item 3"}
			gmenu.AppendItems(newItems)
		}, "Adding new items"},
		{1 * time.Second, func() { gmenu.Search("New") }, "Searching for new items"},
	}

	for _, op := range operations {
		time.Sleep(op.delay)
		t.Log(op.desc)
		op.action()
	}

	t.Log("Hiding GUI...")
	gmenu.HideUI()
	assert.False(t, gmenu.IsShown())

	t.Log("Long running test completed successfully! Stopping app...")
	
	// Stop the app
	gmenu.Quit()
	
	// Wait for app to stop
	select {
	case <-appDone:
		t.Log("App stopped successfully")
	case <-time.After(2 * time.Second):
		t.Log("App stop timed out, but test completed")
	}
}

// TestVisualGUIStressTest performs rapid operations while GUI is visible
// This helps test for visual glitches or hangs under stress
// Run with: go test ./core -run TestVisualGUIStressTest -v
func TestVisualGUIStressTest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping visual GUI stress test in short mode")
	}
	
	// Skip this test in automated test runs due to Fyne main goroutine requirement
	t.Skip("Skipping visual GUI test - requires manual execution with real GUI")

	config := &model.Config{
		MenuID:    "stress-visual-test",
		Title:     "Stress Test Visual",
		Prompt:    "Stress testing:",
		MinWidth:  400,
		MinHeight: 300,
	}

	gmenu, err := NewGMenu(FuzzySearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	// Start with some items
	initialItems := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
	require.NoError(t, gmenu.SetupMenu(initialItems, ""))

	// Start the app in a goroutine
	appDone := make(chan struct{})
	go func() {
		defer close(appDone)
		err := gmenu.RunAppForever()
		if err != nil {
			t.Errorf("RunAppForever failed: %v", err)
		}
	}()

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	t.Log("Starting stress test - GUI will be visible for 5 seconds with rapid operations...")
	gmenu.ShowUI()
	assert.True(t, gmenu.IsShown())

	// Perform rapid operations
	searches := []string{"A", "B", "C", "", "Apple", "Ban", "", "Cherry", "D", ""}
	
	for i := 0; i < 50; i++ {
		// Rapid search changes
		searchTerm := searches[i%len(searches)]
		gmenu.Search(searchTerm)
		
		// Occasionally update items
		if i%10 == 0 {
			newItems := []string{
				"Stress Item " + string(rune('0'+i%10)),
				"Test Item " + string(rune('A'+i%26)),
			}
			gmenu.AppendItems(newItems)
		}
		
		// Small delay to make it visible but still stress the system
		time.Sleep(100 * time.Millisecond)
	}

	t.Log("Hiding GUI...")
	gmenu.HideUI()
	assert.False(t, gmenu.IsShown())

	t.Log("Stress test completed successfully! Stopping app...")
	
	// Stop the app
	gmenu.Quit()
	
	// Wait for app to stop
	select {
	case <-appDone:
		t.Log("App stopped successfully")
	case <-time.After(2 * time.Second):
		t.Log("App stop timed out, but test completed")
	}
}

// TestVisualGUIInteractive creates a simple interactive test
// This test will show a GUI and wait for manual interaction
// Press any key to continue through the test steps
func TestVisualGUIInteractive(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping interactive visual GUI test in short mode")
	}
	
	// Skip this test in automated test runs due to Fyne main goroutine requirement
	t.Skip("Skipping visual GUI test - requires manual execution with real GUI")

	config := &model.Config{
		MenuID:    "interactive-visual-test",
		Title:     "Interactive Visual Test - Try typing to search!",
		Prompt:    "Type to search:",
		MinWidth:  600,
		MinHeight: 400,
		MaxWidth:  800,
		MaxHeight: 600,
	}

	gmenu, err := NewGMenu(FuzzySearch, config)
	require.NoError(t, err)
	defer func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	}()

	// Set up some interesting items to search through
	items := []string{
		"🍎 Apple - Fresh red apple",
		"🍌 Banana - Yellow curved fruit",
		"🍒 Cherry - Small red fruit",
		"🥝 Kiwi - Fuzzy brown fruit",
		"🍓 Strawberry - Red berry",
		"🍊 Orange - Citrus fruit",
		"🍇 Grapes - Purple cluster",
		"🥭 Mango - Tropical yellow fruit",
		"🍑 Peach - Fuzzy orange fruit", 
		"🍐 Pear - Green or yellow fruit",
		"📱 iPhone - Apple smartphone",
		"💻 MacBook - Apple laptop",
		"🖥️ iMac - Apple desktop",
		"⌚ Apple Watch - Smart watch",
		"🎵 Apple Music - Streaming service",
		"📦 Package Manager - Software tool",
		"🔍 Search Engine - Find things",
		"🖱️ Computer Mouse - Pointing device",
		"⌨️ Keyboard - Input device",
		"🖨️ Printer - Output device",
	}

	t.Log("Setting up menu with sample items...")
	require.NoError(t, gmenu.SetupMenu(items, ""))

	// Start the app in a goroutine
	appDone := make(chan struct{})
	go func() {
		defer close(appDone)
		err := gmenu.RunAppForever()
		if err != nil {
			t.Errorf("RunAppForever failed: %v", err)
		}
	}()

	// Give app time to start
	time.Sleep(500 * time.Millisecond)

	t.Log("Showing interactive GUI - try typing to search!")
	t.Log("The window will stay open for 15 seconds so you can interact with it")
	t.Log("You can type to search, use arrow keys to navigate, press ESC to close")
	
	gmenu.ShowUI()
	assert.True(t, gmenu.IsShown())

	// Keep the GUI open for 15 seconds for manual interaction
	time.Sleep(15 * time.Second)

	t.Log("Interactive test completed! Stopping app...")
	
	// Stop the app
	gmenu.Quit()
	
	// Wait for app to stop
	select {
	case <-appDone:
		t.Log("App stopped successfully")
	case <-time.After(2 * time.Second):
		t.Log("App stop timed out, but test completed")
	}
}