package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create [project-name]",
	Short: "Create a new LabraGo project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		repoURL := "https://github.com/GoLabra/labra"
		fmt.Println("🚀 Creating LabraGo project:", projectName)

		// 1. Clone repo
		if err := runCommand("git", []string{"clone", repoURL, projectName}, ""); err != nil {
			fmt.Println("❌ Git clone failed:", err)
			os.Exit(1)
		}

		// 2. Patch go.mod
		goModPath := filepath.Join(projectName, "src", "app", "go.mod")
		if err := patchGoMod(goModPath); err != nil {
			fmt.Println("❌ go.mod patch failed:", err)
			os.Exit(1)
		}

		// 3. Create .env files
		if err := createAppEnvFile(projectName); err != nil {
			fmt.Println("❌ Backend .env failed:", err)
			os.Exit(1)
		}
		if err := createAdminEnvFile(projectName); err != nil {
			fmt.Println("❌ Frontend .env failed:", err)
			os.Exit(1)
		}

		// 4. go mod tidy + go generate (with fallback)
		appPath := filepath.Join(projectName, "src", "app")
		_ = runCommand("go", []string{"mod", "tidy"}, appPath)
		if err := runCommand("go", []string{"generate", "./..."}, appPath); err != nil {
			fmt.Println("⚠️ go generate failed, retrying...")
			_ = runCommand("go", []string{"mod", "tidy"}, appPath)
			_ = runCommand("go", []string{"generate", "./..."}, appPath)
		}

		// 5. Yarn install (frontend)
		adminPath := filepath.Join(projectName, "src", "admin")
		if _, err := os.Stat(filepath.Join(adminPath, "package.json")); err == nil {
			fmt.Println("📦 Installing frontend dependencies with Yarn...")
			_ = runCommand("yarn", []string{"install"}, adminPath)
		}

		// 6. Ensure Postgres
		if err := ensurePostgresUserAndDatabase(projectName); err != nil {
			fmt.Println("⚠️ PostgreSQL setup failed:", err)
		}

		// 7. Done
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
NEXT_PUBLIC_GRAPHQL_ENTITY_PLAYGROUND_URL="http://localhost:4001/eplayground"
`
	path := filepath.Join(projectName, "src", "admin", ".env")
	return os.WriteFile(path, []byte(content), 0644)
}

func ensurePostgresUserAndDatabase(project string) error {
	fmt.Println("🐘 Checking PostgreSQL...")

	// Check if psql exists
	if err := exec.Command("psql", "--version").Run(); err != nil {
		return fmt.Errorf("psql not found. Install it:\n→ macOS: brew install postgresql\n→ Ubuntu: sudo apt install postgresql\n→ Windows: https://postgresql.org/download")
	}

	// Try to create user
	_ = exec.Command("createuser", "-s", "postgres").Run()

	// Check DB
	check := exec.Command("psql", "-U", "postgres", "-tc", fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", project))
	check.Env = append(os.Environ(), "PGPASSWORD=postgres")
	out, _ := check.Output()
	if strings.TrimSpace(string(out)) == "1" {
		fmt.Println("✅ PostgreSQL DB exists:", project)
		return nil
	}

	// Create DB
	create := exec.Command("createdb", "-U", "postgres", project)
	create.Env = append(os.Environ(), "PGPASSWORD=postgres")
	if err := create.Run(); err != nil {
		return fmt.Errorf("failed to create database '%s': %w", project, err)
	}
	fmt.Println("✅ PostgreSQL database created:", project)
	return nil
}

func init() {
	rootCmd.AddCommand(createCmd)
}
