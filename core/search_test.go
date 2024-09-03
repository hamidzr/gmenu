package core

import (
	"testing"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"golang.org/x/exp/rand"
)

func strToItems(titles *[]string) []model.MenuItem {
	if titles == nil {
		return nil
	}
	items := make([]model.MenuItem, len(*titles))
	for i, entry := range *titles {
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

func shuffle[T any](items *[]T) *[]T {
	for i := len(*items) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		(*items)[i], (*items)[j] = (*items)[j], (*items)[i]
	}
	return items
}

func TestWithSeparators(t *testing.T) {
	inputs := &[]string{
		"abcd",
		"ab cd",
	}

	type TestCase struct {
		items         *[]string
		query         string
		expectedItems *[]string
		searchMethod  SearchMethod
	}

	testCases := []TestCase{
		{
			items:         inputs,
			query:         "abc",
			searchMethod:  DirectSearch,
			expectedItems: &[]string{"abcd"},
		},
		{
			items:         inputs,
			query:         "cd ab",
			searchMethod:  SearchWithSeparator(" ", DirectSearch),
			expectedItems: &[]string{"abcd", "ab cd"},
		},
		{
			items:         inputs,
			query:         "abc",
			searchMethod:  SearchWithSeparator(" ", DirectSearch),
			expectedItems: &[]string{"abcd"}, // shouldn't match with the one with space in it.
		},
	}

	for _, tc := range testCases {
		items := strToItems(tc.items)
		res := tc.searchMethod(items, tc.query, false, 3)
		resStrs := itemsToStr(res)
		assert.Equal(t, *tc.expectedItems, resStrs, "query: %s", tc.query)
		// assert.ElementsMatch(t, *tc.expectedItems, resStrs, itemStrs, tc.query, resStrs)
	}
}

func TestFuzzy(t *testing.T) {
	unorderedInputs := &[]string{
		"whatsapp",
		"whaxtsapp",
		"whats app",
	}

	type TestCase struct {
		items             *[]string
		query             string
		expectedRelOrders [][]string
		expectedItems     *[]string
	}

	testCases := []TestCase{
		{
			items: unorderedInputs,
			query: "atsa",
			expectedRelOrders: [][]string{
				{"whatsapp", "whaxtsapp"},
				{"whatsapp", "whats app"},
			},
		},
		{
			unorderedInputs, "ats ap", [][]string{
				{"whats app", "whatsapp"},
				{"whatsapp", "whaxtsapp"},
			}, nil,
		},
	}

	// TODO: permuations instead of rand shuffling.
	for i := 0; i < 10; i++ { // for each order of inputs.
		unorderedInputs = shuffle(unorderedInputs)
		for _, tc := range testCases {
			itemStrs := strToItems(tc.items)
			res := FuzzySearchBrute(itemStrs, tc.query, false, 3)
			resStrs := itemsToStr(res)

			if tc.expectedItems != nil {
				assert.ElementsMatch(t, *tc.expectedItems, resStrs, itemStrs, tc.query, resStrs)
			}

			for _, relOrder := range tc.expectedRelOrders { // for each expected order.
				assert.Len(t, relOrder, 2, "bad test case. expectedRelOrders should have 2 elements")
				assert.NotEqual(t, relOrder[0], relOrder[1], "bad test case. expectedRelOrders should have 2 different elements")
				// ensure the first item comes before the seoncd.
				var sawFirst, sawSecond bool
				for _, item := range resStrs { // scan the results.
					if item == relOrder[0] {
						sawFirst = true
					} else if item == relOrder[1] {
						sawSecond = true
					}
					if sawSecond && !sawFirst {
						assert.Fail(t, "expected the first item to come before the second",
							itemStrs, tc.query, resStrs, relOrder)
					}
				}
				if !(sawFirst && sawSecond) {
					assert.Fail(t, "expected to see both items", resStrs)
				}
			}
		}
	}
}
