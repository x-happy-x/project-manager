package config

// Registry holds a list of registered projects.
type Registry struct {
	Projects []RegProject `yaml:"projects"`
}

// RegProject represents a registered project entry.
type RegProject struct {
	Name string `yaml:"name"`
	Meta string `yaml:"meta"`
	Root string `yaml:"root"`
}

// ProjectMeta contains project metadata and configuration.
type ProjectMeta struct {
	Info struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Root        string `yaml:"root"`
	} `yaml:"info"`

	Func     map[string]FuncDef    `yaml:"func"`
	Commands map[string]CommandDef `yaml:"commands"`
	Docker   DockerDef             `yaml:"docker"`
}

// ParamMeta defines metadata for function parameters.
type ParamMeta struct {
	Required bool   `yaml:"required"`
	Default  string `yaml:"default,omitempty"`
}

// FuncDef defines a function with parameters and script.
type FuncDef struct {
	Params map[string]ParamMeta `yaml:"params"`
	// may be string or []string
	Script any `yaml:"script"`
}

// CommandDef defines a command with description and command lines.
type CommandDef struct {
	Description string `yaml:"description"`
	// may be string or []string
	Cmd any `yaml:"cmd"`
}

// AsLines converts the command to a slice of strings.
func (c CommandDef) AsLines() []string {
	switch v := c.Cmd.(type) {
	case string:
		return []string{v}
	case []any:
		out := make([]string, 0, len(v))
		for _, it := range v {
			if s, ok := it.(string); ok {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

// DockerDef defines Docker Compose configuration.
type DockerDef struct {
	ComposeFile string              `yaml:"compose_file"`
	Groups      map[string][]string `yaml:"groups"`
}

// GlobalConfig defines global configuration settings.
type GlobalConfig struct {
	// allow same structure as ProjectMeta for func/global vars
	Func map[string]FuncDef `yaml:"func"`
	// any other fields accessible via #{global.*} → we keep whole map
	Raw map[string]any `yaml:"-"`
}
