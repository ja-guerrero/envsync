package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func LoadRepoConfig(path string) (*RepoConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening repo config: %w", err)
	}
	defer f.Close()

	var cfg RepoConfig
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)

	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing repo config: %w", err)
	}

	if err := validateRepoConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func LoadUserConfig(path string) (*UserConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("reading user config: %w", err)
	}
	defer f.Close()

	var cfg UserConfig
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parsing user config: %w", err)
	}

	if err := ValidateUserConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
