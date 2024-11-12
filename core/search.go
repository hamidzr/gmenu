package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/hamidzr/gmenu/model"
	"github.com/sahilm/fuzzy"
)

// SearchMethod how to search for items given a keyword.
type SearchMethod func(items []model.MenuItem, query string,
	preserveOrder bool, limit int) []model.MenuItem

// IsDirectMatch checks if a string contains a keyword.
func IsDirectMatch(s, keyword string, smartMatch bool) bool {
	if smartMatch && strings.ToLower(keyword) != keyword {
		return strings.Contains(s, keyword)
	}
	return strings.Contains(strings.ToLower(s), strings.ToLower(keyword))
}

// calculateInsertions calculates the minimum number of insertions needed
// to transform str1 into str2. It assumes only insertions are allowed.
func calculateInsertions(str1, str2 string) int {
	// Convert strings to rune slices for proper handling of Unicode characters
	runes1, runes2 := []rune(str1), []rune(str2)
	len1, len2 := len(runes1), len(runes2)

	if len1 > len2 {
		// If str1 is longer, transformation isn't possible with only insertions
		return -1
	}

	insertions := 0
	i, j := 0, 0

	for i < len1 && j < len2 {
		if runes1[i] == runes2[j] {
			i++
			j++
		} else {
			insertions++
			j++
		}
	}

	// Add remaining characters in str2 to the count of insertions
	if j < len2 {
		insertions += len2 - j
	}

	return insertions
}

func fuzzyContainsConsec(s, query string, ignoreCase bool, minConsecutive int) bool {
	if ignoreCase {
		s, query = strings.ToLower(s), strings.ToLower(query)
	}

	if len(query) < minConsecutive {
		return false
	}

	// Iterate through 's' to find at least 'minConsecutive' consecutive matching characters
	for i := 0; i <= len(s)-minConsecutive; i++ {
		if s[i:i+minConsecutive] == query[0:minConsecutive] {
			// Found the starting point with 'minConsecutive' matching characters
			queryIndex := minConsecutive
			// Continue matching the rest of the query (non-consecutively)
			for j := i + minConsecutive; j < len(s) && queryIndex < len(query); j++ {
				if s[j] == query[queryIndex] {
					queryIndex++
				}
			}
			if queryIndex == len(query) {
				return true
			}
		}
	}
	return false
}

// fuzzyContains checks if all characters in the query exist in the string in order.
// TODO: compute how many chars/ratio violate this?
func fuzzyContains(s, query string, ignoreCase bool) bool {
	return fuzzyContainsConsec(s, query, ignoreCase, 1)
}

// calculates the ratio of the query chars that exists in s in anyorder.
// func matching(s, query string) float32 {
// }

// FuzzySearchV2 is a fuzzy search that uses a different scoring mechanism.
func FuzzySearchV2(items []model.MenuItem, query string, preserveOrder bool, limit int) []model.MenuItem {
	// var results []string
	matchedList := make([]model.MenuItem, 0)
	fmt.Println("Query, order, limit ", query, preserveOrder, limit)

	for _, item := range items {
		// distance := levenshtein.DistanceForStrings([]rune(query), []rune(item.Title), levenshtein.DefaultOptions)
		distance := calculateInsertions(query, item.ComputedTitle())
		maxLen := max(len(query), len(item.ComputedTitle()))
		score := 100 - (distance * 100 / maxLen) // Convert distance to a similarity score

		if score > 0 { // You can adjust this threshold as needed
			fmt.Println(distance, " ", item.ComputedTitle(), " ", query, score)
			// result := fmt.Sprintf("%s (Score: %d%%)", item, score)
			matchedList = append(matchedList, item)
		}
	}

	return matchedList
}

func applyLimit(matches []model.MenuItem, limit int) []model.MenuItem {
	if limit == 0 {
		return matches
	}
	return matches[:min(limit, len(matches))]
}

// DirectSearch matches items directly to a keyword.
func DirectSearch(items []model.MenuItem, keyword string, _ bool, limit int) []model.MenuItem {
	matches := make([]model.MenuItem, 0)
	for _, item := range items {
		if IsDirectMatch(item.ComputedTitle(), keyword, true) {
			matches = append(matches, item)
		}
	}
	return applyLimit(matches, limit)
}

func fuzzySearchBruteConsec(items []model.MenuItem, keyword string, limit int, consecutive int) []model.MenuItem {
	if keyword == "" {
		return items
	}
	direcMatches := make([]model.MenuItem, 0)
	fuzzyMatches := make([]model.MenuItem, 0)
	for _, item := range items {
		if IsDirectMatch(item.ComputedTitle(), keyword, true) {
			direcMatches = append(direcMatches, item)
		} else if fuzzyContainsConsec(item.ComputedTitle(), keyword, true, consecutive) {
			fuzzyMatches = append(fuzzyMatches, item)
		}
	}
	return applyLimit(append(direcMatches, fuzzyMatches...), limit)
}

// FuzzySearchBrute is a brute force fuzzy search.
// Direct matches are prioritized over fuzzy matches.
func FuzzySearchBrute1(items []model.MenuItem, keyword string, _ bool, limit int) []model.MenuItem {
	return fuzzySearchBruteConsec(items, keyword, limit, 1)
}

// FuzzySearchBrute is a brute force fuzzy search.
// Direct matches are prioritized over fuzzy matches.
// A minimum of 2 consecutive characters are required for a fuzzy match.
func FuzzySearchBrute(items []model.MenuItem, keyword string, _ bool, limit int) []model.MenuItem {
	return fuzzySearchBruteConsec(items, keyword, limit, 2)
}

// SearchWithSeparator breaks down the keyword into subqueries.
func SearchWithSeparator(separator string, searchMethod SearchMethod) SearchMethod {
	search := func(items []model.MenuItem, query string, preserveOrder bool, limit int) []model.MenuItem {
		// split keyword into words
		subQs := strings.Split(query, separator)
		matchedSubset := items // copy?
		// matches := make([]model.MenuItem, 0)
		for _, subQ := range subQs {
			// logrus.Trace("Subquery: ", subQ, "subset", matchedSubset)
			matchedSubset = searchMethod(matchedSubset, subQ, false, 0)
		}
		return matchedSubset[:min(limit, len(matchedSubset))]
	}
	return search
}

// DirectSearchWithSeparator is a direct search with a separator.
func DirectSearchWithSeparator(separator string) SearchMethod {
	return SearchWithSeparator(separator, DirectSearch)
}

// filterOutUnlikelyMatches takes in a sorted list of fuzzy matches and returns
// a list of matches with scores greater than 0 if there's any, otherwise
// returns the original list.
func filterOutUnlikelyMatches(matches []fuzzy.Match) []fuzzy.Match {
	if len(matches) == 0 {
		return matches
	}
	if matches[0].Score <= 0 {
		return matches
	}

	positiveScores := make([]fuzzy.Match, 0)
	for _, match := range matches {
		if match.Score > 0 {
			positiveScores = append(positiveScores, match)
		}
	}
	return positiveScores
}

// FuzzySearch fuzzy matches items to a keyword and sorts them by score.
func FuzzySearch(items []model.MenuItem, keyword string,
	preserveOrder bool, limit int,
) []model.MenuItem {
	entries := make([]string, len(items))
	for i, item := range items {
		entries[i] = item.ComputedTitle()
	}

	matches := fuzzy.Find(keyword, entries)
	results := make([]model.MenuItem, 0)
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Index < matches[j].Index
		}
		return matches[i].Score > matches[j].Score
	})
	matches = matches[:min(limit, len(matches))]
	matches = filterOutUnlikelyMatches(matches)
	if !preserveOrder {
		for _, match := range matches {
			results = append(results, items[match.Index])
		}
		return results
	}
	matchIndices := make([]int, 0)
	for _, match := range matches {
		matchIndices = append(matchIndices, match.Index)
	}
	sort.Slice(matchIndices, func(i, j int) bool {
		return matchIndices[i] < matchIndices[j]
	})
	for _, ogIndex := range matchIndices {
		results = append(results, items[ogIndex])
	}

	return results
}

// SearchMethods is a map of search methods.
var SearchMethods = map[string]SearchMethod{
	"direct":  DirectSearch,
	"fuzzy":   SearchWithSeparator(" ", FuzzySearchBrute),
	"fuzzy1":  FuzzySearch,
	"fuzzy2":  FuzzySearchV2,
	"fuzzy3":  FuzzySearchBrute,
	"default": SearchWithSeparator(" ", FuzzySearchBrute),
}
