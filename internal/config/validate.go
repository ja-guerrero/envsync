package config

import (
	"errors"
	"fmt"
)

var validTypes = map[string]bool{
	"string": true,
	"number": true,
	"bool":   true,
	"url":    true,
	"json":   true,
}

func validateRepoConfig(cfg *RepoConfig) error {
	var errs []error

	for name, v := range cfg.Vars {
		if !validTypes[v.Type] {
			errs = append(errs, fmt.Errorf("var %q: unknown type %q", name, v.Type))
		}

		if v.Type != "number" {
			if v.Min != nil {
				errs = append(errs, fmt.Errorf("var %q: min requires type number", name))
			}
			if v.Max != nil {
				errs = append(errs, fmt.Errorf("var %q: max requires type number", name))
			}
		}

		if v.Min != nil && v.Max != nil && *v.Min > *v.Max {
			errs = append(errs, fmt.Errorf("var %q: min %.2f > max %.2f", name, *v.Min, *v.Max))
		}

		if len(v.Enum) > 0 && v.Type != "string" {
			errs = append(errs, fmt.Errorf("var %q: enum requires type string", name))
		}

		if v.Required && v.Default != nil {
			errs = append(errs, fmt.Errorf("var %q: required and default are mutually exclusive", name))
		}

		if v.Required && len(v.RequiredIn) > 0 {
			errs = append(errs, fmt.Errorf("var %q: required and required_in are mutually exclusive", name))
		}

		if v.Required && len(v.RequiredIf) > 0 {
			errs = append(errs, fmt.Errorf("var %q: required and required_if are mutually exclusive", name))
		}
	}

	return errors.Join(errs...)
}

func ValidateUserConfig(_ *UserConfig) error {
	return nil
}
