package render

import (
	"strings"
	"testing"

	"pm/internal/plan"
)

func buildPlan() *plan.Plan {
	p := plan.New()
	p.Pushd("/tmp/project")
	p.Run("echo hello")
	p.Popd()
	return p
}

func TestRender_Bash(t *testing.T) {
	s, err := Render(buildPlan(), "bash", "")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	for _, sub := range []string{"pushd", "echo hello", "popd"} {
		if !strings.Contains(s, sub) {
			t.Fatalf("want %q in script:\n%s", sub, s)
		}
	}
}

func TestRender_Pwsh(t *testing.T) {
	s, err := Render(buildPlan(), "pwsh", "")
	if err != nil {
		t.Fatalf("render error: %v", err)
	}
	for _, sub := range []string{"Push-Location", "echo hello", "Pop-Location"} {
		if !strings.Contains(s, sub) {
			t.Fatalf("want %q in script:\n%s", sub, s)
		}
	}
}
