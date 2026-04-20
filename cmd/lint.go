package cmd

import (
	"github.com/ja-guerrero/envsync/internal/config"
	"github.com/spf13/cobra"
)

var schemaPath string

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint the repo env schema file",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.LoadRepoConfig(schemaPath)
		return err
	},
}

func init() {
	lintCmd.Flags().StringVar(&schemaPath, "schema", ".envsync.yaml", "path to repo env schema")
	rootCmd.AddCommand(lintCmd)
}
