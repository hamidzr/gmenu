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
func IsDirectMatch(s string, keyword string, smartMatch bool) bool {
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

func FuzzySearchV2(items []model.MenuItem, query string, preserveOrder bool, limit int) []model.MenuItem {
	// var results []string
	matchedList := make([]model.MenuItem, 0)
	fmt.Println("Query, order, limit ", query, preserveOrder, limit)

	for _, item := range items {
		// distance := levenshtein.DistanceForStrings([]rune(query), []rune(item.Title), levenshtein.DefaultOptions)
		distance := calculateInsertions(query, item.Title)
		maxLen := max(len(query), len(item.Title))
		score := 100 - (distance * 100 / maxLen) // Convert distance to a similarity score

		if score > 0 { // You can adjust this threshold as needed
			fmt.Println(distance, " ", item.Title, " ", query, score)
			// result := fmt.Sprintf("%s (Score: %d%%)", item, score)
			matchedList = append(matchedList, item)
		}
	}

	return matchedList
}

// DirectSearch matches items directly to a keyword.
func DirectSearch(items []model.MenuItem, keyword string, _ bool, limit int) []model.MenuItem {
	matches := make([]model.MenuItem, 0)
	for _, item := range items {
		if IsDirectMatch(item.Title, keyword, true) {
			matches = append(matches, item)
		}
	}
	if limit == 0 {
		return matches
	}
	return matches[:min(limit, len(matches))]
}

// DirectSearchWithSeparator breaks down the keyword into subqueries.
func DirectSearchWithSeparator(separator string) SearchMethod {
	search := func(items []model.MenuItem, keyword string, _ bool, limit int) []model.MenuItem {
		// split keyword into words
		subQs := strings.Split(keyword, separator)
		newSubset := items // copy?
		// matches := make([]model.MenuItem, 0)
		for _, subQ := range subQs {
			newSubset = DirectSearch(newSubset, subQ, false, 0)
		}
		return newSubset[:min(limit, len(newSubset))]
	}
	return search
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
		entries[i] = item.Title
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
	"direct": DirectSearch,
	"fuzzy":  FuzzySearch,
	"fuzzy2": FuzzySearchV2,
}
