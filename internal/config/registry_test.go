package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_AddResolveRm(t *testing.T) {
	td := t.TempDir()
	t.Setenv("PM_CONFIGS", td)

	// create project with meta
	projDir := filepath.Join(td, "proj")
	if err := os.MkdirAll(projDir, 0o755); err != nil {
		t.Fatal(err)
	}
	metaPath := filepath.Join(projDir, ".pm.meta.yml")
	meta := `
info:
  name: zeta
  description: test
  root: ` + projDir + `
commands:
  run:
    description: run
    cmd: echo ok
`
	if err := os.WriteFile(metaPath, []byte(meta), 0o644); err != nil {
		t.Fatal(err)
	}

	// add
	if err := RegAdd(metaPath); err != nil {
		t.Fatalf("RegAdd: %v", err)
	}

	// resolve by name
	pm, root, err := ResolveProject("zeta")
	if err != nil {
		t.Fatalf("ResolveProject: %v", err)
	}
	if pm.Info.Name != "zeta" || root != projDir {
		t.Fatalf("resolve mismatch: %s %s", pm.Info.Name, root)
	}

	// resolve by path
	pm2, root2, err := ResolveProject(metaPath)
	if err != nil {
		t.Fatalf("ResolveProject by path: %v", err)
	}
	if pm2.Info.Name != "zeta" || root2 != projDir {
		t.Fatalf("resolve by path mismatch")
	}

	// rm
	if err := RegRm("zeta"); err != nil {
		t.Fatalf("RegRm: %v", err)
	}
}
