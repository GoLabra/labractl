package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/GoLabra/labractl/internal/log"
)

// rootCmd represents the base command when called without any subcommands.
// It provides common configuration and flags for all subcommands.
var rootCmd = &cobra.Command{
	Use:   "labractl",
	Short: "CLI tool for LabraGo",
	Long:  `CLI tool to create and start a LabraGo project`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		envDebug := strings.ToLower(os.Getenv("LABRA_DEBUG"))
		enableDebug := debug || envDebug == "1" || envDebug == "true"
		log.Init(enableDebug)
	},
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var debug bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
	checkLatestVersion()
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.labractl.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "", false, "Enable debug logs")
}

// checkLatestVersion fetches the latest release tag from GitHub and compares
// it against the current Version. If a newer version is available, it informs
// the user on stderr. Network errors are silently ignored.
func checkLatestVersion() {
	if Version == "dev" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/repos/GoLabra/labractl/releases/latest", nil)
	if err != nil {
		return
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return
	}

	var r struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return
	}
	latest := strings.TrimSpace(r.TagName)
	if latest == "" {
		return
	}

	current := Version
	if !strings.HasPrefix(current, "v") {
		current = "v" + current
	}
	if !strings.HasPrefix(latest, "v") {
		latest = "v" + latest
	}

	if semver.Compare(current, latest) < 0 {
		fmt.Fprintf(os.Stderr, "A new version of labractl is available: %s\n", latest)
	}
}
