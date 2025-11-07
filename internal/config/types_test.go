package config

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestCommandDef_AsLines_String(t *testing.T) {
	cmd := CommandDef{Cmd: "echo hello"}
	lines := cmd.AsLines()
	if len(lines) != 1 || lines[0] != "echo hello" {
		t.Errorf("expected [echo hello], got %v", lines)
	}
}

func TestCommandDef_AsLines_ArrayAny(t *testing.T) {
	cmd := CommandDef{Cmd: []any{"line1", "line2"}}
	lines := cmd.AsLines()
	if len(lines) != 2 || lines[0] != "line1" || lines[1] != "line2" {
		t.Errorf("expected [line1 line2], got %v", lines)
	}
}

func TestCommandDef_AsLines_ArrayString(t *testing.T) {
	cmd := CommandDef{Cmd: []string{"a", "b", "c"}}
	lines := cmd.AsLines()
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}
}

func TestCommandDef_AsLines_Nil(t *testing.T) {
	cmd := CommandDef{}
	lines := cmd.AsLines()
	if lines != nil {
		t.Errorf("expected nil, got %v", lines)
	}
}

func TestParamMeta_YAML(t *testing.T) {
	yml := `
required: true
default: "test"
`
	var pm ParamMeta
	if err := yaml.Unmarshal([]byte(yml), &pm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !pm.Required || pm.Default != "test" {
		t.Errorf("expected Required=true, Default=test, got %+v", pm)
	}
}

func TestParamMeta_NoDefault(t *testing.T) {
	yml := `
required: true
`
	var pm ParamMeta
	if err := yaml.Unmarshal([]byte(yml), &pm); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !pm.Required || pm.Default != "" {
		t.Errorf("expected Required=true, Default=\"\", got %+v", pm)
	}
}

func TestProjectMeta_UnmarshalYAML(t *testing.T) {
	yml := `
info:
  name: test
  description: Test project
  root: /test
func:
  myfunc:
    params:
      version:
        required: true
    script: "echo @{version}"
commands:
  build:
    description: Build
    cmd: "make build"
docker:
  compose_file: docker-compose.yml
  groups:
    base: [api, db]
`
	var meta ProjectMeta
	if err := yaml.Unmarshal([]byte(yml), &meta); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if meta.Info.Name != "test" {
		t.Errorf("expected name=test, got %s", meta.Info.Name)
	}
	if _, ok := meta.Func["myfunc"]; !ok {
		t.Errorf("expected myfunc in Func")
	}
	if _, ok := meta.Commands["build"]; !ok {
		t.Errorf("expected build in Commands")
	}
	if len(meta.Docker.Groups["base"]) != 2 {
		t.Errorf("expected 2 services in base group")
	}
}
