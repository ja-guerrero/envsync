package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/ja-guerrero/envsync/internal/config"
	"github.com/ja-guerrero/envsync/internal/envfile"
	"github.com/ja-guerrero/envsync/internal/schema"
	"github.com/spf13/cobra"
)

var envFilePath string

var (
	bold    = color.New(color.Bold)
	red     = color.New(color.FgRed)
	green   = color.New(color.FgGreen)
	faint   = color.New(color.Faint)
	varName = color.New(color.FgCyan, color.Bold)
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a .env file against the schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadRepoConfig(schemaPath)
		if err != nil {
			if jsonOutput {
				outputJSON(CLIResult{
					Status:    "error",
					ErrorCode: "E_CONFIG_INVALID",
					Message:   err.Error(),
				})
			}
			return err
		}

		f, err := os.Open(envFilePath)
		if err != nil {
			return exitWithError(fmt.Sprintf("opening env file: %v", err), "E_CONFIG_MISSING")
		}
		defer f.Close()

		env, err := envfile.Parse(f)
		if err != nil {
			return exitWithError(fmt.Sprintf("parsing env file: %v", err), "E_PARSE_INVALID_KEY")
		}

		violations := schema.Validate(cfg.Vars, env)

		if jsonOutput {
			vj := make([]ViolationJSON, len(violations))
			for i, v := range violations {
				vj[i] = ViolationJSON{Var: v.Var, Message: v.Message}
			}
			status := "ok"
			if len(vj) > 0 {
				status = "error"
			}
			return outputJSON(CLIResult{
				Status:     status,
				Violations: vj,
				Message:    fmt.Sprintf("%d violation(s)", len(violations)),
			})
		}

		if len(violations) == 0 {
			green.Printf("✓ ")
			bold.Printf("%s", envFilePath)
			fmt.Printf(" passed all checks against ")
			bold.Printf("%s\n", schemaPath)
			return nil
		}

		red.Printf("✗ ")
		bold.Printf("%s", envFilePath)
		fmt.Printf(" has %d violation(s) against ", len(violations))
		bold.Printf("%s\n\n", schemaPath)

		for _, v := range violations {
			fmt.Printf("  ")
			varName.Printf("%-30s", v.Var)
			faint.Printf("  %s\n", v.Message)
		}

		fmt.Println()
		return fmt.Errorf("%d violation(s) found", len(violations))
	},
}

func init() {
	validateCmd.Flags().StringVar(&schemaPath, "schema", ".envsync.yaml", "path to repo env schema")
	validateCmd.Flags().StringVar(&envFilePath, "env-file", ".env", "path to .env file")
	rootCmd.AddCommand(validateCmd)
}
