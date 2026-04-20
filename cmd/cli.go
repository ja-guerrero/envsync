package cmd

import (
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	verbose    bool
	noColor    bool
)

var rootCmd = &cobra.Command{
	Use:           "envsync",
	Short:         "Validate and sync env vars against a schema",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "show detailed output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable color output")
}

func Execute() error {
	return rootCmd.Execute()
}
