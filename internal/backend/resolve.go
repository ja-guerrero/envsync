package backend

import (
	"os"
	"path/filepath"
	"strings"
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

// enrichVaultParams copies params and fills in addr/token from environment
// variables or ~/.vault-token if they are not already present.
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

// readVaultTokenFile reads the token from ~/.vault-token (written by `vault login`).
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
