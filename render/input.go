package render

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// SearchEntry is a widget.Entry that captures certain key events.
type SearchEntry struct {
	widget.Entry
	OnKeyDown            func(key *fyne.KeyEvent)
	PropagationBlacklist map[fyne.KeyName]bool
	OnFocusLost          func()
}

// SelectAll selects all text in the entry.
func (e *SearchEntry) SelectAll() {
	// TODO: this cannot select anything outside non-alphanumeric characters.
	e.DoubleTapped(nil)
}

// AcceptsTab implements the fyne.Tabbable interface.
func (e *SearchEntry) AcceptsTab() bool {
	return true
}

// TypedKey implements the fyne.TypedKeyReceiver interface.
func (e *SearchEntry) TypedKey(key *fyne.KeyEvent) {
	if e.OnKeyDown != nil {
		e.OnKeyDown(key)
	}
	if e.PropagationBlacklist != nil {
		if e.PropagationBlacklist[key.Name] {
			return
		}
	}
	e.Entry.TypedKey(key)
}

// TypedShortcut implements the fyne.TypedShortcutReceiver interface.
func (e *SearchEntry) TypedShortcut(shortcut fyne.Shortcut) {
	s, ok := shortcut.(*desktop.CustomShortcut)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}
	if s.Mod() == fyne.KeyModifierControl && s.Key() == fyne.KeyL {
		e.SetText("")
	}
}

// FocusLost implements the fyne.Focusable interface.
func (e *SearchEntry) FocusLost() {
	if e.OnFocusLost != nil {
		e.OnFocusLost()
	}
	e.Entry.FocusLost()
}

// NewInputArea returns a horizontal container with input widgets.
func NewInputArea(searchEntry *SearchEntry, matchLabel *widget.Label) *fyne.Container {
	// style the search entry with better visual design
	searchEntry.PlaceHolder = "Type to search..."

	cont := container.NewStack(
		searchEntry,
		matchLabel,
	)
	cont.Layout = NewProportionalLayout(100)
	return cont
}
