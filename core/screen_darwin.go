//go:build darwin

package core

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>

int getScreenWidth() {
	return (int)CGDisplayPixelsWide(CGMainDisplayID());
}

int getScreenHeight() {
	return (int)CGDisplayPixelsHigh(CGMainDisplayID());
}

// Get dimensions of the largest screen (for multi-monitor setups)
void getLargestScreenDimensions(int* width, int* height) {
	*width = 0;
	*height = 0;

	uint32_t displayCount;
	CGDirectDisplayID displays[32];

	if (CGGetActiveDisplayList(32, displays, &displayCount) == kCGErrorSuccess) {
		for (uint32_t i = 0; i < displayCount; i++) {
			int w = (int)CGDisplayPixelsWide(displays[i]);
			int h = (int)CGDisplayPixelsHigh(displays[i]);

			// Use the largest screen by area
			if (w * h > (*width) * (*height)) {
				*width = w;
				*height = h;
			}
		}
	}

	// Fallback to main display if no displays found
	if (*width == 0 || *height == 0) {
		*width = (int)CGDisplayPixelsWide(CGMainDisplayID());
		*height = (int)CGDisplayPixelsHigh(CGMainDisplayID());
	}
}
*/
import "C"

// getScreenSizeMac returns the dimensions of the main display
func getScreenSizeMac() (int, int) {
	return int(C.getScreenWidth()), int(C.getScreenHeight())
}

// getLargestScreenSize returns the dimensions of the largest screen in a multi-monitor setup
func getLargestScreenSize() (int, int) {
	var width, height C.int
	C.getLargestScreenDimensions(&width, &height)
	return int(width), int(height)
}
