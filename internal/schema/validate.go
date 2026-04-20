package schema

import (
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strconv"
)

type Violation struct {
	Var     string
	Message string
}

func Validate(s Schema, env map[string]string) []Violation {
	var violations []Violation

	for name, v := range s {
		val, present := env[name]

		if !present {
			if v.Required {
				violations = append(violations, Violation{
					Var:     name,
					Message: "required but missing",
				})
			}
			continue
		}

		if v.Type != "" {
			if err := checkType(v.Type, val); err != nil {
				violations = append(violations, Violation{Var: name, Message: err.Error()})
				continue // skip further checks if the type is wrong
			}
		}

		if len(v.Enum) > 0 && !slices.Contains(v.Enum, val) {
			violations = append(violations, Violation{
				Var:     name,
				Message: fmt.Sprintf("value %q not in enum %v", val, v.Enum),
			})
		}

		if v.Format != "" {
			re, err := regexp.Compile(v.Format)
			if err != nil {
				violations = append(violations, Violation{
					Var:     name,
					Message: fmt.Sprintf("schema format %q is not a valid regex: %v", v.Format, err),
				})
			} else if !re.MatchString(val) {
				violations = append(violations, Violation{
					Var:     name,
					Message: fmt.Sprintf("value %q does not match format %q", val, v.Format),
				})
			}
		}

		if v.Type == "number" {
			n, _ := strconv.ParseFloat(val, 64) // safe: already passed checkType
			if v.Min != nil && n < *v.Min {
				violations = append(violations, Violation{
					Var:     name,
					Message: fmt.Sprintf("value %g is below minimum %g", n, *v.Min),
				})
			}
			if v.Max != nil && n > *v.Max {
				violations = append(violations, Violation{
					Var:     name,
					Message: fmt.Sprintf("value %g exceeds maximum %g", n, *v.Max),
				})
			}
		}
	}

	return violations
}

func checkType(t, val string) error {
	switch t {
	case "string":
		return nil
	case "number":
		if _, err := strconv.ParseFloat(val, 64); err != nil {
			return fmt.Errorf("expected number, got %q", val)
		}
	case "bool":
		if val != "true" && val != "false" {
			return fmt.Errorf("expected bool (true/false), got %q", val)
		}
	case "url":
		u, err := url.Parse(val)
		if err != nil || u.Scheme == "" || u.Host == "" {
			return fmt.Errorf("expected valid URL, got %q", val)
		}
	case "json":
		if !json.Valid([]byte(val)) {
			return fmt.Errorf("expected valid JSON, got %q", val)
		}
	}
	return nil
}

