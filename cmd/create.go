package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var autoYes bool

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new LabraGo project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		repoURL := "https://github.com/GoLabra/labra"
		fmt.Println("🚀 Creating LabraGo project:", projectName)

		// 1. Check prerequisites
		checkPrerequisites()

		// 2. Choose package manager
		packageManager := choosePackageManager()

		// 3. Clone repo
		if err := runCommand("git", []string{"clone", repoURL, projectName}, ""); err != nil {
			fmt.Println("❌ Git clone failed:", err)
			os.Exit(1)
		}

		// 4. Patch go.mod
		goModPath := filepath.Join(projectName, "src", "app", "go.mod")
		if err := patchGoMod(goModPath); err != nil {
			fmt.Println("❌ go.mod patch failed:", err)
			os.Exit(1)
		}

		// 5. Create .env files
		if err := createAppEnvFile(projectName); err != nil {
			fmt.Println("❌ Backend .env failed:", err)
			os.Exit(1)
		}
		if err := createAdminEnvFile(projectName); err != nil {
			fmt.Println("❌ Frontend .env failed:", err)
			os.Exit(1)
		}

		// 6. Go mod tidy + generate
		appPath := filepath.Join(projectName, "src", "app")
		_ = runCommand("go", []string{"mod", "tidy"}, appPath)
		if err := runCommand("go", []string{"generate", "./..."}, appPath); err != nil {
			fmt.Println("⚠️ go generate failed, retrying...")
			_ = runCommand("go", []string{"mod", "tidy"}, appPath)
			_ = runCommand("go", []string{"generate", "./..."}, appPath)
		}

		// 7. Frontend install
		adminPath := filepath.Join(projectName, "src", "admin")
		if _, err := os.Stat(filepath.Join(adminPath, "package.json")); err == nil {
			fmt.Printf("📦 Installing frontend dependencies with %s...\n", packageManager)
			_ = runCommand(packageManager, []string{"install"}, adminPath)
		}

		// 8. Ensure PostgreSQL
		if err := ensurePostgresUserAndDatabase(projectName); err != nil {
			fmt.Println("⚠️ PostgreSQL setup failed:", err)
		}

		// 9. Done
		fmt.Println("✅ Project created at", projectName)
		fmt.Printf("👉 cd %s\nlabractl start\n", projectName)
	},
}

func runCommand(bin string, args []string, dir string) error {
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func patchGoMod(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	out := strings.Replace(string(data), "// REPLACE_LABRAGO_DEVELOPMENT_API", "replace github.com/GoLabra/labra/src/api => ../api", 1)
	return os.WriteFile(path, []byte(out), 0644)
}

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

func ensurePostgresUserAndDatabase(project string) error {
	fmt.Println("🐘 Checking PostgreSQL...")

	if err := exec.Command("psql", "--version").Run(); err != nil {
		return fmt.Errorf("psql not found. Install it:\n→ macOS: brew install postgresql\n→ Ubuntu: sudo apt install postgresql\n→ Windows: https://postgresql.org/download")
	}

	// Attempt to create the user silently
	_ = exec.Command("createuser", "-s", "postgres").Run()

	checkCmd := exec.Command("psql", "-U", "postgres", "-tc", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s';", project))
	checkCmd.Env = append(os.Environ(), "PGPASSWORD=postgres")

	output, err := checkCmd.CombinedOutput()
	if err != nil {
		fmt.Println("❌ Failed to connect to PostgreSQL or run query.")
		fmt.Println("Output:", string(output))
		return fmt.Errorf("psql error: %w", err)
	}

	// Check if database already exists
	if strings.Contains(string(output), "1") {
		fmt.Println("✅ PostgreSQL DB exists:", project)
		return nil
	}

	// Try to create database
	createCmd := exec.Command("createdb", "-U", "postgres", project)
	createCmd.Env = append(os.Environ(), "PGPASSWORD=postgres")
	if err := createCmd.Run(); err != nil {
		return fmt.Errorf("❌ failed to create database '%s': %w", project, err)
	}

	fmt.Println("✅ PostgreSQL database created:", project)
	return nil
}

func checkPrerequisites() {
	requiredTools := []string{"git", "go", "node", "psql"}

	for _, tool := range requiredTools {
		if exec.Command(tool, "--version").Run() != nil {
			fmt.Printf("⚠️  %s is not detected. You may encounter issues if it's not available at runtime.\n", tool)
			if !autoYes {
				fmt.Print("👉 Do you want to attempt installing it now? (y/N): ")
				answer := readLine()
				if strings.ToLower(answer) != "y" {
					continue
				}
			}
			if err := installTool(tool); err != nil {
				fmt.Printf("⚠️  Failed to install %s: %v\n", tool, err)
			} else {
				fmt.Printf("✅ %s installed successfully\n", tool)
			}
		}
	}
}

func installTool(tool string) error {
	platform := runtime.GOOS
	fmt.Printf("⬇️ Installing %s on %s...\n", tool, platform)

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

func choosePackageManager() string {
	fmt.Print("📦 Choose package manager (npm/yarn) [default: yarn]: ")
	choice := readLine()
	choice = strings.ToLower(strings.TrimSpace(choice))

	if choice == "npm" {
		return "npm"
	}
	return "yarn"
}

func readLine() string {
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimSpace(line)
}

func init() {
	createCmd.Flags().BoolVarP(&autoYes, "yes", "y", false, "Automatic yes to prompts")
	rootCmd.AddCommand(createCmd)
}
