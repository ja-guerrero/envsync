package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	"github.com/ja-guerrero/envsync/internal/envfile"
	"github.com/spf13/cobra"
)

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show what sync would change",
	RunE:  runDiff,
}

func init() {
	diffCmd.Flags().StringVar(&envFilePath, "env-file", ".env", "path to current .env file")
	rootCmd.AddCommand(diffCmd)
}

type DiffEntry struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	OldValue string `json:"old_value,omitempty"`
	NewValue string `json:"new_value,omitempty"`
}

func runDiff(cmd *cobra.Command, args []string) error {
	current := make(map[string]string)
	if f, err := os.Open(envFilePath); err == nil {
		defer f.Close()
		parsed, err := envfile.Parse(f)
		if err != nil {
			return fmt.Errorf("parsing current env file: %w", err)
		}
		current = parsed
	}

	resolved := current

	entries := computeDiff(current, resolved)

	if jsonOutput {
		return outputJSON(CLIResult{
			Status:  "ok",
			Message: fmt.Sprintf("%d changes", countChanges(entries)),
			Data:    entries,
		})
	}

	return printDiff(entries)
}

func computeDiff(current, resolved map[string]string) []DiffEntry {
	allKeys := make(map[string]bool)
	for k := range current {
		allKeys[k] = true
	}
	for k := range resolved {
		allKeys[k] = true
	}

	sorted := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var entries []DiffEntry
	for _, k := range sorted {
		oldVal, inOld := current[k]
		newVal, inNew := resolved[k]

		switch {
		case !inOld && inNew:
			entries = append(entries, DiffEntry{Key: k, Type: "added", NewValue: newVal})
		case inOld && !inNew:
			entries = append(entries, DiffEntry{Key: k, Type: "removed", OldValue: oldVal})
		case oldVal != newVal:
			entries = append(entries, DiffEntry{Key: k, Type: "changed", OldValue: oldVal, NewValue: newVal})
		default:
			entries = append(entries, DiffEntry{Key: k, Type: "unchanged"})
		}
	}

	return entries
}

func printDiff(entries []DiffEntry) error {
	added := color.New(color.FgGreen)
	removed := color.New(color.FgRed)
	changed := color.New(color.FgYellow)
	unchanged := color.New(color.Faint)

	for _, e := range entries {
		switch e.Type {
		case "added":
			added.Printf("  + %-30s", e.Key)
			fmt.Printf("  %s\n", e.NewValue)
		case "removed":
			removed.Printf("  - %-30s", e.Key)
			fmt.Printf("  %s\n", e.OldValue)
		case "changed":
			changed.Printf("  ~ %-30s", e.Key)
			fmt.Printf("  %s → %s\n", e.OldValue, e.NewValue)
		case "unchanged":
			if verbose {
				unchanged.Printf("    %-30s", e.Key)
				fmt.Println("  (unchanged)")
			}
		}
	}

	changes := countChanges(entries)
	if changes == 0 {
		green.Println("  no changes")
	}

	return nil
}

func countChanges(entries []DiffEntry) int {
	n := 0
	for _, e := range entries {
		if e.Type != "unchanged" {
			n++
		}
	}
	return n
}
