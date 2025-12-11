package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pm/internal/plan"
)

type Renderer interface {
	Name() string
	Begin(root string) []string
	RenderOp(op plan.Op) []string
	End() []string
}

var builtins = map[string]Renderer{
	"bash": bashRenderer{},
	"zsh":  bashRenderer{}, // alias
	"pwsh": pwshRenderer{},
}

type externalPlan struct {
	Root string       `json:"root"`
	Ops  []externalOp `json:"ops"`
}
type externalOp struct {
	Kind string `json:"kind"`
	Line string `json:"line,omitempty"`
	Dir  string `json:"dir,omitempty"`
	Msg  string `json:"msg,omitempty"`
}

func Render(pl *plan.Plan, dialect, pluginsDir string) (string, error) {
	if r, ok := builtins[dialect]; ok {
		return renderWith(r, pl), nil
	}
	// external plugin: executable pm-render-<dialect> in pluginsDir
	pluginsDir = expand(pluginsDir)
	exe := filepath.Join(pluginsDir, "pm-render-"+dialect)
	if _, err := os.Stat(exe); err != nil {
		return "", fmt.Errorf("renderer not found: %s (plugins dir %s)", dialect, pluginsDir)
	}
	return renderExternal(exe, pl)
}

func renderWith(r Renderer, pl *plan.Plan) string {
	var buf bytes.Buffer
	begin := false
	for i, op := range pl.Ops {
		if !begin {
			// find first pushd to get root
			if v, ok := op.(plan.OpPushd); ok {
				for _, l := range r.Begin(v.Dir) {
					buf.WriteString(l + "\n")
				}
				begin = true
				continue
			}
			// if first op not pushd, still print begin with cwd "."
			if i == 0 {
				for _, l := range r.Begin(".") {
					buf.WriteString(l + "\n")
				}
				begin = true
			}
		}
		for _, l := range r.RenderOp(op) {
			buf.WriteString(l + "\n")
		}
	}
	for _, l := range r.End() {
		buf.WriteString(l + "\n")
	}
	return buf.String()
}

func renderExternal(exe string, pl *plan.Plan) (string, error) {
	root := "."
	var ops []externalOp
	for _, op := range pl.Ops {
		switch v := op.(type) {
		case plan.OpPushd:
			if root == "." {
				root = v.Dir
			}
			ops = append(ops, externalOp{Kind: "pushd", Dir: v.Dir})
		case plan.OpPopd:
			ops = append(ops, externalOp{Kind: "popd"})
		case plan.OpEcho:
			ops = append(ops, externalOp{Kind: "echo", Msg: v.Line})
		case plan.OpRun:
			ops = append(ops, externalOp{Kind: "run", Line: v.Line})
		}
	}
	payload := externalPlan{Root: root, Ops: ops}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal plan: %w", err)
	}
	cmd := exec.Command(exe, "--render")
	cmd.Stdin = bytes.NewReader(b)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

func expand(s string) string {
	s = os.ExpandEnv(s)
	if strings.HasPrefix(s, "~") {
		if home, _ := os.UserHomeDir(); home != "" {
			return filepath.Join(home, s[1:])
		}
	}
	return s
}
