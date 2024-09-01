package core

import (
	"testing"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
)

func strToItems(titles []string) []model.MenuItem {
	items := make([]model.MenuItem, len(titles))
	for i, entry := range titles {
		items[i] = model.MenuItem{Title: entry}
	}
	return items
}

func itemsToStr(items []model.MenuItem) []string {
	titles := make([]string, len(items))
	for i, entry := range items {
		titles[i] = entry.Title
	}
	return titles
}

func TestFuzzy(t *testing.T) {
	inputList := []string{
		"whatsapp",
		"whaxtsapp",
		"whats app",
	}

	type TestCase struct {
		items         []string
		query         string
		expectedItems []string
	}

	testCases := []TestCase{
		{
			items:         inputList,
			query:         "atsa",
			expectedItems: []string{"whatsapp", "whaxtsapp", "whats app"},
		},
	}

	for _, tc := range testCases {
		items := strToItems(tc.items)
		res := FuzzySearchBrute(items, tc.query, false, 3)
		assert.Equal(t, tc.expectedItems, itemsToStr(res))
	}
}
