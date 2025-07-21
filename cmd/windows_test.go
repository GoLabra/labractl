//go:build windows

package cmd

import "testing"

func TestGetInstallCommandWindows(t *testing.T) {
	if c := getInstallCommand("git", "windows"); c == nil {
		t.Fatal("expected command for windows")
	}
}
