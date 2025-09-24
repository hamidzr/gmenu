package core

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/hamidzr/gmenu/model"
)

// SetItems sets the items to be displayed in the menu.
func (g *GMenu) SetItems(items []string, serializables []model.GmenuSerializable) {
	menuItems := g.menu.titlesToMenuItem(items)
	for i := range serializables {
		// avoid variable capture in loop by using index
		menuItems = append(menuItems, model.MenuItem{AType: &serializables[i]})
	}
	g.menu.ItemsChan <- menuItems
	if g.config.AutoAccept {
		go g.AttemptAutoSelect()
	}
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
	// TODO: add using SetItems?
	g.menu.ItemsChan <- newItems
}

// AttemptAutoSelect attempts to auto select if conditions are met.
func (g *GMenu) AttemptAutoSelect() {
	if !g.config.AutoAccept {
		return
	}

	const (
		autoSelectTimeout = 200 * time.Millisecond
		checkInterval     = 20 * time.Millisecond
	)

	ctx, cancel := context.WithTimeout(g.menu.ctx, autoSelectTimeout)
	defer cancel()

	ticker := time.NewTicker(checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if g.shouldAutoSelect() {
				g.ensureSelectionExitCode(model.NoError)
				g.markSelectionMade()
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// shouldAutoSelect checks if auto-selection conditions are met
func (g *GMenu) shouldAutoSelect() bool {
	g.menu.itemsMutex.Lock()
	defer g.menu.itemsMutex.Unlock()

	// need exactly one filtered item that isn't the loading placeholder
	return len(g.menu.Filtered) == 1 &&
		g.menu.Filtered[0].Title != model.LoadingItem.Title
}

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
	g.WaitForSelection()
	if g.exitCode == model.Unset {
		// this is a valid case in daemon mode.
	} else if g.exitCode != model.NoError {
		return nil, errors.Wrap(g.exitCode, "an error code is set")
	}
	// TODO: cli option for allowing query.
	if selected := g.selectedItem(); selected != nil {
		return selected, nil
	}
	if g.config.AcceptCustomSelection {
		return &model.MenuItem{Title: g.menu.query}, nil
	}
	return nil, model.ErrCustomUserEntry
}
