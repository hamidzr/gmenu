package core

import (
	"fmt"
	"testing"

	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
			items:         unorderedInputs,
			query:         "ats ap",
			expectedItems: &[]string{"whats app"},
		},
	}

	// TODO: permuations instead of rand shuffling.
	for i := 0; i < 10; i++ { // for each order of inputs.
		unorderedInputs = shuffle(unorderedInputs)
		for _, tc := range testCases {
			fmt.Println("test case", tc)
			itemStrs := strToItems(tc.items)
			res := FuzzySearchBrute1(itemStrs, tc.query, false, 3)
			resStrs := itemsToStr(res)

			if tc.expectedItems != nil {
				assert.ElementsMatch(t, *tc.expectedItems, resStrs, itemStrs, tc.query, resStrs)
			}

			for _, relOrder := range tc.expectedRelOrders { // for each expected order.
				assert.Len(t, relOrder, 2, "bad test case. expectedRelOrders should have 2 elements")
				assert.NotEqual(t, relOrder[0], relOrder[1], "bad test case. expectedRelOrders should have 2 different elements")
				require.Greater(t, len(resStrs), 1, "need at least 2 results to compare order", resStrs)

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

func TestFuzzySearch(t *testing.T) {
	items := []model.MenuItem{
		{Title: "apple"},
		{Title: "banana"},
		{Title: "apricot"},
	}
	results := FuzzySearch(items, "ap", false, 10)
	expected := []string{"apple", "apricot"}
	var titles []string
	for _, item := range results {
		titles = append(titles, item.Title)
	}
	assert.Equal(t, expected, titles)
}

func TestFuzzySearchBrute(t *testing.T) {
	type testCase struct {
		name          string
		items         []string
		query         string
		limit         int
		expectedItems []string
	}

	testCases := []testCase{
		// { // TODO
		// 	name: "space should be exception for min consecutive chars",
		// 	items: []string{
		// 		"wha tsapp",
		// 	},
		// 	query: "atsa",
		// 	limit: 10,
		// 	expectedItems: []string{
		// 		"wha tsapp",
		// 	},
		// },
		{
			name: "min consecutive match of 2 chars should hold",
			items: []string{
				"whaxtsapp",
				"whatxsapp",
			},
			query: "atsa",
			limit: 10,
			expectedItems: []string{
				"whatxsapp",
			},
		},
		{
			name: "Query 'atsa' should prioritize 'whatsapp'",
			items: []string{
				"whaxtsapp",
				"whatsapp",
				"whats app",
			},
			query: "atsa",
			limit: 10,
			expectedItems: []string{
				"whatsapp",
				"whats app",
			},
		},
		{
			name: "Query 'atsap' should match all variants",
			items: []string{
				"whaxtsapp",
				"whats app",
				"whatsapp",
			},
			query: "atsap",
			limit: 10,
			expectedItems: []string{
				"whatsapp",
				"whats app",
			},
		},
		{
			name: "Query 'nonexistent' should return empty",
			items: []string{
				"whatsapp",
				"whaxtsapp",
				"whats app",
			},
			query:         "nonexistent",
			limit:         10,
			expectedItems: []string{},
		},
		{
			name: "single chars should not match. min 2",
			items: []string{
				"tg oday",
				"to---day",
				"xxtodayxx",
			},
			query: "today",
			limit: 10,
			expectedItems: []string{
				"xxtodayxx",
				"to---day",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			items := strToItems(&tc.items)
			results := fuzzySearchBruteConsec(items, tc.query, tc.limit, 2)
			resultTitles := itemsToStr(results)
			assert.Equal(t, tc.expectedItems, resultTitles)
		})
	}
}
