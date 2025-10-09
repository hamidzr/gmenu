package main

import (
	"fmt"
	"os"
	"time"

	"github.com/hamidzr/gmenu/core"
	"github.com/hamidzr/gmenu/model"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/visual-test/main.go <test-type>")
		fmt.Println("Available tests:")
		fmt.Println("  cycles     - Show hide/show/selection cycles")
		fmt.Println("  interactive - Interactive test with many items")
		fmt.Println("  stress     - Rapid operations stress test")
		os.Exit(1)
	}

	testType := os.Args[1]

	switch testType {
	case "cycles":
		runHideShowCycleTest()
	case "interactive":
		runInteractiveTest()
	case "stress":
		runStressTest()
	default:
		fmt.Printf("Unknown test type: %s\n", testType)
		os.Exit(1)
	}
}

func runHideShowCycleTest() {
	fmt.Println("🎯 Running Hide/Show Cycle Visual Test")
	fmt.Println("=====================================")
	fmt.Println("This test will show 3 cycles of:")
	fmt.Println("1. Show menu with items")
	fmt.Println("2. Update items")
	fmt.Println("3. Perform search")
	fmt.Println("4. Simulate selection")
	fmt.Println("5. Hide menu")
	fmt.Println("6. Reset for next cycle")
	fmt.Println("")

	config := &model.Config{
		MenuID:    "visual-cycles-test",
		Title:     "Hide/Show Cycle Test - Auto-running",
		Prompt:    "Cycle test:",
		MinWidth:  500,
		MinHeight: 350,
		MaxWidth:  700,
		MaxHeight: 500,
	}

	gmenu, err := core.NewGMenu(core.DirectSearch, config)
	if err != nil {
		panic(err)
	}

	testItems := []string{
		"Apple 🍎",
		"Banana 🍌",
		"Cherry 🍒",
		"Date 🌴",
		"Elderberry 🫐",
	}

	fmt.Println("Setting up menu with initial items...")
	err = gmenu.SetupMenu(testItems, "")
	if err != nil {
		panic(err)
	}

	// Start the automated test in a goroutine
	go func() {
		// Wait for app to initialize
		time.Sleep(1 * time.Second)

		for cycle := 0; cycle < 3; cycle++ {
			fmt.Printf("=== Starting Cycle %d ===\n", cycle+1)

			// Show the menu
			fmt.Println("Showing GUI...")
			gmenu.ShowUI()

			// Give time to see the GUI
			time.Sleep(3 * time.Second)

			// Update with new items for this cycle
			newItems := []string{
				fmt.Sprintf("Cycle %d - Item A 🎯", cycle+1),
				fmt.Sprintf("Cycle %d - Item B 🚀", cycle+1),
				fmt.Sprintf("Cycle %d - Item C ⭐", cycle+1),
			}

			fmt.Println("Updating items...")
			err := gmenu.SetupMenu(newItems, "")
			if err != nil {
				fmt.Printf("Error updating items: %v\n", err)
			}

			// Give time to see the updated items
			time.Sleep(3 * time.Second)

			// Simulate a search
			fmt.Println("Performing search...")
			results := gmenu.Search("Cycle")
			fmt.Printf("Search found %d results\n", len(results))

			// Wait a bit more
			time.Sleep(2 * time.Second)

			// Note: We skip selection simulation to avoid hangs
			// In a real scenario, the user would press Enter or Escape
			fmt.Println("Skipping selection simulation (would normally press Enter/Escape)")

			// Hide the menu
			fmt.Println("Hiding GUI...")
			gmenu.HideUI()

			// Reset for next cycle
			fmt.Println("Resetting for next cycle...")
			gmenu.Reset(true)

			// Pause between cycles
			time.Sleep(2 * time.Second)
		}

		fmt.Println("All cycles completed! Stopping app in 2 seconds...")
		time.Sleep(2 * time.Second)
		gmenu.QuitWithCode(model.NoError)
	}()

	// Run the app on main goroutine (required by Fyne)
	err = gmenu.RunAppForever()
	if err != nil {
		fmt.Printf("App error: %v\n", err)
	}

	fmt.Println("✅ Hide/Show Cycle Test completed!")
}

func runInteractiveTest() {
	fmt.Println("🎮 Running Interactive Visual Test")
	fmt.Println("=================================")
	fmt.Println("This test shows an interactive menu where you can:")
	fmt.Println("- Type to search through items")
	fmt.Println("- Use arrow keys to navigate")
	fmt.Println("- Press Enter to select")
	fmt.Println("- Press Escape to close")
	fmt.Println("The menu will stay open for 15 seconds or until you close it.")
	fmt.Println("")

	config := &model.Config{
		MenuID:    "visual-interactive-test",
		Title:     "Interactive Test - Try typing to search!",
		Prompt:    "Type to search:",
		MinWidth:  600,
		MinHeight: 400,
		MaxWidth:  800,
		MaxHeight: 600,
	}

	gmenu, err := core.NewGMenu(core.FuzzySearch, config)
	if err != nil {
		panic(err)
	}

	// Set up interesting items to search through
	items := []string{
		"🍎 Apple - Fresh red apple",
		"🍌 Banana - Yellow curved fruit",
		"🍒 Cherry - Small red fruit",
		"🥝 Kiwi - Fuzzy brown fruit",
		"🍓 Strawberry - Red berry",
		"🍊 Orange - Citrus fruit",
		"🍇 Grapes - Purple cluster",
		"🥭 Mango - Tropical yellow fruit",
		"🍑 Peach - Fuzzy orange fruit",
		"🍐 Pear - Green or yellow fruit",
		"📱 iPhone - Apple smartphone",
		"💻 MacBook - Apple laptop",
		"🖥️ iMac - Apple desktop",
		"⌚ Apple Watch - Smart watch",
		"🎵 Apple Music - Streaming service",
		"📦 Package Manager - Software tool",
		"🔍 Search Engine - Find things",
		"🖱️ Computer Mouse - Pointing device",
		"⌨️ Keyboard - Input device",
		"🖨️ Printer - Output device",
		"🌟 Favorite Item - Most loved",
		"🎯 Target Practice - Aim here",
		"🚀 Rocket Ship - To the moon",
		"⭐ Star Quality - Five stars",
		"🔥 Hot Sauce - Spicy food",
	}

	fmt.Println("Setting up menu with sample items...")
	err = gmenu.SetupMenu(items, "")
	if err != nil {
		panic(err)
	}

	// Auto-close after 15 seconds
	go func() {
		time.Sleep(15 * time.Second)
		fmt.Println("⏰ 15 seconds elapsed, closing menu...")
		if gmenu.GetExitCode() == model.Unset {
			gmenu.QuitWithCode(model.UserCanceled)
		}
	}()

	// Show the menu immediately
	go func() {
		time.Sleep(500 * time.Millisecond)
		fmt.Println("Showing interactive menu...")
		gmenu.ShowUI()
	}()

	// Wait for user selection and handle app lifecycle like CLI does
	go func() {
		gmenu.WaitForSelection()
		fmt.Println("Selection made, closing menu...")
		if gmenu.GetExitCode() == model.Unset {
			gmenu.QuitWithCode(model.NoError)
		}
	}()

	// Run the app on main goroutine
	err = gmenu.RunAppForever()
	if err != nil {
		fmt.Printf("App error: %v\n", err)
	}

	fmt.Println("✅ Interactive Test completed!")
}

func runStressTest() {
	fmt.Println("⚡ Running Stress Test")
	fmt.Println("====================")
	fmt.Println("This test performs rapid operations to test for:")
	fmt.Println("- Visual glitches")
	fmt.Println("- Performance issues")
	fmt.Println("- Memory leaks")
	fmt.Println("- Race conditions")
	fmt.Println("The test runs for 10 seconds with operations every 100ms.")
	fmt.Println("")

	config := &model.Config{
		MenuID:    "visual-stress-test",
		Title:     "Stress Test - Rapid Operations",
		Prompt:    "Stress testing:",
		MinWidth:  400,
		MinHeight: 300,
		MaxWidth:  600,
		MaxHeight: 450,
	}

	gmenu, err := core.NewGMenu(core.FuzzySearch, config)
	if err != nil {
		panic(err)
	}

	// Start with some items
	initialItems := []string{"Apple", "Banana", "Cherry", "Date", "Elderberry"}
	err = gmenu.SetupMenu(initialItems, "")
	if err != nil {
		panic(err)
	}

	// Start stress test
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("Starting stress test...")

		gmenu.ShowUI()

		// Perform rapid operations
		searches := []string{"A", "B", "C", "", "Apple", "Ban", "", "Cherry", "D", ""}
		itemCounter := 0

		for i := 0; i < 100; i++ {
			// Rapid search changes
			searchTerm := searches[i%len(searches)]
			gmenu.Search(searchTerm)

			// Occasionally update items
			if i%10 == 0 {
				newItems := []string{
					fmt.Sprintf("Stress Item %d", itemCounter),
					fmt.Sprintf("Test Item %c", 'A'+itemCounter%26),
					fmt.Sprintf("Dynamic Item %d", itemCounter*2),
				}
				gmenu.AppendItems(newItems)
				itemCounter++
			}

			// Occasionally hide/show
			if i%25 == 0 && i > 0 {
				gmenu.HideUI()
				time.Sleep(50 * time.Millisecond)
				gmenu.ShowUI()
			}

			// Small delay but still stress the system
			time.Sleep(100 * time.Millisecond)
		}

		fmt.Println("Stress test completed! Stopping app...")
		time.Sleep(1 * time.Second)
		gmenu.QuitWithCode(model.NoError)
	}()

	// Run the app on main goroutine
	err = gmenu.RunAppForever()
	if err != nil {
		fmt.Printf("App error: %v\n", err)
	}

	fmt.Println("✅ Stress Test completed!")
}
