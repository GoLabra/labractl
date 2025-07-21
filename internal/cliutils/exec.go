// Package cliutils contains helper utilities shared across CLI commands.
package cliutils

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

// RunCommand executes a command with the provided arguments inside a
// working directory. Stdout and stderr are attached to the current
// process so output streams to the user.
func RunCommand(bin string, args []string, dir string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ReadLine reads a single line from STDIN and trims any trailing
// newline characters. It is used when prompting the user for input.
func ReadLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}
