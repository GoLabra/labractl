package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/GoLabra/labractl/internal/cliutils"
)

// startCmd runs the LabraGo backend and frontend concurrently.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start both backend and frontend servers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cliutils.Emoji("🚦", "->"), "Preparing LabraGo start...")

		root := "."
		packageJsonPath := filepath.Join(root, "package.json")

		// 1. If no package.json, run `yarn init -y`
		if _, err := os.Stat(packageJsonPath); err != nil {
			fmt.Println(cliutils.Emoji("📦", "[pkg]"), "No package.json found. Initializing Yarn project...")
			if err := cliutils.RunCommand("yarn", []string{"init", "-y"}, root); err != nil {
				fmt.Println(cliutils.Emoji("❌", "X"), "Failed to initialize Yarn project:", err)
				os.Exit(1)
			}
		}

		// 2. Read + parse package.json
		data, err := os.ReadFile(packageJsonPath)
		if err != nil {
			fmt.Println(cliutils.Emoji("❌", "X"), "Failed to read package.json:", err)
			os.Exit(1)
		}

		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err != nil {
			fmt.Println(cliutils.Emoji("❌", "X"), "Failed to parse package.json:", err)
			os.Exit(1)
		}

		// 3. Add missing scripts
		backendPath := filepath.Join("src", "app")
		frontendPath := filepath.Join("src", "admin")
		scripts := map[string]string{
			"start":          "concurrently \"yarn start:backend\" \"yarn start:frontend\"",
			"start:backend":  fmt.Sprintf("cd %s && go run main.go start", backendPath),
			"start:frontend": fmt.Sprintf("cd %s && yarn dev", frontendPath),
		}
		modified := false
		if pkg["scripts"] == nil {
			pkg["scripts"] = map[string]interface{}{}
			modified = true
		}
		s := pkg["scripts"].(map[string]interface{})
		for k, v := range scripts {
			if _, ok := s[k]; !ok {
				s[k] = v
				modified = true
			}
		}
		if modified {
			newData, _ := json.MarshalIndent(pkg, "", "  ")
			if err := os.WriteFile(packageJsonPath, newData, 0644); err != nil {
				fmt.Println(cliutils.Emoji("❌", "X"), "Failed to update package.json:", err)
				os.Exit(1)
			}
			fmt.Println(cliutils.Emoji("🛠", "[update]"), "package.json updated with start scripts.")
		}

		// 4. Check/install concurrently
		if err := exec.Command("yarn", "list", "--pattern", "concurrently").Run(); err != nil {
			fmt.Println(cliutils.Emoji("📦", "[pkg]"), "Installing concurrently...")
			if err := cliutils.RunCommand("yarn", []string{"add", "concurrently", "--dev"}, root); err != nil {
				fmt.Println(cliutils.Emoji("❌", "X"), "Failed to install concurrently:", err)
				os.Exit(1)
			}
		}

		// 5. Run yarn start
		fmt.Println(cliutils.Emoji("🚀", "->"), "Starting LabraGo backend + frontend")
		run := exec.Command("yarn", "start")
		run.Stdout = os.Stdout
		run.Stderr = os.Stderr
		run.Stdin = os.Stdin
		if err := run.Run(); err != nil {
			fmt.Println(cliutils.Emoji("❌", "X"), "Failed to run yarn start:", err)
			os.Exit(1)
		}
	},
}

// init registers the start command with the root command.
func init() {
	rootCmd.AddCommand(startCmd)
}
