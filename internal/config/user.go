// internal/config/user.go
package config

type UserConfig struct {
	Version  int                `yaml:"version"`
	Backends []BackendInstance  `yaml:"backends"`
	Defaults map[string]string  `yaml:"defaults,omitempty"`
	Profiles map[string]Profile `yaml:"profiles,omitempty"`
}

type BackendInstance struct {
	Name   string                 `yaml:"name"`
	Type   string                 `yaml:"type"`
	Params map[string]interface{} `yaml:",inline"`
}

type Profile struct {
	Defaults map[string]string `yaml:"defaults"`
}
