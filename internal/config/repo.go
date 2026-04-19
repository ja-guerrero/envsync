package config

import "github.com/ja-guerrero/envsync/internal/schema"

type RepoConfig struct {
	Version      int                       `yaml:"version"`
	Project      *Project                  `yaml:"project,omitempty"`
	Vars         schema.Schema             `yaml:"vars"`
	Environments map[string]EnvironmentRef `yaml:"environments,omitempty"`
}

type Project struct {
	Name string `yaml:"name"`
}

type EnvironmentRef struct {
	Backend string                 `yaml:"backend"`
	Params  map[string]interface{} `yaml:",inline"`
}
