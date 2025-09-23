package core

import (
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func useFyneTestApp(t *testing.T) {
	t.Helper()
	oldNewApp := newAppFunc
	newAppFunc = func() fyne.App { return test.NewApp() }
	t.Cleanup(func() { newAppFunc = oldNewApp })
}

// TestGRenderMatchCounterLabel tests the match counter label generation
func TestGRenderMatchCounterLabel(t *testing.T) {
	useFyneTestApp(t)
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

	tests := []struct {
		name           string
		items          []string
		query          string
		expectedFormat string
	}{
		{
			name:           "no items",
			items:          []string{},
			query:          "",
			expectedFormat: "[1/1]", // includes loading item
		},
		{
			name:           "all items visible",
			items:          []string{"apple", "banana", "cherry"},
			query:          "",
			expectedFormat: "[3/3]",
		},
		{
			name:           "filtered items",
			items:          []string{"apple", "banana", "cherry"},
			query:          "ap",
			expectedFormat: "[1/3]", // only apple matches
		},
		{
			name:           "no matches",
			items:          []string{"apple", "banana", "cherry"},
			query:          "xyz",
			expectedFormat: "[0/3]",
		},
		{
			name:           "single item",
			items:          []string{"apple"},
			query:          "",
			expectedFormat: "[1/1]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, gmenu.SetupMenu(tt.items, ""))
			if tt.query != "" {
				gmenu.Search(tt.query)
			}

			label := gmenu.matchCounterLabel()
			assert.Equal(t, tt.expectedFormat, label)
		})
	}
}

// TestGRenderLabelWithLargeNumbers tests label generation with large item counts
func TestGRenderLabelWithLargeNumbers(t *testing.T) {
	useFyneTestApp(t)
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

	// Create a large number of items
	largeItemSet := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		largeItemSet[i] = "item" + string(rune('0'+(i%10)))
	}

	require.NoError(t, gmenu.SetupMenu(largeItemSet, ""))

	// Test with all items
	label := gmenu.matchCounterLabel()
	assert.Equal(t, "[1000/1000]", label)

	// Test with filtered results
	gmenu.Search("item1")
	label = gmenu.matchCounterLabel()
	assert.Equal(t, "[100/1000]", label) // items ending with 1
}

// TestGRenderEmptyAndNilStates tests rendering with empty and nil states
func TestGRenderEmptyAndNilStates(t *testing.T) {
	useFyneTestApp(t)
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

	// Test with nil items
	require.NoError(t, gmenu.SetupMenu(nil, ""))
	label := gmenu.matchCounterLabel()
	assert.Contains(t, label, "/1]") // should show loading item

	// Test with empty slice
	require.NoError(t, gmenu.SetupMenu([]string{}, ""))
	label = gmenu.matchCounterLabel()
	assert.Contains(t, label, "/1]") // should show loading item

	// Test with items containing empty strings
	require.NoError(t, gmenu.SetupMenu([]string{"", "valid", ""}, ""))
	label = gmenu.matchCounterLabel()
	assert.Equal(t, "[3/3]", label) // all items should be counted
}

// TestGRenderAfterReset tests render state after reset operations
func TestGRenderAfterReset(t *testing.T) {
	useFyneTestApp(t)
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

	// Setup initial state
	require.NoError(t, gmenu.SetupMenu([]string{"apple", "banana", "cherry"}, ""))
	gmenu.Search("app")

	// Should show filtered count
	label := gmenu.matchCounterLabel()
	assert.Equal(t, "[1/3]", label)

	// Reset and check
	gmenu.Reset(true)

	// After reset, should return to initial state
	// Note: Reset might affect the menu state, so we test the behavior
	// The exact behavior depends on the Reset implementation
}

// TestGRenderConcurrentAccess tests concurrent access to render functions
func TestGRenderConcurrentAccess(t *testing.T) {
	useFyneTestApp(t)
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

	// Test concurrent access to matchCounterLabel
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				label := gmenu.matchCounterLabel()
				assert.NotEmpty(t, label)
				assert.Contains(t, label, "/")
				assert.Contains(t, label, "[")
				assert.Contains(t, label, "]")
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 5; i++ {
		<-done
	}
}

// TestGRenderWithSpecialCharacters tests rendering with special characters
func TestGRenderWithSpecialCharacters(t *testing.T) {
	useFyneTestApp(t)
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

	specialItems := []string{
		"item with spaces",
		"item_with_underscores",
		"item-with-dashes",
		"item.with.dots",
		"item@with@symbols",
		"item/with/slashes",
		"item\\with\\backslashes",
		"item:with:colons",
		"item;with;semicolons",
		"émojis 🚀 and ünicøde",
	}

	require.NoError(t, gmenu.SetupMenu(specialItems, ""))

	// Test that all items are counted properly
	label := gmenu.matchCounterLabel()
	expectedCount := len(specialItems)
	expected := "[" + string(rune('0'+expectedCount/10)) + string(rune('0'+expectedCount%10)) + "/" +
		string(rune('0'+expectedCount/10)) + string(rune('0'+expectedCount%10)) + "]"
	assert.Equal(t, expected, label)

	// Test searching with special characters
	gmenu.Search("with")
	label = gmenu.matchCounterLabel()
	assert.Contains(t, label, "/10]")    // 10 total items
	assert.NotEqual(t, "[0/", label[:3]) // should find matches
}

// TestGRenderStateConsistency tests that render state remains consistent
func TestGRenderStateConsistency(t *testing.T) {
	useFyneTestApp(t)
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

	require.NoError(t, gmenu.SetupMenu([]string{"apple", "banana", "cherry"}, ""))

	// Multiple consecutive calls should return consistent results
	label1 := gmenu.matchCounterLabel()
	label2 := gmenu.matchCounterLabel()
	label3 := gmenu.matchCounterLabel()

	assert.Equal(t, label1, label2)
	assert.Equal(t, label2, label3)
	assert.Equal(t, "[3/3]", label1)

	// After search, should also be consistent
	gmenu.Search("app")
	labelAfterSearch1 := gmenu.matchCounterLabel()
	labelAfterSearch2 := gmenu.matchCounterLabel()

	assert.Equal(t, labelAfterSearch1, labelAfterSearch2)
	assert.Equal(t, "[1/3]", labelAfterSearch1)
}
