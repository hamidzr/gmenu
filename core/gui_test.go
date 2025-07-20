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

// TestGUIComponentInitialization tests that all GUI components are properly initialized
func TestGUIComponentInitialization(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-gui",
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
	require.NotNil(t, gmenu)

	// Test UI components are initialized
	require.NotNil(t, gmenu.ui, "UI should be initialized")
	require.NotNil(t, gmenu.ui.MainWindow, "MainWindow should be initialized")
	require.NotNil(t, gmenu.ui.SearchEntry, "SearchEntry should be initialized")
	require.NotNil(t, gmenu.ui.ItemsCanvas, "ItemsCanvas should be initialized")
	require.NotNil(t, gmenu.ui.MenuLabel, "MenuLabel should be initialized")

	// Test window properties
	assert.Equal(t, config.Title, gmenu.ui.MainWindow.Title())
	assert.False(t, gmenu.ui.MainWindow.FullScreen())

	// Test initial search entry state
	assert.False(t, gmenu.ui.SearchEntry.Disabled())
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text)

	// Test menu label - it starts with "menulabel" and gets updated later
	assert.Contains(t, []string{"menulabel", config.Prompt}, gmenu.ui.MenuLabel.Text)
}

// TestSearchEntryInteraction tests user interaction with the search entry
func TestSearchEntryInteraction(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-search",
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

	testItems := []string{"apple", "banana", "cherry", "date", "elderberry"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test initial state
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text)

	// Test typing in search entry
	test.Type(gmenu.ui.SearchEntry, "app")
	assert.Equal(t, "app", gmenu.ui.SearchEntry.Text)

	// Give a moment for search to process
	time.Sleep(10 * time.Millisecond)

	// Test clearing search entry
	gmenu.ui.SearchEntry.SetText("")
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text)

	// Test setting text programmatically
	gmenu.ui.SearchEntry.SetText("cherry")
	assert.Equal(t, "cherry", gmenu.ui.SearchEntry.Text)
}

// TestKeyboardNavigation tests keyboard navigation through menu items
func TestKeyboardNavigation(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-keyboard",
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

	// Test initial selection
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Test down arrow key
	downEvent := &fyne.KeyEvent{Name: fyne.KeyDown}
	gmenu.ui.SearchEntry.TypedKey(downEvent)
	// Selection should move down
	assert.Equal(t, 1, gmenu.menu.Selected)

	// Test up arrow key
	upEvent := &fyne.KeyEvent{Name: fyne.KeyUp}
	gmenu.ui.SearchEntry.TypedKey(upEvent)
	// Selection should move back up
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Test wrapping at the top
	gmenu.ui.SearchEntry.TypedKey(upEvent)
	// Should wrap to last item
	assert.Equal(t, len(testItems)-1, gmenu.menu.Selected)

	// Test wrapping at the bottom
	gmenu.menu.Selected = len(testItems) - 1
	gmenu.ui.SearchEntry.TypedKey(downEvent)
	// Should wrap to first item
	assert.Equal(t, 0, gmenu.menu.Selected)
}

// TestNumericSelection tests numeric key selection
func TestNumericSelection(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-numeric",
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

	testItems := []string{"item1", "item2", "item3", "item4", "item5"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test numeric key selection
	key1Event := &fyne.KeyEvent{Name: fyne.Key1}
	gmenu.ui.SearchEntry.TypedKey(key1Event)
	assert.Equal(t, 0, gmenu.menu.Selected) // 1 maps to index 0

	key3Event := &fyne.KeyEvent{Name: fyne.Key3}
	gmenu.ui.SearchEntry.TypedKey(key3Event)
	assert.Equal(t, 2, gmenu.menu.Selected) // 3 maps to index 2
}

// TestSearchFiltering tests that search properly filters items
func TestSearchFiltering(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-filter",
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

	testItems := []string{"apple", "application", "banana", "cherry", "app"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test initial state - all items should be visible
	assert.Len(t, gmenu.menu.items, len(testItems))

	// Test filtering with "app"
	test.Type(gmenu.ui.SearchEntry, "app")
	time.Sleep(10 * time.Millisecond) // Allow search to process

	// Should show items containing "app"
	filteredItems := gmenu.menu.Filtered
	assert.GreaterOrEqual(t, len(filteredItems), 2) // At least "apple" and "app"

	// Clear search
	gmenu.ui.SearchEntry.SetText("")
	time.Sleep(10 * time.Millisecond)

	// All items should be visible again
	filteredItems = gmenu.menu.Filtered
	assert.Len(t, filteredItems, len(testItems))
}

// TestGUIStateManagement tests UI state changes
func TestGUIStateManagement(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-state",
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

	// Test initial enabled state
	assert.False(t, gmenu.ui.SearchEntry.Disabled())

	// Test disabling after selection
	gmenu.markSelectionMade()
	assert.True(t, gmenu.ui.SearchEntry.Disabled())
	assert.True(t, gmenu.selectionFuse.IsBroken())

	// Test re-enabling after reset
	gmenu.Reset(true)
	assert.False(t, gmenu.ui.SearchEntry.Disabled())
}

// TestWindowDimensions tests window sizing
func TestWindowDimensions(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-dimensions",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    false,
		MinWidth:              800,
		MinHeight:             400,
		MaxWidth:              1600,
		MaxHeight:             1000,
		AcceptCustomSelection: true,
	}

	searchMethod := SearchMethods["fuzzy"]
	gmenu, err := NewGMenu(searchMethod, config)
	require.NoError(t, err)

	// Test that dimensions are stored correctly
	assert.Equal(t, float32(800), gmenu.dims.MinWidth)
	assert.Equal(t, float32(400), gmenu.dims.MinHeight)
	assert.Equal(t, float32(1600), gmenu.dims.MaxWidth)
	assert.Equal(t, float32(1000), gmenu.dims.MaxHeight)
}

// TestMenuItemRendering tests that menu items are properly rendered
func TestMenuItemRendering(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-render",
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

	testItems := []string{"short", "this is a longer item", "🚀 emoji item", ""}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Test that items canvas is not nil
	require.NotNil(t, gmenu.ui.ItemsCanvas)
	require.NotNil(t, gmenu.ui.ItemsCanvas.Container)

	// The container should have been set up for rendering
	assert.NotNil(t, gmenu.ui.ItemsCanvas.Container.Layout)
}

// TestInitialQueryHandling tests handling of initial query
func TestInitialQueryHandling(t *testing.T) {
	config := &model.Config{
		Title:                 "Test Menu",
		Prompt:                "Search",
		MenuID:                "test-initial",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "initial",
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

	testItems := []string{"initial value", "other", "initial setup"}
	require.NoError(t, gmenu.SetupMenu(testItems, config.InitialQuery))

	// Search entry should have the initial query
	assert.Equal(t, config.InitialQuery, gmenu.ui.SearchEntry.Text)
}
