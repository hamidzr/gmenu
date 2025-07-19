package core

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestKeyboardShortcuts tests all keyboard shortcuts and key handlers
func TestKeyboardShortcuts(t *testing.T) {
	t.Skip("Keyboard test needs investigation - filtered items issue")

	// Initialize test app to avoid Fyne theme issues
	testApp := test.NewApp()
	defer testApp.Quit()

	config := &model.Config{
		Title:                 "Keyboard Test",
		Prompt:                "Search",
		MenuID:                "keyboard-test",
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
	gmenu.SetupMenu(testItems, "")

	// Test Enter key (selection)
	enterEvent := &fyne.KeyEvent{Name: fyne.KeyReturn}
	gmenu.ui.SearchEntry.TypedKey(enterEvent)
	assert.True(t, gmenu.selectionFuse.IsBroken())

	// Reset for next test
	gmenu.Reset(true)

	// Test Escape key (cancellation)
	escapeEvent := &fyne.KeyEvent{Name: fyne.KeyEscape}
	gmenu.ui.SearchEntry.TypedKey(escapeEvent)
	assert.True(t, gmenu.selectionFuse.IsBroken())
	assert.Equal(t, model.UserCanceled, gmenu.exitCode)

	// Reset for next test
	gmenu.Reset(true)

	// Test Tab key (should be accepted)
	assert.True(t, gmenu.ui.SearchEntry.AcceptsTab())

	// Test arrow keys navigation
	testNavigationKeys(t, gmenu)

	// Test numeric keys
	testNumericKeys(t, gmenu)
}

func testNavigationKeys(t *testing.T, gmenu *GMenu) {
	// Reset to known state and re-setup menu
	gmenu.Reset(true)
	testItems := []string{"item1", "item2", "item3", "item4", "item5"}
	gmenu.SetupMenu(testItems, "")
	gmenu.menu.Selected = 0

	// Test Down arrow
	downEvent := &fyne.KeyEvent{Name: fyne.KeyDown}
	gmenu.ui.SearchEntry.TypedKey(downEvent)
	assert.Equal(t, 1, gmenu.menu.Selected)

	// Test Up arrow
	upEvent := &fyne.KeyEvent{Name: fyne.KeyUp}
	gmenu.ui.SearchEntry.TypedKey(upEvent)
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Test wrapping at top (up from first item)
	gmenu.ui.SearchEntry.TypedKey(upEvent)
	assert.Equal(t, len(gmenu.menu.Filtered)-1, gmenu.menu.Selected)

	// Test wrapping at bottom (down from last item)
	gmenu.ui.SearchEntry.TypedKey(downEvent)
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Test Page Down
	pageDownEvent := &fyne.KeyEvent{Name: fyne.KeyPageDown}
	gmenu.ui.SearchEntry.TypedKey(pageDownEvent)
	// Should move selection down by page size or to end

	// Test Page Up
	pageUpEvent := &fyne.KeyEvent{Name: fyne.KeyPageUp}
	gmenu.ui.SearchEntry.TypedKey(pageUpEvent)
	// Should move selection up by page size or to beginning

	// Test Home key
	homeEvent := &fyne.KeyEvent{Name: fyne.KeyHome}
	gmenu.ui.SearchEntry.TypedKey(homeEvent)
	assert.Equal(t, 0, gmenu.menu.Selected)

	// Test End key
	endEvent := &fyne.KeyEvent{Name: fyne.KeyEnd}
	gmenu.ui.SearchEntry.TypedKey(endEvent)
	assert.Equal(t, len(gmenu.menu.Filtered)-1, gmenu.menu.Selected)
}

func testNumericKeys(t *testing.T, gmenu *GMenu) {
	// Reset to known state and re-setup menu
	gmenu.Reset(true)
	testItems := []string{"item1", "item2", "item3", "item4", "item5"}
	gmenu.SetupMenu(testItems, "")

	numericKeys := []fyne.KeyName{
		fyne.Key1, fyne.Key2, fyne.Key3, fyne.Key4, fyne.Key5,
		fyne.Key6, fyne.Key7, fyne.Key8, fyne.Key9,
	}

	for i, key := range numericKeys {
		gmenu.Reset(true)

		keyEvent := &fyne.KeyEvent{Name: key}
		gmenu.ui.SearchEntry.TypedKey(keyEvent)

		// Should select the corresponding item (1-indexed to 0-indexed)
		if i < len(gmenu.menu.Filtered) {
			assert.Equal(t, i, gmenu.menu.Selected)
		}
	}
}

// TestKeyboardShortcutsWithModifiers tests Ctrl+key combinations
func TestKeyboardShortcutsWithModifiers(t *testing.T) {
	config := &model.Config{
		Title:                 "Shortcut Test",
		Prompt:                "Search",
		MenuID:                "shortcut-test",
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
	gmenu.SetupMenu(testItems, "")

	// Test Ctrl+L (clear search)
	gmenu.ui.SearchEntry.SetText("some text")
	ctrlL := &desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyL}
	gmenu.ui.SearchEntry.TypedShortcut(ctrlL)
	assert.Equal(t, "", gmenu.ui.SearchEntry.Text)

	// Test Ctrl+C (copy - should be handled by default Entry behavior)
	gmenu.ui.SearchEntry.SetText("copy this")
	gmenu.ui.SearchEntry.SelectAll()
	ctrlC := &desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyC}
	// This mainly tests that it doesn't crash
	assert.NotPanics(t, func() {
		gmenu.ui.SearchEntry.TypedShortcut(ctrlC)
	})

	// Test Ctrl+V (paste - should be handled by default Entry behavior)
	ctrlV := &desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyV}
	assert.NotPanics(t, func() {
		gmenu.ui.SearchEntry.TypedShortcut(ctrlV)
	})

	// Test Ctrl+A (select all)
	gmenu.ui.SearchEntry.SetText("select all this")
	ctrlA := &desktop.CustomShortcut{Modifier: fyne.KeyModifierControl, KeyName: fyne.KeyA}
	assert.NotPanics(t, func() {
		gmenu.ui.SearchEntry.TypedShortcut(ctrlA)
	})
}

// TestNumericSelectionDisabled tests when numeric selection is disabled
func TestNumericSelectionDisabled(t *testing.T) {
	config := &model.Config{
		Title:                 "No Numeric Test",
		Prompt:                "Search",
		MenuID:                "no-numeric-test",
		SearchMethod:          "fuzzy",
		PreserveOrder:         false,
		InitialQuery:          "",
		AutoAccept:            false,
		TerminalMode:          false,
		NoNumericSelection:    true, // Disable numeric selection
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
	gmenu.SetupMenu(testItems, "")

	originalSelected := gmenu.menu.Selected

	// Try numeric key - should not change selection
	key1Event := &fyne.KeyEvent{Name: fyne.Key1}
	gmenu.ui.SearchEntry.TypedKey(key1Event)

	// Selection should remain unchanged when numeric selection is disabled
	assert.Equal(t, originalSelected, gmenu.menu.Selected)
}

// TestKeyPropagationBlacklist tests that certain keys can be blocked
func TestKeyPropagationBlacklist(t *testing.T) {
	config := &model.Config{
		Title:                 "Propagation Test",
		Prompt:                "Search",
		MenuID:                "propagation-test",
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
	gmenu.SetupMenu(testItems, "")

	// Set up blacklist for Enter key
	gmenu.ui.SearchEntry.PropagationBlacklist = map[fyne.KeyName]bool{
		fyne.KeyReturn: true,
	}

	originalText := "test text"
	gmenu.ui.SearchEntry.SetText(originalText)

	// Try Enter key - should be blocked from normal Entry processing
	enterEvent := &fyne.KeyEvent{Name: fyne.KeyReturn}
	gmenu.ui.SearchEntry.TypedKey(enterEvent)

	// The OnKeyDown handler should still be called, but Entry.TypedKey should be blocked
	// Selection should still be made because the key handler processes it
	assert.True(t, gmenu.selectionFuse.IsBroken())
}

// TestFocusHandling tests focus-related keyboard behavior
func TestFocusHandling(t *testing.T) {
	config := &model.Config{
		Title:                 "Focus Test",
		Prompt:                "Search",
		MenuID:                "focus-test",
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
	gmenu.SetupMenu(testItems, "")

	// Test focus loss callback
	focusLost := false
	gmenu.ui.SearchEntry.OnFocusLost = func() {
		focusLost = true
	}

	// Simulate focus lost
	gmenu.ui.SearchEntry.FocusLost()
	assert.True(t, focusLost)

	// Test focus gained
	gmenu.ui.SearchEntry.FocusGained()
	// Should not crash or cause issues
}

// TestKeyHandlerCallback tests custom key handler functionality
func TestKeyHandlerCallback(t *testing.T) {
	config := &model.Config{
		Title:                 "Key Handler Test",
		Prompt:                "Search",
		MenuID:                "key-handler-test",
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
	gmenu.SetupMenu(testItems, "")

	// Track key events
	var lastKeyEvent *fyne.KeyEvent
	gmenu.ui.SearchEntry.OnKeyDown = func(key *fyne.KeyEvent) {
		lastKeyEvent = key
	}

	// Send a key event
	downEvent := &fyne.KeyEvent{Name: fyne.KeyDown}
	gmenu.ui.SearchEntry.TypedKey(downEvent)

	// Verify the callback was called
	require.NotNil(t, lastKeyEvent)
	assert.Equal(t, fyne.KeyDown, lastKeyEvent.Name)
}

// TestSpecialKeys tests handling of special function keys
func TestSpecialKeys(t *testing.T) {
	config := &model.Config{
		Title:                 "Special Keys Test",
		Prompt:                "Search",
		MenuID:                "special-keys-test",
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
	gmenu.SetupMenu(testItems, "")

	specialKeys := []fyne.KeyName{
		fyne.KeyF1, fyne.KeyF2, fyne.KeyF3, fyne.KeyF4,
		fyne.KeyInsert, fyne.KeyDelete,
		fyne.KeyLeft, fyne.KeyRight,
		fyne.KeySpace, fyne.KeyBackspace,
	}

	// Test that special keys don't cause crashes
	for _, key := range specialKeys {
		t.Run(string(key), func(t *testing.T) {
			keyEvent := &fyne.KeyEvent{Name: key}
			assert.NotPanics(t, func() {
				gmenu.ui.SearchEntry.TypedKey(keyEvent)
			})
		})
	}
}
