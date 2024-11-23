package model

// GmenuSerializable defines the interface for any item that can be serialized for gmenu.
type GmenuSerializable interface {
	// Serialize returns a string representation of the item that can be shown in gmenu.
	// implemented by value type.
	Serialize() string // rename String?
}

// MenuItem represents an item in the menu.
type MenuItem struct {
	Title string
	AType *GmenuSerializable
	Score int
}

// ComputedTitle returns the title of the menu item.
func (m *MenuItem) ComputedTitle() string {
	if m.Title != "" {
		return m.Title
	}
	if m.AType != nil {
		return (*m.AType).Serialize()
	}
	return ""
}

// Serialize implements GmenuSerializable for MenuItem.
// CHECK: is it accurate? why do we have the separation here.
// func (m MenuItem) Serialize() string {
// 	return m.ComputedTitle()
// }

// TestSerializable is a test type that implements GmenuSerializable.
type TestSerializable struct{}

// Serialize returns a string representation of the item that can be shown in gmenu.
func (t TestSerializable) Serialize() string {
	return "TestSerializable"
}
