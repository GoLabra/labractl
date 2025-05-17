package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start both backend and frontend servers",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚦 Preparing LabraGo start...")

		root := "."
		packageJsonPath := filepath.Join(root, "package.json")

		// 1. If no package.json, run `yarn init -y`
		if _, err := os.Stat(packageJsonPath); err != nil {
			fmt.Println("📦 No package.json found. Initializing Yarn project...")
			if err := runCommand("yarn", []string{"init", "-y"}, root); err != nil {
				fmt.Println("❌ Failed to initialize Yarn project:", err)
				os.Exit(1)
			}
		}

		// 2. Read + parse package.json
		data, err := os.ReadFile(packageJsonPath)
		if err != nil {
			fmt.Println("❌ Failed to read package.json:", err)
			os.Exit(1)
		}

		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err != nil {
			fmt.Println("❌ Failed to parse package.json:", err)
			os.Exit(1)
		}

		// 3. Add missing scripts
		scripts := map[string]string{
			"start":          "concurrently \"yarn start:backend\" \"yarn start:frontend\"",
			"start:backend":  "cd src/app && go run main.go start",
			"start:frontend": "cd src/admin && yarn dev",
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
				fmt.Println("❌ Failed to update package.json:", err)
				os.Exit(1)
			}
			fmt.Println("🛠 package.json updated with start scripts.")
		}

		// 4. Check/install concurrently
		if err := exec.Command("yarn", "list", "--pattern", "concurrently").Run(); err != nil {
			fmt.Println("📦 Installing concurrently...")
			if err := runCommand("yarn", []string{"add", "concurrently", "--dev"}, root); err != nil {
				fmt.Println("❌ Failed to install concurrently:", err)
				os.Exit(1)
			}
		}

		// 5. Run yarn start
		fmt.Println("🚀 Starting LabraGo backend + frontend")
		run := exec.Command("yarn", "start")
		run.Stdout = os.Stdout
		run.Stderr = os.Stderr
		run.Stdin = os.Stdin
		if err := run.Run(); err != nil {
			fmt.Println("❌ Failed to run yarn start:", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
