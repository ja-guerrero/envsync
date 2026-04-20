package cmd

import (
	"fmt"

	"github.com/ja-guerrero/envsync/internal/config"
	"github.com/spf13/cobra"
)

var schemaPath string

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Lint the repo env schema file",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := config.LoadRepoConfig(schemaPath)
		if err != nil {
			if jsonOutput {
				outputJSON(CLIResult{
					Status:    "error",
					ErrorCode: "E_SCHEMA_INVALID_FIELD",
					Message:   err.Error(),
				})
			}
			return err
		}

		if jsonOutput {
			return outputJSON(CLIResult{
				Status:  "ok",
				Message: fmt.Sprintf("%s is valid", schemaPath),
			})
		}

		green.Printf("✓ ")
		bold.Printf("%s", schemaPath)
		fmt.Println(" is valid")
		return nil
	},
}

func init() {
	lintCmd.Flags().StringVar(&schemaPath, "schema", ".envsync.yaml", "path to repo env schema")
	rootCmd.AddCommand(lintCmd)
}
