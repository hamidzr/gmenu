package core

import (
	"context"
	"sync"

	"github.com/hamidzr/gmenu/constant"
	"github.com/hamidzr/gmenu/model"
)

type menu struct {
	items      []model.MenuItem
	query      string
	itemsMutex sync.Mutex
	ctx        context.Context
	queryMutex sync.Mutex
	ItemsChan  chan []model.MenuItem

	Filtered []model.MenuItem
	// zero-based index of the selected item in the filtered list
	Selected int
	// ResultText   string
	// MatchCount is the number of items that matched the search query.
	MatchCount    int
	SearchMethod  SearchMethod
	preserveOrder bool
	resultLimit   int
}

func newMenu(
	ctx context.Context,
	itemTitles []string,
	initValue string,
	searchMethod SearchMethod,
	preserveOrder bool,
) (*menu, error) {
	m := menu{
		ctx:           ctx,
		Selected:      0,
		SearchMethod:  searchMethod,
		resultLimit:   10,
		ItemsChan:     make(chan []model.MenuItem, 10), // bounded channel to prevent memory leaks
		query:         initValue,
		preserveOrder: preserveOrder,
	}
	items := m.titlesToMenuItem(itemTitles)

	if len(items) == 0 {
		items = []model.MenuItem{model.LoadingItem}
	}
	m.items = items

	m.Search(initValue)
	return &m, nil
}

// Filters the menu filtered list to only include items that match the keyword.
func (m *menu) Search(keyword string) {
    // Update the query under its own lock
    m.queryMutex.Lock()
    m.query = keyword
    m.queryMutex.Unlock()

    // Compute filtered results and update shared menu state under items lock
    m.itemsMutex.Lock()
    if keyword == "" {
        // Use current items snapshot directly
        m.Filtered = m.items
    } else {
        // Run search based on a stable snapshot of items
        itemsSnapshot := m.items
        m.Filtered = m.SearchMethod(itemsSnapshot, keyword, m.preserveOrder, 0)
    }
    if len(m.Filtered) > 0 {
        m.Selected = 0
    } else {
        m.Selected = constant.UnsetInt
    }
    m.MatchCount = len(m.Filtered)
    if len(m.Filtered) > m.resultLimit {
        m.Filtered = m.Filtered[:m.resultLimit]
    }
    m.itemsMutex.Unlock()
}

func (m *menu) titlesToMenuItem(titles []string) []model.MenuItem {
	items := make([]model.MenuItem, len(titles))
	for i, entry := range titles {
		items[i] = model.MenuItem{Title: entry}
	}
	return items
}
