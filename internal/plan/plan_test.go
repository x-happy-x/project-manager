package plan

import (
	"strings"
	"testing"

	"pm/internal/config"
)

func TestDockerUp_Basic(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "compose.yml",
			Groups: map[string][]string{
				"base": {"db", "redis"},
			},
		},
	}
	cmds := DockerUp(meta, []string{"@base", "api"})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if !strings.Contains(cmd, "docker compose -f") {
		t.Fatalf("missing docker compose -f: %s", cmd)
	}
	if !strings.Contains(cmd, "up -d") {
		t.Fatalf("missing up -d: %s", cmd)
	}
	for _, s := range []string{"db", "redis", "api"} {
		if !strings.Contains(cmd, s) {
			t.Fatalf("expected %q in %q", s, cmd)
		}
	}
}
