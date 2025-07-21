package cliutils

import "runtime"

// IsWindows reports whether the CLI is running on Windows.
var IsWindows = runtime.GOOS == "windows"

// Emoji returns the symbol if the terminal likely supports it, or a fallback
// ASCII representation on Windows where emoji support may be limited.
func Emoji(symbol, fallback string) string {
	if IsWindows {
		return fallback
	}
	return symbol
}
