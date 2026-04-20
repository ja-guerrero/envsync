package cmd

// Shared flag variables used across multiple commands.
// Each command registers these on its own flag set in init().
var (
	schemaPath  string // sync, lint, validate
	envFilePath string // validate, diff
)
