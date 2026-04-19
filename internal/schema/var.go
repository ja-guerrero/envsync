package schema

type Schema map[string]Var

type Var struct {
	Type        string            `yaml:"type,omitempty"`
	Required    bool              `yaml:"required,omitempty"`
	RequiredIn  []string          `yaml:"required_in,omitempty"`
	RequiredIf  map[string]string `yaml:"required_if,omitempty"`
	Default     *string           `yaml:"default,omitempty"`
	Secret      bool              `yaml:"secret,omitempty"`
	Format      string            `yaml:"format,omitempty"`
	Enum        []string          `yaml:"enum,omitempty"`
	Min         *float64          `yaml:"min,omitempty"`
	Max         *float64          `yaml:"max,omitempty"`
	Description string            `yaml:"description,omitempty"`
}
