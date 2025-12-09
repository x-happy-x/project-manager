package test_utils

import (
	"os"
	"path/filepath"
	"pm/internal/config"
	"pm/internal/dsl"
	"pm/internal/plan"
	"pm/internal/render"
	"pm/internal/templ"
	"strings"
	"testing"
)

// TestCase описывает один тестовый сценарий
type TestCase struct {
	Name         string            // имя теста
	MetaFile     string            // файл с .pm.meta.yml
	GlobalFile   string            // опциональный файл global.yml
	ExpectedFile string            // файл с ожидаемым выводом
	Command      string            // команда для выполнения (например, "rawproj docker compose ps")
	Dialect      string            // bash или pwsh
	EnvVars      map[string]string // дополнительные переменные окружения
}

// RunTestCase выполняет один тестовый сценарий
func RunTestCase(t *testing.T, tc TestCase) {
	t.Helper()

	td := t.TempDir()
	t.Setenv("PM_CONFIGS", td)

	// Установим дополнительные переменные окружения, если есть
	for k, v := range tc.EnvVars {
		t.Setenv(k, v)
	}

	// Создаём проектную директорию
	projDir := MustMkdir(t, filepath.Join(td, tc.Name))

	// Читаем и подготавливаем meta файл
	metaContent := LoadTestdata(t, tc.MetaFile)
	metaContent = strings.ReplaceAll(metaContent, "__PROJECT_DIR__", projDir)
	metaPath := WriteMeta(t, projDir, metaContent)

	// Если есть global файл, создаём его
	if tc.GlobalFile != "" {
		globalContent := LoadTestdata(t, tc.GlobalFile)
		WriteGlobal(t, td, globalContent)
	}

	// Регистрируем проект
	if err := config.RegAdd(metaPath); err != nil {
		t.Fatalf("RegAdd: %v", err)
	}

	// Генерируем скрипт
	script := GenerateScript(t, tc.Command, tc.Dialect)

	// Читаем ожидаемый результат
	expected := LoadTestdata(t, tc.ExpectedFile)
	expected = strings.ReplaceAll(expected, "__PROJECT_DIR__", projDir)

	// Проверяем, что все строки из expected присутствуют в script
	for _, line := range strings.Split(strings.TrimSpace(expected), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		AssertContains(t, script, line)
	}
}

// loadTestdata загружает содержимое файла из testdata
func LoadTestdata(t *testing.T, filename string) string {
	t.Helper()
	path := filepath.Join("testdata", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read testdata file %s: %v", path, err)
	}
	return string(data)
}

func GenerateScript(t *testing.T, command string, dialect string) string {
	t.Helper()

	tail := strings.Split(command, " ")
	projectRef := tail[0]
	tail = tail[1:]

	meta, root, err := config.ResolveProject(projectRef)
	if err != nil {
		t.Fatalf("ResolveProject: %v", err)
	}
	global, _ := config.LoadGlobal()

	chunks := dsl.SplitColonCommands(tail)
	pl := plan.New()
	pl.Pushd(root)

	// raw режим
	if len(chunks) == 1 && chunks[0].Name == "__RAW__" {
		line := strings.Join(chunks[0].Args, " ")
		if strings.TrimSpace(line) != "" {
			pl.Run(line)
		}
		pl.Popd()
		out, err := render.Render(pl, dialect, "")
		if err != nil {
			t.Fatalf("render: %v", err)
		}
		return out
	}

	// именованные :команды
	for _, ch := range chunks {
		if ch.Name == "help" {
			pl.Echo("# pm: project commands:")
			for k, v := range meta.Commands {
				desc := strings.TrimSpace(v.Description)
				pl.Echo("  :" + k + "  - " + desc)
			}
			pl.Echo("# pm: built-ins: :up (docker compose up -d ...)")
			continue
		}
		if ch.Name == "up" {
			cmds := plan.DockerUp(meta, ch.Args)
			for _, c := range cmds {
				pl.Run(c)
			}
			continue
		}
		cmd, ok := meta.Commands[ch.Name]
		if !ok {
			pl.Echo("# pm: unknown command :" + ch.Name)
			continue
		}
		params := map[string]string{"args": strings.Join(ch.Args, " ")}
		for _, raw := range cmd.AsLines() {
			rendered, err := templ.RenderString(raw, params, meta, global, nil)
			if err != nil {
				pl.Echo("# pm: template error: " + err.Error())
				continue
			}
			if strings.TrimSpace(rendered) != "" {
				pl.Run(rendered)
			}
		}
	}

	pl.Popd()
	out, err := render.Render(pl, dialect, "")
	if err != nil {
		t.Fatalf("render: %v", err)
	}

	t.Log("Generated script: " + out)
	return out
}

func MustMkdir(t *testing.T, p string) string {
	t.Helper()
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", p, err)
	}
	return p
}

func WriteMeta(t *testing.T, dir string, content string) string {
	t.Helper()
	path := filepath.Join(dir, ".pm.meta.yml")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write meta: %v", err)
	}
	return path
}

func WriteGlobal(t *testing.T, pmHome string, content string) {
	t.Helper()
	gdir := filepath.Join(pmHome)
	if err := os.MkdirAll(gdir, 0o755); err != nil {
		t.Fatalf("mkdir global dir: %v", err)
	}
	gfile := filepath.Join(gdir, "global.yml")
	if err := os.WriteFile(gfile, []byte(content), 0o644); err != nil {
		t.Fatalf("write global: %v", err)
	}
}

func AssertContains(t *testing.T, s, sub string) {
	t.Helper()
	if !strings.Contains(s, sub) {
		t.Fatalf("expected script to contain:\n  %q\n--- script ---\n%s", sub, s)
	}
}
