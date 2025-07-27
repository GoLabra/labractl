package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GoLabra/labractl/internal/cliutils"
	"github.com/GoLabra/labractl/internal/log"
)

var autoYes bool

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new LabraGo project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		repoURL := "https://github.com/GoLabra/labra"
		log.Infof("🚀 Creating LabraGo project: %s", projectName)

		// 1. Check prerequisites
		checkPrerequisites()

		// 2. Choose package manager
		packageManager := choosePackageManager()

		// 3. Clone repo
		if err := cliutils.RunCommand("git", []string{"clone", repoURL, projectName}, ""); err != nil {
			log.Errorf("❌ Git clone failed: %v", err)
			os.Exit(1)
		}

		// 4. Patch go.mod
		goModPath := filepath.Join(projectName, "src", "app", "go.mod")
		if err := patchGoMod(goModPath); err != nil {
			log.Errorf("❌ go.mod patch failed: %v", err)
			os.Exit(1)
		}

		// 5. Create .env files
		if err := createAppEnvFile(projectName); err != nil {
			log.Errorf("❌ Backend .env failed: %v", err)
			os.Exit(1)
		}
		if err := createAdminEnvFile(projectName); err != nil {
			log.Errorf("❌ Frontend .env failed: %v", err)
			os.Exit(1)
		}

		// 6. Go mod tidy + generate
		appPath := filepath.Join(projectName, "src", "app")
		_ = cliutils.RunCommand("go", []string{"mod", "tidy"}, appPath)
		if err := cliutils.RunCommand("go", []string{"generate", "./..."}, appPath); err != nil {
			log.Warnf("⚠️ go generate failed, retrying...")
			_ = cliutils.RunCommand("go", []string{"mod", "tidy"}, appPath)
			_ = cliutils.RunCommand("go", []string{"generate", "./..."}, appPath)
		}

		// 7. Frontend install
		adminPath := filepath.Join(projectName, "src", "admin")
		if _, err := os.Stat(filepath.Join(adminPath, "package.json")); err == nil {
			log.Infof("📦 Installing frontend dependencies with %s...", packageManager)
			_ = cliutils.RunCommand(packageManager, []string{"install"}, adminPath)
		}

		// 8. Ensure PostgreSQL
		if err := ensurePostgresUserAndDatabase(projectName); err != nil {
			log.Warnf("⚠️ PostgreSQL setup failed: %v", err)
		}

		// 9. Done
		log.Infof("✅ Project created at %s", projectName)
		log.Infof("👉 cd %s\nlabractl start", projectName)
	},
}

// patchGoMod updates the replace directive in go.mod to point
// to the local LabraGo API for development purposes.
func patchGoMod(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := strings.Replace(string(data), "// REPLACE_LABRAGO_DEVELOPMENT_API", "replace github.com/GoLabra/labra/src/api => ../api", 1)
	return os.WriteFile(path, []byte(out), 0644)
}

// createAppEnvFile writes a default backend .env configuration
// to the generated project so it can run out of the box.
func createAppEnvFile(projectName string) error {
	appPath := filepath.Join(projectName, "src", "app")
	schemaPath, _ := filepath.Abs(filepath.Join(appPath, "ent", "schema"))

	env := fmt.Sprintf(`# LabraGo Environment

SERVER_PORT=4001
SECRET_KEY=supersecretdevkey

DSN=postgres://postgres:postgres@localhost:5432/%s?sslmode=disable
DB_DIALECT=postgres

ENT_SCHEMA_PATH=%s

CENTRIFUGO_API_ADDRESS=http://localhost:8000
CENTRIFUGO_API_KEY=secretkey
`, projectName, schemaPath)

	return os.WriteFile(filepath.Join(appPath, ".env"), []byte(env), 0644)
}

// createAdminEnvFile writes the required environment variables for
// the frontend admin app.
func createAdminEnvFile(projectName string) error {
	content := `NEXT_PUBLIC_BRAND_PRODUCT_NAME="Labra·GO"
NEXT_PUBLIC_BRAND_COLOR="blue"
NEXT_PUBLIC_GRAPHQL_API_URL="http://localhost:4001"
NEXT_PUBLIC_GRAPHQL_QUERY_API_URL="http://localhost:4001/query"
NEXT_PUBLIC_GRAPHQL_QUERY_SUBSCRIPTION_URL="ws://localhost:4001/query"
NEXT_PUBLIC_GRAPHQL_QUERY_PLAYGROUND_URL="http://localhost:4001/playground"
NEXT_PUBLIC_GRAPHQL_ENTITY_API_URL="http://localhost:4001/entity"
NEXT_PUBLIC_GRAPHQL_ENTITY_PLAYGROUND_URL="http://localhost:4001/eplayground"`

	path := filepath.Join(projectName, "src", "admin", ".env")
	return os.WriteFile(path, []byte(content), 0644)
}

// ensurePostgresUserAndDatabase verifies that the postgres user and
// database exist, creating them if necessary.
func ensurePostgresUserAndDatabase(project string) error {
	log.Infof("🐘 Checking PostgreSQL...")

	if err := exec.Command("psql", "--version").Run(); err != nil {
		return fmt.Errorf("psql not found. Install it:\n→ macOS: brew install postgresql\n→ Ubuntu: sudo apt install postgresql\n→ Windows: https://postgresql.org/download")
	}

	// Attempt to create the user silently
	_ = exec.Command("createuser", "-s", "postgres").Run()

	checkCmd := exec.Command("psql", "-U", "postgres", "-tc", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s';", project))
	checkCmd.Env = append(os.Environ(), "PGPASSWORD=postgres")

	output, err := checkCmd.CombinedOutput()
	if err != nil {
		log.Errorf("❌ Failed to connect to PostgreSQL or run query.")
		log.Debugf("Output: %s", string(output))
		return fmt.Errorf("psql error: %w", err)
	}

	// Check if database already exists
	if strings.Contains(string(output), "1") {
		log.Infof("✅ PostgreSQL DB exists: %s", project)
		return nil
	}

	// Try to create database
	createCmd := exec.Command("createdb", "-U", "postgres", project)
	createCmd.Env = append(os.Environ(), "PGPASSWORD=postgres")
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("❌ failed to create database '%s': %w", project, err)
	}

	log.Infof("✅ PostgreSQL database created: %s", project)
	return nil
}

// checkPrerequisites ensures required tools like git and go are
// available and offers to install them when missing.
func checkPrerequisites() {
	requiredTools := []string{"git", "node", "psql"}

	// Check Go separately with better detection
	if !checkGoInstallation() {
		log.Warnf("⚠️  Go is not detected or not working properly. You may encounter issues if it's not available at runtime.")
		if !autoYes {
			fmt.Printf("%s Do you want to attempt installing Go now? (y/N): ", cliutils.Emoji("👉", "->"))
			answer := cliutils.ReadLine()
			if strings.ToLower(answer) == "y" {
				if err := installTool("go"); err != nil {
					log.Warnf("⚠️  Failed to install Go: %v", err)
				} else {
					log.Infof("✅ Go installed successfully")
				}
			}
		}
	}

	// Check other tools
	for _, tool := range requiredTools {
		if exec.Command(tool, "--version").Run() != nil {
			log.Warnf("⚠️  %s is not detected. You may encounter issues if it's not available at runtime.", tool)
			if !autoYes {
				fmt.Printf("%s Do you want to attempt installing it now? (y/N): ", cliutils.Emoji("👉", "->"))
				answer := cliutils.ReadLine()
				if strings.ToLower(answer) != "y" {
					continue
				}
			}
			if err := installTool(tool); err != nil {
				log.Warnf("⚠️  Failed to install %s: %v", tool, err)
			} else {
				log.Infof("✅ %s installed successfully", tool)
			}
		}
	}
}

// checkGoInstallation verifies that Go is properly installed and working.
// It checks if the go command exists, is executable, and can run go version.
func checkGoInstallation() bool {
	// First check if go command exists in PATH
	goPath, err := exec.LookPath("go")
	if err != nil {
		log.Debugf("Go not found in PATH: %v", err)
		return false
	}

	// Check if the file is executable
	if runtime.GOOS != "windows" {
		info, err := os.Stat(goPath)
		if err != nil {
			log.Debugf("Cannot stat go binary: %v", err)
			return false
		}
		if info.Mode()&0111 == 0 {
			log.Debugf("Go binary is not executable")
			return false
		}
	}

	// Test if go version works
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		log.Debugf("go version command failed: %v", err)
		return false
	}

	versionOutput := strings.TrimSpace(string(output))
	log.Debugf("Go version output: %s", versionOutput)

	// Verify it's actually a Go installation
	if !strings.Contains(versionOutput, "go version") {
		log.Debugf("go version output doesn't contain expected format")
		return false
	}

	log.Infof("✅ Go detected: %s", versionOutput)
	return true
}

// installTool attempts to install a missing tool using platform
// specific package managers.
func installTool(tool string) error {
	platform := runtime.GOOS
	log.Infof("⬇️ Installing %s on %s...", tool, platform)

	var cmd *exec.Cmd

	switch tool {
	case "git":
		cmd = getInstallCommand("git", platform)
	case "go":
		cmd = getInstallCommand("golang", platform)
	case "node":
		cmd = getInstallCommand("node", platform)
	case "psql":
		cmd = getInstallCommand("postgresql", platform)
	default:
		return fmt.Errorf("no install instructions for %s", tool)
	}

	if cmd == nil {
		return fmt.Errorf("automatic install not supported, please install %s manually", tool)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getInstallCommand returns the command used to install a given
// package for the current platform.
func getInstallCommand(pkg, platform string) *exec.Cmd {
	switch platform {
	case "darwin":
		return exec.Command("brew", "install", pkg)
	case "linux":
		return exec.Command("sudo", "apt", "install", "-y", pkg)
	case "windows":
		if exec.Command("choco", "--version").Run() == nil {
			return exec.Command("choco", "install", pkg, "-y")
		}
		return nil
	default:
		return nil
	}
}

// choosePackageManager prompts the user for their preferred
// package manager, defaulting to yarn.
func choosePackageManager() string {
	fmt.Printf("%s Choose package manager (npm/yarn) [default: yarn]: ", cliutils.Emoji("📦", "[pkg]"))
	choice := cliutils.ReadLine()
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "npm" {
		return "npm"
	}
	return "yarn"
}

// init registers the create command with the root command.
func init() {
	createCmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Automatic yes to prompts")
	rootCmd.AddCommand(createCmd)
}
