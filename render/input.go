package render

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

// SearchEntry is a widget.Entry that captures certain key events.
type SearchEntry struct {
	widget.Entry
	OnKeyDown    func(key *fyne.KeyEvent)
	DisabledKeys map[fyne.KeyName]bool
}

// SelectAll selects all text in the entry.
func (e *SearchEntry) SelectAll() {
	// TODO: this cannot select anything outside non-alphanumeric characters.
	e.Entry.DoubleTapped(nil)
}

// TypedKey implements the fyne.TypedKeyReceiver interface.
func (e *SearchEntry) TypedKey(key *fyne.KeyEvent) {
	if e.OnKeyDown != nil {
		e.OnKeyDown(key)
	}
	if e.DisabledKeys != nil {
		if e.DisabledKeys[key.Name] {
			return
		}
	}
	e.Entry.TypedKey(key)
}
