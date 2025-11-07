package templ

import (
	"testing"

	"pm/internal/config"
)

func metaForTest() *config.ProjectMeta {
	return &config.ProjectMeta{
		Info: struct {
			Name        string "yaml:\"name\""
			Description string "yaml:\"description\""
			Root        string "yaml:\"root\""
		}{
			Name: "subzero", Description: "desc", Root: "/proj",
		},
		Func: map[string]config.FuncDef{
			"use-java": {
				Params: map[string]config.ParamMeta{
					"version": {Required: true, Default: ""},
				},
				Script: "sdk use java @{version}",
			},
		},
		Commands: nil,
		Docker:   config.DockerDef{},
	}
}

func globalForTest() *config.GlobalConfig {
	return &config.GlobalConfig{
		Func: map[string]config.FuncDef{
			"say": {Script: "echo @{text}"},
		},
		Raw: map[string]any{
			"vars": map[string]any{"x": "42"},
		},
	}
}

func TestRender_ParamsEnvCfg(t *testing.T) {
	meta := metaForTest()
	global := globalForTest()

	src := "E=#{info.name} X=#{global.vars.x} HOME=${HOME} ARG=@{p}"
	out, err := RenderString(src, map[string]string{"p": "v"}, meta, global, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if want := "E=subzero"; !contains(out, want) {
		t.Fatalf("want %q in %q", want, out)
	}
	if want := "X=42"; !contains(out, want) {
		t.Fatalf("want %q in %q", want, out)
	}
	if want := "ARG=v"; !contains(out, want) {
		t.Fatalf("want %q in %q", want, out)
	}
}

func TestRender_FunctionCall_LocalAndGlobal(t *testing.T) {
	meta := metaForTest()
	global := globalForTest()

	src := "_{use-java(version=21.0.8-tem)} && _{global.say(text='hi')}"
	out, err := RenderString(src, nil, meta, global, nil)
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	if want := "sdk use java 21.0.8-tem"; !contains(out, want) {
		t.Fatalf("want %q in %q", want, out)
	}
	if want := "echo hi"; !contains(out, want) {
		t.Fatalf("want %q in %q", want, out)
	}
}

func contains(s, sub string) bool { return len(s) >= len(sub) && (stringIndex(s, sub) >= 0) }

// tiny inline strstr to avoid extra imports
func stringIndex(s, sub string) int {
outer:
	for i := 0; i+len(sub) <= len(s); i++ {
		for j := 0; j < len(sub); j++ {
			if s[i+j] != sub[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
