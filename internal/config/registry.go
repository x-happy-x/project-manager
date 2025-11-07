package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func pmHome() string {
	if v := os.Getenv("PM_CONFIGS"); v != "" {
		return expand(v)
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "pm")
}

func registryFile() string { return filepath.Join(pmHome(), "registry.yml") }
func globalFile() string   { return filepath.Join(pmHome(), "global.yml") }

func ensureHome() error {
	return os.MkdirAll(pmHome(), 0o755)
}

func loadRegistry() (*Registry, error) {
	if err := ensureHome(); err != nil {
		return nil, err
	}
	path := registryFile()
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			reg := &Registry{Projects: []RegProject{}}
			if err := saveRegistry(reg); err != nil {
				return nil, err
			}
			return reg, nil
		}
		return nil, err
	}
	var reg Registry
	if err := yaml.Unmarshal(b, &reg); err != nil {
		return nil, err
	}
	if reg.Projects == nil {
		reg.Projects = []RegProject{}
	}
	return &reg, nil
}

func saveRegistry(r *Registry) error {
	if err := ensureHome(); err != nil {
		return err
	}
	b, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(registryFile(), b, 0o644)
}

func RegAdd(metaPath string) error {
	metaPath = expand(metaPath)
	if _, err := os.Stat(metaPath); err != nil {
		return fmt.Errorf("meta file not found: %s", metaPath)
	}
	meta, err := LoadProjectMeta(metaPath)
	if err != nil {
		return err
	}
	if meta.Info.Name == "" || meta.Info.Root == "" {
		return errors.New("meta.yml must contain info.name and info.root")
	}
	reg, err := loadRegistry()
	if err != nil {
		return err
	}
	// dedup by name
	out := make([]RegProject, 0, len(reg.Projects))
	for _, p := range reg.Projects {
		if p.Name != meta.Info.Name {
			out = append(out, p)
		}
	}
	out = append(out, RegProject{
		Name: meta.Info.Name,
		Meta: abs(metaPath),
		Root: abs(expand(meta.Info.Root)),
	})
	reg.Projects = out
	return saveRegistry(reg)
}

func RegRm(name string) error {
	reg, err := loadRegistry()
	if err != nil {
		return err
	}
	changed := false
	out := make([]RegProject, 0, len(reg.Projects))
	for _, p := range reg.Projects {
		if p.Name == name {
			changed = true
			continue
		}
		out = append(out, p)
	}
	if !changed {
		return fmt.Errorf("project not found: %s", name)
	}
	reg.Projects = out
	return saveRegistry(reg)
}

func RegLs() error {
	reg, err := loadRegistry()
	if err != nil {
		return err
	}
	if len(reg.Projects) == 0 {
		fmt.Println("# pm: empty. Use: pm add /path/to/.pm.meta.yml")
		return nil
	}
	fmt.Print("# pm: projects\n\n")
	for _, p := range reg.Projects {
		fmt.Printf("- %s\n  meta: %s\n  root: %s\n", p.Name, p.Meta, p.Root)
	}
	return nil
}

func ResolveProject(nameOrPath string) (*ProjectMeta, string, error) {
	cand := expand(nameOrPath)
	if fileExists(cand) && (extEq(cand, ".yml") || extEq(cand, ".yaml")) {
		meta, err := LoadProjectMeta(cand)
		if err != nil {
			return nil, "", err
		}
		root := abs(expand(meta.Info.Root))
		return meta, root, nil
	}
	reg, err := loadRegistry()
	if err != nil {
		return nil, "", err
	}
	for _, p := range reg.Projects {
		if p.Name == nameOrPath {
			meta, err := LoadProjectMeta(p.Meta)
			if err != nil {
				return nil, "", err
			}
			return meta, abs(p.Root), nil
		}
	}
	return nil, "", fmt.Errorf("project not found in registry or file does not exist: %s", nameOrPath)
}

func LoadProjectMeta(path string) (*ProjectMeta, error) {
	b, err := os.ReadFile(expand(path))
	if err != nil {
		return nil, err
	}
	var m ProjectMeta
	if err := yaml.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

func LoadGlobal() (*GlobalConfig, error) {
	path := globalFile()
	b, err := os.ReadFile(path)
	if err != nil {
		return &GlobalConfig{Func: map[string]FuncDef{}, Raw: map[string]any{}}, nil
	}
	var gc GlobalConfig
	if err := yaml.Unmarshal(b, &gc); err != nil {
		return nil, err
	}
	// keep raw
	var raw map[string]any
	if err := yaml.Unmarshal(b, &raw); err == nil {
		gc.Raw = raw
	}
	if gc.Func == nil {
		gc.Func = map[string]FuncDef{}
	}
	if gc.Raw == nil {
		gc.Raw = map[string]any{}
	}
	return &gc, nil
}

func expand(s string) string {
	s = os.ExpandEnv(s)
	if strings.HasPrefix(s, "~") {
		home, _ := os.UserHomeDir()
		if home != "" {
			s = filepath.Join(home, s[1:])
		}
	}
	return s
}
func abs(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}
func fileExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
func extEq(p, ext string) bool {
	return strings.EqualFold(filepath.Ext(p), ext)
}
