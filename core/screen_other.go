//go:build !darwin

package core

// getScreenSizeMac returns fallback dimensions for non-macOS platforms
func getScreenSizeMac() (int, int) {
	// Fallback to common desktop resolution
	return 1920, 1080
}

// getLargestScreenSize returns fallback dimensions for non-macOS platforms
func getLargestScreenSize() (int, int) {
	// Fallback to common desktop resolution
	return 1920, 1080
}
