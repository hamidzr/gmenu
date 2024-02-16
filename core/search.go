package core

import (
	"sort"
	"strings"

	"github.com/hamidzr/gmenu/model"
	"github.com/sahilm/fuzzy"
)

// SearchMethod how to search for items given a keyword.
type SearchMethod func(items []model.MenuItem, query string,
	preserveOrder bool, limit int) []model.MenuItem

func isDirectMatch(item model.MenuItem, keyword string) bool {
	return strings.Contains(strings.ToLower(item.Title), strings.ToLower(keyword))
}

// DirectSearch matches items directly to a keyword.
func DirectSearch(items []model.MenuItem, keyword string, _ bool, limit int) []model.MenuItem {
	matches := make([]model.MenuItem, 0)
	for _, item := range items {
		if isDirectMatch(item, keyword) {
			matches = append(matches, item)
		}
	}
	return matches[:min(limit, len(matches))]
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
	preserveOrder bool, limit int) []model.MenuItem {
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

var SearchMethods = map[string]SearchMethod{
	"direct": DirectSearch,
	"fuzzy":  FuzzySearch,
}
