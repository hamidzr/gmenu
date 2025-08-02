package core

import (
	"context"
	"fmt"
	"testing"
	"time"

	"fyne.io/fyne/v2/test"
	"github.com/hamidzr/gmenu/model"
	"github.com/stretchr/testify/require"
)

// TestDetectGUIHang is designed to detect when the GUI becomes unresponsive
// This test specifically recreates the scenario where the user reported
// "I saw apples in the list then it immeidately hung"
func TestDetectGUIHang(t *testing.T) {
	// Skip in CI environments where GUI cannot be displayed
	if testing.Short() {
		t.Skip("Skipping GUI hang detection test in short mode")
	}

	fmt.Println("Starting hang detection test...")
	fmt.Println("This test reproduces the exact scenario that caused the hang")

	// Use Fyne's test app instead of real GUI app
	app := test.NewApp()

	config := &model.Config{
		MenuID:    "hang-detection-test",
		Title:     "Hang Detection Test",
		Prompt:    "Select an item:",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenuWithApp(app, DirectSearch, config)
	require.NoError(t, err)
	defer cleanupGMenu(gmenu)

	// Set up menu with apple items (same as visual test that hung)
	testItems := []string{"apple", "banana", "cherry", "apricot", "avocado"}
	require.NoError(t, gmenu.SetupMenu(testItems, ""))

	// Create a responsiveness checker that monitors if the GUI is responding
	respChecker := NewResponsivenessChecker(gmenu, 1*time.Second)

	fmt.Println("1. Starting app in background...")

	// Start the app in a goroutine using test app
	appDone := make(chan error, 1)
	go func() {
		appDone <- gmenu.RunAppForever()
	}()

	// Wait for app to initialize
	time.Sleep(100 * time.Millisecond)

	fmt.Println("2. Showing GUI with apple items...")

	// Show the GUI - this should display the menu with items
	gmenu.ShowUI()

	fmt.Println("3. GUI should be visible now with apple items...")
	fmt.Println("4. Starting responsiveness monitoring...")

	// Start monitoring responsiveness
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	hangDetected := make(chan bool, 1)

	go func() {
		hung := respChecker.MonitorForHang(ctx)
		hangDetected <- hung
	}()

	// Wait a bit to let the GUI stabilize and be visible
	time.Sleep(100 * time.Millisecond)

	fmt.Println("5. Attempting to trigger the hang scenario...")

	// Try to recreate the exact scenario that caused the hang
	// This simulates what happens when a selection is attempted
	go func() {
		time.Sleep(50 * time.Millisecond)

		// This is the problematic code that was causing hangs
		fmt.Println("   Setting exit code (this was causing the hang)...")
		gmenu.SetExitCode(model.NoError)

		// The hang occurs because SetExitCode doesn't properly complete the selection
		// without calling markSelectionMade() which is unexported
		fmt.Println("   Hang should be detected now if the bug is present...")
	}()

	// Wait for either hang detection or test completion
	select {
	case hung := <-hangDetected:
		if hung {
			t.Errorf("GUI HANG DETECTED: The GUI became unresponsive during the test")
			fmt.Println("ERROR: GUI hang confirmed - this reproduces the user's issue")
			fmt.Println("       The SetExitCode() call without proper selection completion causes the hang")
		} else {
			fmt.Println("SUCCESS: No hang detected, GUI remained responsive")
		}
	case <-ctx.Done():
		fmt.Println("Test completed without hang detection")
	}

	// Try to clean up gracefully
	fmt.Println("6. Cleaning up...")
	gmenu.HideUI()
	app.Quit() // Use test app quit instead

	// Wait for app to finish
	select {
	case <-appDone:
		fmt.Println("App terminated successfully")
	case <-time.After(1 * time.Second):
		fmt.Println("Warning: App did not terminate cleanly")
	}
}

// ResponsivenessChecker monitors if the GUI is responding to state queries
type ResponsivenessChecker struct {
	gmenu    *GMenu
	interval time.Duration
}

func NewResponsivenessChecker(gmenu *GMenu, interval time.Duration) *ResponsivenessChecker {
	return &ResponsivenessChecker{
		gmenu:    gmenu,
		interval: interval,
	}
}

// MonitorForHang checks if the GUI stops responding to basic operations
func (rc *ResponsivenessChecker) MonitorForHang(ctx context.Context) bool {
	ticker := time.NewTicker(rc.interval)
	defer ticker.Stop()

	consecutiveFailures := 0
	maxFailures := 3 // Consider hung after 3 consecutive failures

	for {
		select {
		case <-ctx.Done():
			return false // Test completed, no hang detected
		case <-ticker.C:
			// Test basic responsiveness
			if rc.isResponsive() {
				consecutiveFailures = 0
				fmt.Printf("   GUI responsive check: OK (failures reset to 0)\n")
			} else {
				consecutiveFailures++
				fmt.Printf("   GUI responsive check: FAILED (failure count: %d/%d)\n", consecutiveFailures, maxFailures)

				if consecutiveFailures >= maxFailures {
					fmt.Printf("   HANG DETECTED: %d consecutive failures\n", consecutiveFailures)
					return true // Hang detected
				}
			}
		}
	}
}

// isResponsive tests if the GUI can respond to basic state queries
func (rc *ResponsivenessChecker) isResponsive() bool {
	// Use a timeout for each responsiveness check
	done := make(chan bool, 1)
	var success bool

	go func() {
		defer func() {
			// Recover from any panics during responsiveness check
			if r := recover(); r != nil {
				fmt.Printf("   Panic during responsiveness check: %v\n", r)
				success = false
			}
			done <- success
		}()

		// Try multiple basic operations that should work if GUI is responsive
		// Test 1: Try to access mutex-protected state safely
		rc.gmenu.uiMutex.Lock()
		isShown := rc.gmenu.isShown
		rc.gmenu.uiMutex.Unlock()

		// Test 2: Check if we can access the running state
		isRunning := rc.gmenu.isRunning

		// Test 3: Try to do a simple search operation
		_ = rc.gmenu.Search("test")

		// If we got here without hanging, the GUI is responsive
		_ = isShown
		_ = isRunning
		success = true
	}()

	// Wait for responsiveness check with timeout
	select {
	case result := <-done:
		return result
	case <-time.After(1 * time.Second):
		fmt.Printf("   Responsiveness check timed out\n")
		return false // Timeout indicates unresponsiveness
	}
}

// TestHangDetectionWithSimplifiedCase tests a simpler version without GUI complexity
func TestHangDetectionWithSimplifiedCase(t *testing.T) {
	fmt.Println("Testing hang detection with simplified scenario...")

	config := &model.Config{
		MenuID:    "simplified-hang-test",
		Title:     "Simplified Test",
		Prompt:    "test>",
		MinWidth:  300,
		MinHeight: 200,
	}

	gmenu, err := NewGMenu(DirectSearch, config)
	require.NoError(t, err)
	defer cleanupGMenu(gmenu)

	require.NoError(t, gmenu.SetupMenu([]string{"item1", "item2"}, ""))

	// Test that SetExitCode alone doesn't complete the selection properly
	fmt.Println("Setting exit code without proper selection completion...")
	gmenu.SetExitCode(model.NoError)

	// This should still work if the GUI is responsive
	results := gmenu.Search("item")
	require.Len(t, results, 2, "Search should still work after SetExitCode")

	fmt.Println("Simplified test completed - no hang detected")
}

func cleanupGMenu(gmenu *GMenu) {
	if gmenu != nil {
		// Try to clean up gracefully
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("Panic during cleanup: %v\n", r)
			}
		}()

		gmenu.HideUI()

		// Note: PID file cleanup is handled by the app itself
	}
}
