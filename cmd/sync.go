package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/ja-guerrero/envsync/internal/backend"
	"github.com/ja-guerrero/envsync/internal/config"
	"github.com/ja-guerrero/envsync/internal/envfile"
	"github.com/ja-guerrero/envsync/internal/schema"
	"github.com/spf13/cobra"
)

var (
	envName        string
	userConfigPath string
	outputPath     string
	secretsOnly    bool
	dryRun         bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync env vars from backends",
	RunE:  runSync,
}

func init() {
	home, _ := os.UserHomeDir()
	defaultConfig := filepath.Join(home, ".envsync", "config.yaml")

	syncCmd.Flags().StringVar(&envName, "env", "", "environment to sync (required)")
	syncCmd.Flags().StringVar(&schemaPath, "schema", ".envsync.yaml", "path to repo env schema")
	syncCmd.Flags().StringVar(&userConfigPath, "config", defaultConfig, "path to user config")
	syncCmd.Flags().StringVar(&outputPath, "output", ".env", "output .env file path")
	syncCmd.Flags().BoolVar(&secretsOnly, "secrets-only", false, "only fetch vars marked secret")
	syncCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show changes without writing")
	syncCmd.MarkFlagRequired("env")
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	repoCfg, err := config.LoadRepoConfig(schemaPath)
	if err != nil {
		return err
	}

	envRef, ok := repoCfg.Environments[envName]
	if !ok {
		return fmt.Errorf("environment %q not found in schema", envName)
	}

	backends, err := resolveBackends(repoCfg.Vars, envRef)
	if err != nil {
		return err
	}

	varSources := buildVarSources(repoCfg.Vars, envRef, secretsOnly)
	if len(varSources) == 0 {
		fmt.Println("no variables to sync")
		return nil
	}

	fetched, err := backend.SyncVars(ctx, varSources, backends)
	if err != nil {
		return err
	}

	merged := applyDefaults(repoCfg.Vars, fetched, varSources)

	violations := schema.Validate(repoCfg.Vars, merged)
	if len(violations) > 0 {
		red.Printf("✗ ")
		bold.Printf("synced values have %d violation(s)\n\n", len(violations))
		for _, v := range violations {
			fmt.Printf("  ")
			varName.Printf("%-30s", v.Var)
			faint.Printf("  %s\n", v.Message)
		}
		fmt.Println()
	}

	if dryRun {
		yellow := color.New(color.FgYellow)
		yellow.Printf("~ ")
		fmt.Printf("dry run: would write %d vars to %s\n", len(merged), outputPath)
		return nil
	}

	if err := envfile.Write(outputPath, merged); err != nil {
		return fmt.Errorf("writing env file: %w", err)
	}

	green.Printf("✓ ")
	fmt.Printf("synced %d vars to %s\n", len(merged), outputPath)
	return nil
}

func resolveBackends(vars schema.Schema, envRef config.EnvironmentRef) (map[string]backend.Backend, error) {
	needed := make(map[string]bool)

	for _, v := range vars {
		if v.Source != nil {
			needed[v.Source.Backend] = true
		}
	}

	if envRef.Backend != "" {
		needed[envRef.Backend] = true
	}
	if envRef.Type != "" {
		needed[envRef.Type] = true
	}

	if len(needed) == 0 {
		return nil, nil
	}

	userCfg, err := config.LoadUserConfig(userConfigPath)
	if err != nil {
		return nil, fmt.Errorf("loading user config: %w", err)
	}

	backends := make(map[string]backend.Backend, len(needed))
	for name := range needed {
		var backendType string
		var params map[string]interface{}

		for _, b := range userCfg.Backends {
			if b.Name == name {
				backendType = b.Type
				params = b.Params
				break
			}
		}

		if backendType == "" {
			backendType = name
			params = envRef.Params
		}

		b, err := backend.Resolve(backendType, params)
		if err != nil {
			return nil, fmt.Errorf("resolving backend %q: %w", name, err)
		}
		backends[name] = b
	}

	return backends, nil
}

func buildVarSources(vars schema.Schema, envRef config.EnvironmentRef, secretsOnly bool) map[string]*backend.VarSource {
	sources := make(map[string]*backend.VarSource)

	envPath := ""
	if p, ok := envRef.Params["path"]; ok {
		if s, ok := p.(string); ok {
			envPath = s
		}
	}

	envBackend := envRef.Backend
	if envBackend == "" {
		envBackend = envRef.Type
	}

	for name, v := range vars {
		if secretsOnly && !v.Secret {
			continue
		}

		if v.Source != nil {
			key := v.Source.Key
			if key == "" {
				key = name
			}
			sources[name] = &backend.VarSource{
				BackendName: v.Source.Backend,
				Path:        v.Source.Path,
				Key:         key,
			}
		} else if envBackend != "" && envPath != "" {
			sources[name] = &backend.VarSource{
				BackendName: envBackend,
				Path:        envPath,
				Key:         name,
			}
		}
	}

	return sources
}

func applyDefaults(vars schema.Schema, fetched map[string]string, sources map[string]*backend.VarSource) map[string]string {
	merged := make(map[string]string, len(sources))

	for k, v := range fetched {
		merged[k] = v
	}

	for name := range sources {
		if _, ok := merged[name]; ok {
			continue
		}
		if v, ok := vars[name]; ok && v.Default != nil {
			merged[name] = *v.Default
		}
	}

	return merged
}
