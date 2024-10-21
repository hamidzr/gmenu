package core

import "fmt"

// matchCounterLabel returns the label for the match counter.
func (g *GMenu) matchCounterLabel() string {
	return fmt.Sprintf("[%d/%d]", g.menu.MatchCount, len(g.menu.items))
}
