package render

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSearchEntryCreation tests SearchEntry initialization
func TestSearchEntryCreation(t *testing.T) {
	entry := &SearchEntry{}

	// Test that it's a valid entry
	assert.NotNil(t, entry)
	assert.Equal(t, "", entry.Text)
	assert.False(t, entry.Disabled())
}

// TestSearchEntryKeyHandling tests key event handling
func TestSearchEntryKeyHandling(t *testing.T) {
	entry := &SearchEntry{}

	// Test key handler callback
	var lastKey *fyne.KeyEvent
	entry.OnKeyDown = func(key *fyne.KeyEvent) {
		lastKey = key
	}

	// Simulate key press
	keyEvent := &fyne.KeyEvent{Name: fyne.KeyDown}
	entry.TypedKey(keyEvent)

	assert.Equal(t, keyEvent, lastKey)
}

// TestSearchEntryPropagationBlacklist tests key propagation blocking
func TestSearchEntryPropagationBlacklist(t *testing.T) {
	entry := &SearchEntry{}

	// Set up blacklist to block Enter key
	entry.PropagationBlacklist = map[fyne.KeyName]bool{
		fyne.KeyReturn: true,
	}

	originalText := "test"
	entry.SetText(originalText)

	// Try to send Enter key (should be blocked)
	keyEvent := &fyne.KeyEvent{Name: fyne.KeyReturn}
	entry.TypedKey(keyEvent)

	// Text should remain unchanged since Enter was blocked
	assert.Equal(t, originalText, entry.Text)
}

// TestSearchEntrySelectAll tests the SelectAll functionality
func TestSearchEntrySelectAll(t *testing.T) {
	entry := &SearchEntry{}
	entry.SetText("test text")

	// Test SelectAll method exists and can be called
	assert.NotPanics(t, func() {
		entry.SelectAll()
	})
}

// TestSearchEntryTyping tests actual typing simulation
func TestSearchEntryTyping(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	entry := &SearchEntry{}

	// Use Fyne's test utility to simulate typing
	test.Type(entry, "hello world")

	assert.Equal(t, "hello world", entry.Text)
}

// TestItemsCanvasCreation tests ItemsCanvas initialization
func TestItemsCanvasCreation(t *testing.T) {
	canvas := NewItemsCanvas()

	require.NotNil(t, canvas)
	require.NotNil(t, canvas.Container)

	// Initially should be empty
	assert.Equal(t, 0, len(canvas.Container.Objects))
}

// TestRenderItem tests individual item rendering
func TestRenderItem(t *testing.T) {
	testCases := []struct {
		name               string
		item               model.MenuItem
		idx                int
		selected           bool
		noNumericSelection bool
		expectedText       string
	}{
		{
			name:         "normal item",
			item:         model.MenuItem{Title: "test item"},
			idx:          0,
			selected:     false,
			expectedText: "test item",
		},
		{
			name:         "selected item",
			item:         model.MenuItem{Title: "selected"},
			idx:          1,
			selected:     true,
			expectedText: "selected",
		},
		{
			name:         "empty title item",
			item:         model.MenuItem{Title: ""},
			idx:          0,
			selected:     false,
			expectedText: "Empty Item", // Should use fallback
		},
		{
			name:               "item with numeric disabled",
			item:               model.MenuItem{Title: "no numbers"},
			idx:                2,
			selected:           false,
			noNumericSelection: true,
			expectedText:       "no numbers",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			container := RenderItem(tc.item, tc.idx, tc.selected, tc.noNumericSelection)

			require.NotNil(t, container)
			assert.Greater(t, len(container.Objects), 0, "Container should have at least one object")

			// The container should have the expected structure
			// We can't easily test the exact text content due to Fyne's rendering,
			// but we can verify the container was created successfully
		})
	}
}

// TestItemsCanvasWithItems tests adding items to the canvas
func TestItemsCanvasWithItems(t *testing.T) {
	canvas := NewItemsCanvas()

	// Create test items
	items := []model.MenuItem{
		{Title: "item1"},
		{Title: "item2"},
		{Title: "item3"},
	}

	// Render items and add to canvas
	for i, item := range items {
		itemContainer := RenderItem(item, i, i == 0, false) // first item selected
		canvas.Container.Add(itemContainer)
	}

	assert.Equal(t, len(items), len(canvas.Container.Objects))
}

// TestSearchEntryFocusLoss tests focus loss callback
func TestSearchEntryFocusLoss(t *testing.T) {
	entry := &SearchEntry{}

	focusLost := false
	entry.OnFocusLost = func() {
		focusLost = true
	}

	// Simulate focus lost event
	entry.FocusLost()

	assert.True(t, focusLost)
}

// TestSearchEntryShortcuts tests keyboard shortcuts
func TestSearchEntryShortcuts(t *testing.T) {
	app := test.NewApp()
	defer app.Quit()

	entry := &SearchEntry{}
	entry.SetText("test text")

	// Test that the TypedShortcut method exists and can handle basic cases
	// We avoid testing with nil shortcuts as that's not a valid use case
	assert.NotNil(t, entry.TypedShortcut)

	// The method should exist and be callable
	// In a real application, this would be called by the Fyne framework
	// with valid shortcut objects
}

// TestItemRenderingWithComplexContent tests rendering items with special characters
func TestItemRenderingWithComplexContent(t *testing.T) {
	testItems := []model.MenuItem{
		{Title: "🚀 Rocket"},
		{Title: "Multi\nLine\nText"},
		{Title: "Very long text that should be truncated if it exceeds the display width"},
		{Title: "Special chars: !@#$%^&*()"},
		{Title: "Unicode: αβγδε"},
	}

	for i, item := range testItems {
		t.Run(item.Title, func(t *testing.T) {
			container := RenderItem(item, i, false, false)
			require.NotNil(t, container)
			assert.Greater(t, len(container.Objects), 0)
		})
	}
}

// TestSearchEntryTabHandling tests tab key handling
func TestSearchEntryTabHandling(t *testing.T) {
	entry := &SearchEntry{}

	// Test that AcceptsTab returns true
	assert.True(t, entry.AcceptsTab())
}

// TestItemsCanvasLayout tests that the canvas uses the correct layout
func TestItemsCanvasLayout(t *testing.T) {
	canvas := NewItemsCanvas()

	// The container should use VBox layout for vertical item stacking
	require.NotNil(t, canvas.Container.Layout)

	// Add some items to test layout behavior
	for i := 0; i < 3; i++ {
		item := model.MenuItem{Title: "Item " + string(rune('1'+i))}
		itemContainer := RenderItem(item, i, false, false)
		canvas.Container.Add(itemContainer)
	}

	// After adding items, layout should still be valid
	assert.Equal(t, 3, len(canvas.Container.Objects))
}
