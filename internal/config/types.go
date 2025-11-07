package config

type Registry struct {
	Projects []RegProject `yaml:"projects"`
}

type RegProject struct {
	Name string `yaml:"name"`
	Meta string `yaml:"meta"`
	Root string `yaml:"root"`
}

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

type ParamMeta struct {
	Required bool   `yaml:"required"`
	Default  string `yaml:"default,omitempty"`
}
type FuncDef struct {
	Params map[string]ParamMeta `yaml:"params"`
	// may be string or []string
	Script any `yaml:"script"`
}

type CommandDef struct {
	Description string `yaml:"description"`
	// may be string or []string
	Cmd any `yaml:"cmd"`
}

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

type DockerDef struct {
	ComposeFile string              `yaml:"compose_file"`
	Groups      map[string][]string `yaml:"groups"`
}

type GlobalConfig struct {
	// allow same structure as ProjectMeta for func/global vars
	Func map[string]FuncDef `yaml:"func"`
	// any other fields accessible via #{global.*} → we keep whole map
	Raw map[string]any `yaml:"-"`
}
