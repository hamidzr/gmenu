package core

import (
	"fmt"

	"github.com/hamidzr/gmenu/constant"
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
	g.menu.ItemsChan <- menuItems
	g.menu.itemsMutex.Unlock()
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

// PrependItems adds items to the beginning of the menu.
func (g *GMenu) PrependItems(items []string) {
	g.addItems(items, false)
}

// AppendItems adds items to the end of the menu.
func (g *GMenu) AppendItems(items []string) {
	// fmt.Println("appending len items", len(items))
	g.addItems(items, true)
}

// SelectedValue returns the selected item.
func (g *GMenu) SelectedValue() (*model.MenuItem, error) {
	g.SelectionWg.Wait()
	if g.ExitCode == constant.UnsetInt {
		// this is a valid case in daemon mode.
	} else if g.ExitCode != 0 {
		return nil, fmt.Errorf("gmenu exited with code %d", g.ExitCode)
	}
	// TODO: cli option for allowing query.
	if g.menu.Selected >= 0 && g.menu.Selected < len(g.menu.Filtered)+1 {
		selected := g.menu.Filtered[g.menu.Selected]
		return &selected, nil
	}
	return &model.MenuItem{Title: g.menu.query}, nil
}
