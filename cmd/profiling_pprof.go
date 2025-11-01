//go:build pprof

package main

import (
	"fmt"
	"os"
	"runtime/pprof"
)

func startProfiling() func() {
	f, err := os.Create("cpu.pprof")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create CPU profile: %v\n", err)
		return func() {}
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		fmt.Fprintf(os.Stderr, "could not start CPU profile: %v\n", err)
		_ = f.Close()
		return func() {}
	}

	return func() {
		pprof.StopCPUProfile()
		if err := f.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "could not close CPU profile: %v\n", err)
		}
	}
}
