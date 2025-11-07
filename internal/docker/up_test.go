package docker

import (
	"strings"
	"testing"

	"pm/internal/config"
	"pm/internal/plan"
)

func TestDockerUp_NoArgs(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "docker-compose.yml",
		},
	}
	cmds := plan.DockerUp(meta, []string{})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if !strings.Contains(cmds[0], "docker compose -f") {
		t.Errorf("expected docker compose -f, got %s", cmds[0])
	}
	if !strings.Contains(cmds[0], "up -d") {
		t.Errorf("expected up -d, got %s", cmds[0])
	}
}

func TestDockerUp_WithServices(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "compose.yaml",
		},
	}
	cmds := plan.DockerUp(meta, []string{"api", "db"})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if !strings.Contains(cmd, "api") || !strings.Contains(cmd, "db") {
		t.Errorf("expected api and db in command: %s", cmd)
	}
}

func TestDockerUp_WithGroups(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "docker-compose.yml",
			Groups: map[string][]string{
				"base": {"redis", "postgres"},
				"all":  {"api", "worker"},
			},
		},
	}
	cmds := plan.DockerUp(meta, []string{"@base", "extra"})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	if !strings.Contains(cmd, "redis") || !strings.Contains(cmd, "postgres") {
		t.Errorf("expected redis and postgres from @base: %s", cmd)
	}
	if !strings.Contains(cmd, "extra") {
		t.Errorf("expected extra: %s", cmd)
	}
}

func TestDockerUp_MultipleGroups(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "docker-compose.yml",
			Groups: map[string][]string{
				"db":  {"postgres", "redis"},
				"app": {"api", "worker"},
			},
		},
	}
	cmds := plan.DockerUp(meta, []string{"@db", "@app", "nginx"})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	// should contain all services from both groups
	services := []string{"postgres", "redis", "api", "worker", "nginx"}
	for _, s := range services {
		if !strings.Contains(cmd, s) {
			t.Errorf("expected %s in command: %s", s, cmd)
		}
	}
}

func TestDockerUp_UnknownGroup(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "docker-compose.yml",
			Groups: map[string][]string{
				"base": {"api"},
			},
		},
	}
	// @unknown doesn't exist, should be passed as-is (harmless)
	cmds := plan.DockerUp(meta, []string{"@unknown", "db"})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	cmd := cmds[0]
	// @unknown treated as literal
	if !strings.Contains(cmd, "@unknown") {
		t.Errorf("expected @unknown in command: %s", cmd)
	}
	if !strings.Contains(cmd, "db") {
		t.Errorf("expected db in command: %s", cmd)
	}
}

func TestDockerUp_DefaultComposeFile(t *testing.T) {
	meta := &config.ProjectMeta{
		Docker: config.DockerDef{
			ComposeFile: "",
		},
	}
	cmds := plan.DockerUp(meta, []string{})
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	// should default to docker-compose.yml
	if !strings.Contains(cmds[0], "docker-compose.yml") {
		t.Errorf("expected docker-compose.yml as default: %s", cmds[0])
	}
}
