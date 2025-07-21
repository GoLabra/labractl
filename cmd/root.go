package cmd

import (
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
