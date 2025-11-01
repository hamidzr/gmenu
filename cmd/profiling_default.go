//go:build !pprof

package main

func startProfiling() func() {
	return func() {}
}
