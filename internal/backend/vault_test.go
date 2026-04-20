package backend

import (
	"context"
	"testing"
)

// mockBackend implements Backend for testing.
type mockBackend struct {
	secrets map[string]map[string]string // path -> key -> value
}

func (m *mockBackend) Get(ctx context.Context, ref Ref) (string, error) {
	path, ok := m.secrets[ref.Path]
	if !ok {
		return "", &ErrKeyNotFound{Key: ref.String(), Backend: "mock"}
	}
	val, ok := path[ref.Key]
	if !ok {
		return "", &ErrKeyNotFound{Key: ref.String(), Backend: "mock"}
	}
	return val, nil
}

func (m *mockBackend) GetBatch(ctx context.Context, refs []Ref) (map[string]string, error) {
	result := make(map[string]string, len(refs))
	for _, ref := range refs {
		val, err := m.Get(ctx, ref)
		if err != nil {
			return nil, err
		}
		result[ref.Key] = val
	}
	return result, nil
}

func TestMockBackendGet(t *testing.T) {
	b := &mockBackend{
		secrets: map[string]map[string]string{
			"myapp/db": {
				"password": "secret123",
				"host":     "localhost",
			},
		},
	}

	ctx := context.Background()

	val, err := b.Get(ctx, Ref{Path: "myapp/db", Key: "password"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "secret123" {
		t.Errorf("got %q, want %q", val, "secret123")
	}

	_, err = b.Get(ctx, Ref{Path: "myapp/db", Key: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestMockBackendGetBatch(t *testing.T) {
	b := &mockBackend{
		secrets: map[string]map[string]string{
			"myapp/db": {
				"password": "secret123",
				"host":     "localhost",
			},
		},
	}

	ctx := context.Background()
	refs := []Ref{
		{Path: "myapp/db", Key: "password"},
		{Path: "myapp/db", Key: "host"},
	}

	result, err := b.GetBatch(ctx, refs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result["password"] != "secret123" {
		t.Errorf("password: got %q, want %q", result["password"], "secret123")
	}
}

func TestNewBackendUnknownType(t *testing.T) {
	_, err := NewBackend("nonexistent", nil)
	if err == nil {
		t.Fatal("expected error for unknown backend type")
	}
}
