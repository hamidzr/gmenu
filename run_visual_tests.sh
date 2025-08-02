#!/bin/bash

# Script to run visual GUI tests that actually show the gmenu interface

echo "🎯 Running Visual GUI Tests for gmenu"
echo "======================================"
echo ""
echo "These tests will show the actual gmenu GUI interface."
echo "You'll see windows appearing and disappearing as the tests run."
echo ""

# Function to run a specific visual test
run_visual_test() {
    local test_name=$1
    local description=$2
    local executable_arg=$3
    
    echo "🔥 Running: $test_name"
    echo "📝 Description: $description"
    echo "⏳ Starting test..."
    echo ""
    
    if [ -n "$executable_arg" ]; then
        # Run the dedicated visual test executable
        go run cmd/visual-test/main.go "$executable_arg"
    else
        # Run the Go test (headless)
        go test ./core -run "$test_name" -v -timeout=30s
    fi
    
    if [ $? -eq 0 ]; then
        echo "✅ $test_name completed successfully!"
    else
        echo "❌ $test_name failed!"
    fi
    echo ""
    echo "----------------------------------------"
    echo ""
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    echo "❌ Error: Please run this script from the gmenu project root directory"
    exit 1
fi

echo "Choose which visual test to run:"
echo "1) Hide/Show Cycle Test (shows 3 automated cycles with visible GUI)"
echo "2) Interactive Test (15-second interactive session with real GUI)"
echo "3) Stress Test (rapid operations for 10 seconds with visible GUI)"
echo "4) All Visual Tests"
echo "5) Headless Tests (regular test suite without GUI)"
echo "6) Exit"
echo ""
read -p "Enter your choice (1-6): " choice

case $choice in
    1)
        run_visual_test "Hide/Show Cycle Test" "Tests 3 automated hide/show cycles with visible GUI" "cycles"
        ;;
    2)
        run_visual_test "Interactive Test" "Interactive session with real GUI - you can type and search!" "interactive"
        ;;
    3)
        run_visual_test "Stress Test" "Rapid operations with visible GUI to test performance" "stress"
        ;;
    4)
        echo "🚀 Running all visual tests..."
        echo ""
        run_visual_test "Hide/Show Cycle Test" "Tests 3 automated hide/show cycles with visible GUI" "cycles"
        echo ""
        run_visual_test "Interactive Test" "Interactive session with real GUI - you can type and search!" "interactive"  
        echo ""
        run_visual_test "Stress Test" "Rapid operations with visible GUI to test performance" "stress"
        echo "🎉 All visual tests completed!"
        ;;
    5)
        echo "🚀 Running headless tests (no GUI)..."
        echo ""
        run_visual_test "TestVisualGUIHideShowCycle" "Headless hide/show test" ""
        run_visual_test "TestVisualGUILongRunning" "Headless long running test" ""
        run_visual_test "TestVisualGUIStressTest" "Headless stress test" ""
        echo "🎉 All headless tests completed!"
        ;;
    6)
        echo "👋 Exiting..."
        exit 0
        ;;
    *)
        echo "❌ Invalid choice. Please run the script again."
        exit 1
        ;;
esac

echo ""
echo "🏁 Visual testing session completed!"
echo ""
echo "💡 Tips:"
echo "   • Run individual visual tests: go run cmd/visual-test/main.go cycles"
echo "   • Available visual tests: cycles, interactive, stress"
echo "   • Run headless tests: go test ./core -run TestVisualGUI -v"
echo "   • Run hang prevention tests: go test ./core -run TestWaitForSelection -v"