package core

import (
	"testing"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchEntryStateAfterReset tests the core issue with SearchEntry state management
func TestSearchEntryStateAfterReset(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return testApp }
	defer func() { newAppFunc = oldNewApp }()

	// Create a test config
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-menu",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1920,
		MaxHeight:             1080,
		AcceptCustomSelection: true,
	}

	// Create GMenu instance
	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	// Setup initial menu
	testItems := []string{"apple", "banana", "cherry", "date"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test initial state
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false initially")
	assert.False(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be enabled initially")

	// Simulate ShowUI (which initializes the WaitGroup)
	gmenu.ShowUI()

	// Simulate user making a selection (this disables the SearchEntry)
	gmenu.markSelectionMade()
	assert.True(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be true after selection")
	assert.True(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be disabled after selection")

	// Reset the menu (this should fix the state)
	gmenu.Reset(true)
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false after Reset")
	assert.False(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be enabled after Reset")

	// Test that SearchEntry can accept text
	gmenu.ui.SearchEntry.SetText("test input")
	assert.Equal(t, "test input", gmenu.ui.SearchEntry.Text, "SearchEntry should accept new text after Reset")
}

// TestResetResetsAllNecessaryState ensures Reset method resets all the required state
func TestResetResetsAllNecessaryState(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return testApp }
	defer func() { newAppFunc = oldNewApp }()

	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-menu-state",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1920,
		MaxHeight:             1080,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	testItems := []string{"item1", "item2", "item3"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Set some state that should be reset
	gmenu.markSelectionMade() // This will break the fuse
	gmenu.exitCode = model.NoError
	gmenu.ui.SearchEntry.Disable()
	gmenu.menu.Selected = 2

	// Verify the state is set
	assert.True(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be true before reset")
	assert.Equal(t, model.NoError, gmenu.exitCode, "exitCode should be NoError before reset")
	assert.True(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be disabled before reset")
	assert.Equal(t, 2, gmenu.menu.Selected, "Selected should be 2 before reset")

	// Reset
	gmenu.Reset(true)

	// Verify all state is properly reset
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false after reset")
	assert.Equal(t, model.Unset, gmenu.exitCode, "exitCode should be Unset after reset")
	assert.False(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be enabled after reset")
	assert.Equal(t, 0, gmenu.menu.Selected, "Selected should be 0 after reset")
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text, "SearchEntry text should be empty after reset with resetInput=true")
}

// TestSearchEntryInputAfterHideShowReset reproduces the issue where
// SearchEntry becomes unresponsive after hide/show/reset cycles
func TestSearchEntryInputAfterHideShowReset(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return testApp }
	defer func() { newAppFunc = oldNewApp }()

	// Create a test config
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-menu",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1920,
		MaxHeight:             1080,
		AcceptCustomSelection: true,
	}

	// Create GMenu instance
	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	// Setup initial menu
	testItems := []string{"apple", "banana", "cherry", "date"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test initial state
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false initially")
	assert.True(t, gmenu.ui.SearchEntry.Disabled() == false, "SearchEntry should be enabled initially")

	// Simulate first show
	gmenu.ShowUI()
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false after ShowUI")
	assert.True(t, gmenu.ui.SearchEntry.Disabled() == false, "SearchEntry should be enabled after ShowUI")

	// Simulate user making a selection (this would normally happen via key press)
	gmenu.markSelectionMade()
	assert.True(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be true after selection")
	assert.True(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be disabled after selection")

	// Hide the UI
	gmenu.HideUI()
	assert.False(t, gmenu.IsShown(), "UI should be hidden")

	// Reset the menu (this is where the bug might be)
	gmenu.Reset(true)
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false after Reset")
	assert.True(t, gmenu.ui.SearchEntry.Disabled() == false, "SearchEntry should be enabled after Reset")

	// Show UI again (second time)
	gmenu.ShowUI()
	assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false after second ShowUI")
	assert.True(t, gmenu.ui.SearchEntry.Disabled() == false, "SearchEntry should be enabled after second ShowUI")

	// Test that we can simulate typing (this is where the issue manifests)
	// In a real scenario, this would be user typing, but we can simulate it
	originalText := gmenu.ui.SearchEntry.Text
	gmenu.ui.SearchEntry.SetText("test input")

	// Give a small delay to allow any async operations
	time.Sleep(10 * time.Millisecond)

	newText := gmenu.ui.SearchEntry.Text
	assert.Equal(t, "test input", newText, "SearchEntry should accept new text after second show")
	assert.NotEqual(t, originalText, newText, "SearchEntry text should have changed")

	// Test that OnChanged callback works (this simulates actual typing)
	callbackTriggered := false
	gmenu.ui.SearchEntry.OnChanged = func(text string) {
		callbackTriggered = true
	}

	gmenu.ui.SearchEntry.SetText("another test")
	time.Sleep(10 * time.Millisecond)

	assert.True(t, callbackTriggered, "OnChanged callback should be triggered")

	// Clean up
	gmenu.HideUI()
}

// TestMultipleHideShowCycles tests multiple cycles to ensure the issue is reproducible
func TestMultipleHideShowCycles(t *testing.T) {
	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return testApp }
	defer func() { newAppFunc = oldNewApp }()

	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-menu-cycles",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              600,
		MinHeight:             300,
		MaxWidth:              1920,
		MaxHeight:             1080,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	testItems := []string{"item1", "item2", "item3"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test multiple cycles
	for i := 0; i < 3; i++ {
		t.Logf("Testing cycle %d", i+1)

		// Show UI
		gmenu.ShowUI()
		assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false at start of cycle %d", i+1)
		assert.False(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be enabled at start of cycle %d", i+1)

		// Simulate selection
		gmenu.markSelectionMade()
		assert.True(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be true after selection in cycle %d", i+1)
		assert.True(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be disabled after selection in cycle %d", i+1)

		// Hide and reset
		gmenu.HideUI()
		gmenu.Reset(true)

		// Verify state after reset
		assert.False(t, gmenu.selectionFuse.IsBroken(), "hasSelection should be false after reset in cycle %d", i+1)
		assert.False(t, gmenu.ui.SearchEntry.Disabled(), "SearchEntry should be enabled after reset in cycle %d", i+1)

		// Test input capability
		testText := "cycle" + string(rune('0'+i+1))
		gmenu.ui.SearchEntry.SetText(testText)
		time.Sleep(5 * time.Millisecond)

		assert.Equal(t, testText, gmenu.ui.SearchEntry.Text, "SearchEntry should accept input in cycle %d", i+1)
	}

	gmenu.HideUI()
}
