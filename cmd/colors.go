package cmd

import "github.com/fatih/color"

// Shared color styles used across multiple commands.
var (
	bold    = color.New(color.Bold)
	red     = color.New(color.FgRed)
	green   = color.New(color.FgGreen)
	yellow  = color.New(color.FgYellow)
	faint   = color.New(color.Faint)
	varName = color.New(color.FgCyan, color.Bold)
)
