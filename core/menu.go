package core

import (
	"errors"
	"sync"

	"github.com/hamidzr/gmenu/model"
)

type menu struct {
	items      []model.MenuItem
	query      string
	itemsMutex sync.Mutex
	queryMutex sync.Mutex
	ItemsChan  chan []model.MenuItem

	Filtered []model.MenuItem
	// zero-based index of the selected item in the filtered list
	Selected int
	// ResultText   string
	// MatchCount is the number of items that matched the search query.
	MatchCount    int
	SearchMethod  SearchMethod
	resultLimit   int
	preserveOrder bool
}

func newMenu(
	itemTitles []string,
	initValue string,
	searchMethod SearchMethod,
	preserveOrder bool,
) (*menu, error) {
	m := menu{
		Selected:      0,
		SearchMethod:  searchMethod,
		resultLimit:   10,
		ItemsChan:     make(chan []model.MenuItem),
		query:         initValue,
		preserveOrder: preserveOrder,
	}
	items := m.titlesToMenuItem(itemTitles)
	m.items = items

	if len(items) == 0 {
		return nil, errors.New("Menu must have at least one item")
	}

	m.Search(initValue)
	return &m, nil
}

// Filters the menu filtered list to only include items that match the keyword.
func (m *menu) Search(keyword string) {
	m.queryMutex.Lock()
	m.query = keyword
	m.queryMutex.Unlock()
	if keyword == "" {
		m.Filtered = m.items
	} else {
		// start := time.Now()
		m.Filtered = m.SearchMethod(m.items, keyword, m.preserveOrder, m.resultLimit)
		// elapsed := time.Since(start)
		// fmt.Println("Search took", elapsed)
	}
	if len(m.Filtered) > 0 {
		m.Selected = 0
	} else {
		m.Selected = unsetInt
	}
	m.MatchCount = len(m.Filtered)
	if len(m.Filtered) > m.resultLimit {
		m.Filtered = m.Filtered[:m.resultLimit]
	}
}

func (m *menu) titlesToMenuItem(titles []string) []model.MenuItem {
	items := make([]model.MenuItem, len(titles))
	for i, entry := range titles {
		items[i] = model.MenuItem{Title: entry}
	}
	return items
}
