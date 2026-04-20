package backend

import (
	"context"
	"fmt"
)

// VarSource describes where a single variable's value comes from.
type VarSource struct {
	BackendName string
	Path        string
	Key         string
}

// SyncVars fetches all variables from their respective backends.
// It groups refs by backend for efficient batch reads.
func SyncVars(ctx context.Context, vars map[string]*VarSource, backends map[string]Backend) (map[string]string, error) {
	type varRef struct {
		varName string
		ref     Ref
	}
	groups := make(map[string][]varRef)
	for varName, src := range vars {
		groups[src.BackendName] = append(groups[src.BackendName], varRef{
			varName: varName,
			ref:     Ref{Path: src.Path, Key: src.Key},
		})
	}

	result := make(map[string]string, len(vars))

	for backendName, vrefs := range groups {
		b, ok := backends[backendName]
		if !ok {
			return nil, &CodedError{
				Code:     ECodeBackendNotFound,
				Category: CategoryUser,
				Message:  fmt.Sprintf("backend %q not configured", backendName),
			}
		}

		refs := make([]Ref, len(vrefs))
		for i, vr := range vrefs {
			refs[i] = vr.ref
		}

		fetched, err := b.GetBatch(ctx, refs)
		if err != nil {
			return nil, fmt.Errorf("backend %q: %w", backendName, err)
		}

		for _, vr := range vrefs {
			val, ok := fetched[vr.ref.Key]
			if !ok {
				continue // key not in backend — skip, let the validator catch required vars
			}
			result[vr.varName] = val
		}
	}

	return result, nil
}
