package core

import (
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/require"
)

func newTestGMenu(t *testing.T) *GMenu {
	app := test.NewApp()
	config := &model.Config{
		MenuID:    "focus-loss-test",
		Title:     "Focus Loss Test",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenuWithApp(app, DirectSearch, config)
	require.NoError(t, err)
	t.Cleanup(func() {
		if gmenu.menuCancel != nil {
			gmenu.menuCancel()
		}
	})
	return gmenu
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool) {
	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return
		}
		if time.Now().After(deadline) {
			require.FailNow(t, "condition not met before timeout")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestFocusLossCancelsWhenNoSelection(t *testing.T) {
	gmenu := newTestGMenu(t)
	defer cleanupGMenu(gmenu)

	require.NoError(t, gmenu.SetupMenu([]string{"alpha", "beta"}, ""))

	// Trigger focus loss without making a selection
	gmenu.ui.SearchEntry.OnFocusLost()

	waitForCondition(t, 200*time.Millisecond, func() bool {
		return gmenu.GetExitCode() == model.UserCanceled && gmenu.selectionFuse.IsBroken()
	})
}

func TestHandleItemClickBeatsFocusLossCancellation(t *testing.T) {
	gmenu := newTestGMenu(t)
	defer cleanupGMenu(gmenu)

	require.NoError(t, gmenu.SetupMenu([]string{"alpha", "beta"}, ""))

	// Simulate focus loss firing before the item click handler runs
	gmenu.ui.SearchEntry.OnFocusLost()
	gmenu.handleItemClick(0)

	waitForCondition(t, 200*time.Millisecond, func() bool {
		return gmenu.GetExitCode() == model.NoError && gmenu.selectionFuse.IsBroken()
	})
}
