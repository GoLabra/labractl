package cmd

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/GoLabra/labractl/internal/cliutils"
	"github.com/GoLabra/labractl/internal/log"
)

// startCmd runs the LabraGo backend and frontend concurrently.
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start both backend and frontend servers",
	Run: func(cmd *cobra.Command, args []string) {
		log.Infof("🚦 Preparing LabraGo start...")

		root := "."
		packageJsonPath := filepath.Join(root, "package.json")

		// 1. If no package.json, run `yarn init -y`
		if _, err := os.Stat(packageJsonPath); err != nil {
			log.Infof("📦 No package.json found. Initializing Yarn project...")
			if err := cliutils.RunCommand("yarn", []string{"init", "-y"}, root); err != nil {
				log.Errorf("❌ Failed to initialize Yarn project: %v", err)
				os.Exit(1)
			}
		}

		// 2. Read + parse package.json
		data, err := os.ReadFile(packageJsonPath)
		if err != nil {
			log.Errorf("❌ Failed to read package.json: %v", err)
			os.Exit(1)
		}

		var pkg map[string]interface{}
		if err := json.Unmarshal(data, &pkg); err != nil {
			log.Errorf("❌ Failed to parse package.json: %v", err)
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
				log.Errorf("❌ Failed to update package.json: %v", err)
				os.Exit(1)
			}
			log.Infof("🛠 package.json updated with start scripts.")
		}

		// 4. Check/install concurrently
		if err := exec.Command("yarn", "list", "--pattern", "concurrently").Run(); err != nil {
			log.Infof("📦 Installing concurrently...")
			if err := cliutils.RunCommand("yarn", []string{"add", "concurrently", "--dev"}, root); err != nil {
				log.Errorf("❌ Failed to install concurrently: %v", err)
				os.Exit(1)
			}
		}

		// 5. Run yarn start
		log.Infof("🚀 Starting LabraGo backend + frontend")
		run := exec.Command("yarn", "start")
		run.Stdout = os.Stdout
		run.Stderr = os.Stderr
		run.Stdin = os.Stdin
		if err := run.Run(); err != nil {
			log.Errorf("❌ Failed to run yarn start: %v", err)
			os.Exit(1)
		}
	},
}

// init registers the start command with the root command.
func init() {
	rootCmd.AddCommand(startCmd)
}
