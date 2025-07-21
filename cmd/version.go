package cmd

import (
	"fmt"
	"github.com/GoLabra/labractl/internal/cliutils"
	"github.com/spf13/cobra"
)

// Version is the current version of labractl. It is overridden at build time using -ldflags.
var Version = "dev"

// version is set at build time using -ldflags.
var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print labractl version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(cliutils.Emoji("🧰", "version:"), version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
