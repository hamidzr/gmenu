package core

import "fmt"

// matchCounterLabel returns the label for the match counter.
func (g *GMenu) matchCounterLabel() string {
    // Snapshot current menu pointer, then lock its items
    g.menuMutex.RLock()
    m := g.menu
    g.menuMutex.RUnlock()
    if m == nil {
        return "[0/0]"
    }
    m.itemsMutex.Lock()
    matchCount := m.MatchCount
    total := len(m.items)
    m.itemsMutex.Unlock()
    return fmt.Sprintf("[%d/%d]", matchCount, total)
}
