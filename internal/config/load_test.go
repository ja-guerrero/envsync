package config

import (
	"testing"
)

func TestLoadRepoConfig(t *testing.T) {
	cfg, err := LoadRepoConfig("testdata/envschema.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != 1 {
		t.Errorf("version: got %d, want 1", cfg.Version)
	}

	if cfg.Project == nil || cfg.Project.Name != "my-app" {
		t.Errorf("project name: got %v, want my-app", cfg.Project)
	}

	db, ok := cfg.Vars["DATABASE_URL"]
	if !ok {
		t.Fatal("expected DATABASE_URL in vars")
	}
	if !db.Required {
		t.Error("DATABASE_URL should be required")
	}
	if !db.Secret {
		t.Error("DATABASE_URL should be secret")
	}

	if _, ok := cfg.Environments["production"]; !ok {
		t.Error("expected production environment")
	}
}

func TestLoadRepoConfig_FileNotFound(t *testing.T) {
	_, err := LoadRepoConfig("testdata/nonexistent.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadRepoConfig_InvalidYAML(t *testing.T) {
	_, err := LoadRepoConfig("testdata/invalid.yaml")
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

// KnownFields(true) is set on the decoder, so unknown fields are a parse error.
func TestLoadRepoConfig_Typo(t *testing.T) {
	_, err := LoadRepoConfig("testdata/typo.yaml")
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}
}

// required: true with a default set is a semantic contradiction caught by the validator.
func TestLoadRepoConfig_RequiredWithDefault(t *testing.T) {
	_, err := LoadRepoConfig("testdata/required_with_default.yaml")
	if err == nil {
		t.Fatal("expected validation error for required+default, got nil")
	}
}
