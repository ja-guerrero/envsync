package backend

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hashicorp/vault-client-go"
)

// Resolve builds a Backend from a backend type and params map.
// For vault backends, it enriches params with VAULT_ADDR / VAULT_TOKEN
// env var fallbacks before delegating to the Registry.
func Resolve(backendType string, params map[string]interface{}) (Backend, error) {
	if backendType == "vault" {
		params = enrichVaultParams(params)
	}
	return NewBackend(backendType, params)
}

func enrichVaultParams(params map[string]interface{}) map[string]interface{} {
	enriched := make(map[string]interface{}, len(params))
	for k, v := range params {
		enriched[k] = v
	}
	if _, ok := enriched["addr"]; !ok {
		if addr := os.Getenv("VAULT_ADDR"); addr != "" {
			enriched["addr"] = addr
		}
	}
	if _, ok := enriched["token"]; !ok {
		if token := os.Getenv("VAULT_TOKEN"); token != "" {
			enriched["token"] = token
		} else if token := readVaultTokenFile(); token != "" {
			enriched["token"] = token
		}
	}
	return enriched
}

func stringParam(params map[string]interface{}, key, fallback string) string {
	if v, ok := params[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return fallback
}

func intParam(params map[string]interface{}, key string, fallback int) int {
	if v, ok := params[key]; ok {
		switch n := v.(type) {
		case int:
			return n
		case float64:
			return int(n)
		}
	}
	return fallback
}

func readVaultTokenFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".vault-token"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func init() {
	Registry["vault"] = NewVaultBackend
}

type KVVersion int

const (
	KVv1 KVVersion = 1
	KVv2 KVVersion = 2
)

// VaultBackend implements Backend for HashiCorp Vault KV stores.
type VaultBackend struct {
	client    *vault.Client
	mountPath string
	version   KVVersion
}

// NewVaultBackend creates a VaultBackend from params.
func NewVaultBackend(params map[string]interface{}) (Backend, error) {
	addr := stringParam(params, "addr", "")
	token := stringParam(params, "token", "")
	mount := stringParam(params, "mount", "secret")
	kvVersion := KVVersion(intParam(params, "kv_version", 2))

	if addr == "" {
		return nil, &CodedError{
			Code:     ECodeConfigMissing,
			Category: CategoryUser,
			Message:  "vault: addr required (set in config or VAULT_ADDR env var)",
		}
	}
	if token == "" {
		return nil, &CodedError{
			Code:     ECodeConfigMissing,
			Category: CategoryUser,
			Message:  "vault: token required (set in config, VAULT_TOKEN env var, or vault login)",
		}
	}
	if kvVersion != KVv1 && kvVersion != KVv2 {
		return nil, &CodedError{
			Code:     ECodeConfigInvalid,
			Category: CategoryUser,
			Message:  fmt.Sprintf("vault: unsupported KV version %d (must be 1 or 2)", kvVersion),
		}
	}

	client, err := vault.New(
		vault.WithAddress(addr),
		vault.WithRequestTimeout(30*time.Second),
	)
	if err != nil {
		return nil, &CodedError{
			Code:     ECodeBackendAuth,
			Category: CategoryBackend,
			Message:  fmt.Sprintf("creating vault client at %s", addr),
			Cause:    err,
		}
	}

	if err := client.SetToken(token); err != nil {
		return nil, &CodedError{
			Code:     ECodeBackendAuth,
			Category: CategoryBackend,
			Message:  "invalid vault token",
		}
	}

	return &VaultBackend{
		client:    client,
		mountPath: mount,
		version:   kvVersion,
	}, nil
}

func (v *VaultBackend) Get(ctx context.Context, ref Ref) (string, error) {
	secrets, err := v.readPath(ctx, ref.Path)
	if err != nil {
		return "", err
	}
	val, ok := secrets[ref.Key]
	if !ok {
		return "", &ErrKeyNotFound{Key: ref.Key, Backend: "vault"}
	}
	return val, nil
}

func (v *VaultBackend) GetBatch(ctx context.Context, refs []Ref) (map[string]string, error) {
	// Group refs by path to minimize Vault reads
	byPath := make(map[string][]Ref)
	for _, ref := range refs {
		byPath[ref.Path] = append(byPath[ref.Path], ref)
	}

	result := make(map[string]string, len(refs))
	for path, pathRefs := range byPath {
		secrets, err := v.readPath(ctx, path)
		if err != nil {
			return nil, err
		}
		for _, ref := range pathRefs {
			val, ok := secrets[ref.Key]
			if !ok {
				continue // key not at this path — caller decides if it's required
			}
			result[ref.Key] = val
		}
	}

	return result, nil
}

func (v *VaultBackend) readPath(ctx context.Context, secretPath string) (map[string]string, error) {
	raw, err := v.readRaw(ctx, secretPath)
	if err != nil {
		return nil, err
	}

	secrets := make(map[string]string, len(raw))
	for k, val := range raw {
		str, ok := val.(string)
		if !ok {
			return nil, &CodedError{
				Code:     ECodeBackendKeyNotFound,
				Category: CategoryBackend,
				Message:  fmt.Sprintf("vault key %q at path %q: expected string, got %T", k, secretPath, val),
			}
		}
		secrets[k] = str
	}
	return secrets, nil
}

func (v *VaultBackend) readRaw(ctx context.Context, secretPath string) (map[string]interface{}, error) {
	switch v.version {
	case KVv1:
		resp, err := v.client.Secrets.KvV1Read(ctx, secretPath, vault.WithMountPath(v.mountPath))
		if err != nil {
			return nil, &CodedError{
				Code:     ECodeBackendTimeout,
				Category: CategoryBackend,
				Message:  fmt.Sprintf("reading vault kv v1 at %s/%s", v.mountPath, secretPath),
				Cause:    err,
			}
		}
		return resp.Data, nil
	case KVv2:
		resp, err := v.client.Secrets.KvV2Read(ctx, secretPath, vault.WithMountPath(v.mountPath))
		if err != nil {
			return nil, &CodedError{
				Code:     ECodeBackendTimeout,
				Category: CategoryBackend,
				Message:  fmt.Sprintf("reading vault kv v2 at %s/%s", v.mountPath, secretPath),
				Cause:    err,
			}
		}
		return resp.Data.Data, nil
	default:
		return nil, &CodedError{
			Code:     ECodeInternal,
			Category: CategoryBug,
			Message:  fmt.Sprintf("unsupported KV version %d", v.version),
		}
	}
}
