package backend

import (
	"context"
	"fmt"
)

// Ref identifies a single secret in a backend.
type Ref struct {
	Path string // secret path (e.g. "myapp/db")
	Key  string // key within the secret (e.g. "connection_string")
}

func (r Ref) String() string {
	return fmt.Sprintf("%s/%s", r.Path, r.Key)
}

// Backend is the interface for secret storage providers.
type Backend interface {
	// Get retrieves a single secret value.
	Get(ctx context.Context, ref Ref) (string, error)

	// GetBatch retrieves multiple secrets efficiently.
	// Implementations should group by path for batch reads where possible.
	GetBatch(ctx context.Context, refs []Ref) (map[string]string, error)
}

// Registry maps backend type names to constructor functions.
var Registry = map[string]func(params map[string]interface{}) (Backend, error){}

// NewBackend creates a backend by type name using the registry.
func NewBackend(backendType string, params map[string]interface{}) (Backend, error) {
	constructor, ok := Registry[backendType]
	if !ok {
		return nil, &CodedError{
			Code:     ECodeBackendNotFound,
			Category: CategoryUser,
			Message:  fmt.Sprintf("unknown backend type %q", backendType),
		}
	}
	return constructor(params)
}
