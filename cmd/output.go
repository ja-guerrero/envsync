package cmd

import (
	"encoding/json"
	"fmt"
	"os"
)

// CLIResult is the JSON envelope returned when --json is set.
type CLIResult struct {
	Status     string          `json:"status"`
	ErrorCode  string          `json:"error_code,omitempty"`
	Message    string          `json:"message,omitempty"`
	Violations []ViolationJSON `json:"violations,omitempty"`
	Data       interface{}     `json:"data,omitempty"`
}

// ViolationJSON is a single schema violation in JSON output.
type ViolationJSON struct {
	Var     string `json:"var"`
	Message string `json:"message"`
}

// outputJSON writes result to stdout as indented JSON and returns nil.
func outputJSON(result CLIResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

// exitWithError returns an error that also emits a JSON result when --json is
// set, so callers can do: return exitWithError(msg, code)
func exitWithError(msg string, code string) error {
	if jsonOutput {
		_ = outputJSON(CLIResult{
			Status:    "error",
			ErrorCode: code,
			Message:   msg,
		})
	}
	return fmt.Errorf("%s", msg)
}
