package core

import (
	"github.com/pkg/errors"

	"github.com/hamidzr/gmenu/model"
)

// SetItems sets the items to be displayed in the menu.
func (g *GMenu) SetItems(items []string, serializables []model.GmenuSerializable) {
	menuItems := g.menu.titlesToMenuItem(items)
	for _, item := range serializables {
		myItem := item
		menuItems = append(menuItems, model.MenuItem{AType: &myItem})
	}
	g.menu.itemsMutex.Lock()
	defer g.menu.itemsMutex.Unlock()
	g.menu.ItemsChan <- menuItems
}

// addItems adds items to the menu.
func (g *GMenu) addItems(items []string, tail bool) {
	newMenuItems := g.menu.titlesToMenuItem(items)
	g.menu.itemsMutex.Lock()
	var newItems []model.MenuItem
	if tail {
		newItems = append(g.menu.items, newMenuItems...)
	} else {
		newItems = append(newMenuItems, g.menu.items...)
	}
	g.menu.itemsMutex.Unlock()
	g.menu.ItemsChan <- newItems
	// TODO: add using SetItems?
}

// // AttemptAutoSelect attempts to auto select if conditions are met.
// func (g *GMenu) AttemptAutoSelect() {
// 	if g.config.InitialQuery == "" {
// 		return
// 	}
// 	// items might not be loaded needs to lock menu.items access.?
// 	// g.menu.Search(g.config.InitialQuery)
// 	// fmt.Println(g.menu.items)
// 	if g.config.AutoAccept && len(g.menu.Filtered) == 1 {
// 		g.selectionMade()
// 	}
// }

// PrependItems adds items to the beginning of the menu.
func (g *GMenu) PrependItems(items []string) {
	g.addItems(items, false)
}

// AppendItems adds items to the end of the menu.
func (g *GMenu) AppendItems(items []string) {
	// fmt.Println("appending len items", len(items))
	g.addItems(items, true)
}

// selectedItem returns the selected item if in bound or nil.
func (g *GMenu) selectedItem() *model.MenuItem {
	if g.menu.Selected >= 0 && g.menu.Selected < len(g.menu.Filtered) {
		selected := g.menu.Filtered[g.menu.Selected]
		return &selected
	}
	return nil
}

// SelectedValue returns the selected item.
// TODO: support for context cancellations.
func (g *GMenu) SelectedValue() (*model.MenuItem, error) {
	g.SelectionWg.Wait()
	if g.ExitCode == model.Unset {
		// this is a valid case in daemon mode.
	} else if g.ExitCode != model.NoError {
		return nil, errors.Wrap(g.ExitCode, "an error code is set")
	}
	// TODO: cli option for allowing query.
	if selected := g.selectedItem(); selected != nil {
		return selected, nil
	}
	if g.config.AcceptCustomSelection {
		return &model.MenuItem{Title: g.menu.query}, nil
	}
	return nil, model.CustomUserEntry
}
