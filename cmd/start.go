package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/GoLabra/labractl/internal/log"
)

// startCmd runs the LabraGo backend and frontend concurrently.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start both backend and frontend servers",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("🚦 Preparing LabraGo start...")

		// Ensure Yarn uses node-modules linker and no PnP in this workspace
		ensureYarnNodeModules(".")

		// Start backend and frontend processes
		backendCmd := exec.Command("go", "run", "main.go")
		backendCmd.Dir = filepath.Join("resources", "app")
		backendStdout, _ := backendCmd.StdoutPipe()
		backendStderr, _ := backendCmd.StderrPipe()

		frontendCmd := exec.Command("yarn", "dev")
		frontendCmd.Dir = filepath.Join("resources", "admin")
		frontendStdout, _ := frontendCmd.StdoutPipe()
		frontendStderr, _ := frontendCmd.StderrPipe()

		stream := func(prefix string, r io.Reader) {
			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.TrimSpace(line) == "" {
					continue
				}
				fmt.Printf("[%s] %s\n", prefix, line)
			}
		}

		if err := backendCmd.Start(); err != nil {
			log.Errorf("❌ Failed to start backend: %v", err)
			os.Exit(1)
		}
		go stream("backend", backendStdout)
		go stream("backend", backendStderr)

		if err := frontendCmd.Start(); err != nil {
			log.Errorf("❌ Failed to start frontend: %v", err)
			_ = backendCmd.Process.Kill()
			os.Exit(1)
		}
		go stream("frontend", frontendStdout)
		go stream("frontend", frontendStderr)

		log.Infof("🚀 Starting LabraGo backend + frontend")

		// Readiness checks
		type status struct {
			name string
			ok   bool
			addr string
		}
		results := make(chan status, 2)
		check := func(name, addr string) {
			deadline := time.Now().Add(60 * time.Second)
			for time.Now().Before(deadline) {
				conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
				if err == nil {
					_ = conn.Close()
					results <- status{name: name, ok: true, addr: addr}
					return
				}
				time.Sleep(1 * time.Second)
			}
			results <- status{name: name, ok: false, addr: addr}
		}
		go check("backend", "127.0.0.1:4000")
		go check("frontend", "127.0.0.1:3000")

		b := <-results
		f := <-results
		summarize := func(s status) string {
			if s.ok {
				return fmt.Sprintf("%s: started (%s)", s.name, s.addr)
			}
			return fmt.Sprintf("%s: not ready", s.name)
		}
		log.Infof("✅ Startup summary → %s | %s", summarize(b), summarize(f))

		backendErrCh := make(chan error, 1)
		frontendErrCh := make(chan error, 1)
		go func() { backendErrCh <- backendCmd.Wait() }()
		go func() { frontendErrCh <- frontendCmd.Wait() }()

		select {
		case err := <-backendErrCh:
			if err != nil {
				log.Errorf("❌ Backend exited: %v", err)
			} else {
				log.Infof("✅ Backend exited")
			}
		case err := <-frontendErrCh:
			if err != nil {
				log.Errorf("❌ Frontend exited: %v", err)
			} else {
				log.Infof("✅ Frontend exited")
			}
		}
	},
}

// init registers the start command with the root command.
func init() {
	rootCmd.AddCommand(startCmd)
}

// ensureYarnNodeModules makes sure Yarn uses the node-modules linker and disables PnP
func ensureYarnNodeModules(root string) {
	yrc := filepath.Join(root, ".yarnrc.yml")
	// Only write if missing
	if _, err := os.Stat(yrc); err != nil {
		_ = os.WriteFile(yrc, []byte("nodeLinker: node-modules\n"), 0644)
	}
	// Remove PnP file if present
	_ = os.Remove(filepath.Join(root, ".pnp.cjs"))
}
