package e2e

import (
	"path/filepath"
	"strings"
	"testing"

	"pm/internal/config"
	. "pm/internal/e2e/test_utils"
)

func TestE2E_RawCommand(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "raw_command",
		MetaFile:     "raw_command.meta.yml",
		ExpectedFile: "raw_command.expected",
		Command:      "rawproj docker compose ps",
		Dialect:      "bash",
	})
}

func TestE2E_CommandSequence_BuildRun_WithTemplating(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "command_sequence",
		MetaFile:     "command_sequence.meta.yml",
		ExpectedFile: "command_sequence.expected",
		Command:      "subzero :build -Pprod :run",
		Dialect:      "bash",
	})
}

func TestE2E_DockerUp_Groups(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "docker_up_groups",
		MetaFile:     "docker_up_groups.meta.yml",
		ExpectedFile: "docker_up_groups.expected",
		Command:      "dock :up @base api",
		Dialect:      "bash",
	})
}

func TestE2E_GlobalFuncsAndVars(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "global_funcs",
		MetaFile:     "global_funcs.meta.yml",
		GlobalFile:   "global_funcs.global.yml",
		ExpectedFile: "global_funcs.expected",
		Command:      "gproj :hello",
		Dialect:      "bash",
	})
}

func TestE2E_HelpEchoesCommands(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "help_echo",
		MetaFile:     "help_echo.meta.yml",
		ExpectedFile: "help_echo.expected",
		Command:      "helpme :help",
		Dialect:      "bash",
	})
}

func TestE2E_EnvSubstitution(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "env_substitution",
		MetaFile:     "env_substitution.meta.yml",
		ExpectedFile: "env_substitution.expected",
		Command:      "envtest :show",
		Dialect:      "bash",
		EnvVars:      map[string]string{"TEST_VAR": "hello_world"},
	})
}

func TestE2E_ConfigPathSubstitution(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "config_path_substitution",
		MetaFile:     "config_path_substitution.meta.yml",
		ExpectedFile: "config_path_substitution.expected",
		Command:      "cfgtest :info",
		Dialect:      "bash",
	})
}

func TestE2E_MultilineFuncScript(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "multiline_func",
		MetaFile:     "multiline_func.meta.yml",
		ExpectedFile: "multiline_func.expected",
		Command:      "multiline :deploy",
		Dialect:      "bash",
	})
}

func TestE2E_MultipleCommandsWithArgs(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "multiple_commands",
		MetaFile:     "multiple_commands.meta.yml",
		ExpectedFile: "multiple_commands.expected",
		Command:      "multi :clean dist :build -j4 :test ./...",
		Dialect:      "bash",
	})
}

func TestE2E_DockerUpEmptyArgs(t *testing.T) {
	tc := TestCase{
		Name:         "docker_up_empty",
		MetaFile:     "docker_up_empty.meta.yml",
		ExpectedFile: "docker_up_empty.expected",
		Command:      "docker :up",
		Dialect:      "bash",
	}

	td := t.TempDir()
	t.Setenv("PM_CONFIGS", td)

	for k, v := range tc.EnvVars {
		t.Setenv(k, v)
	}

	projDir := MustMkdir(t, filepath.Join(td, tc.Name))
	metaContent := LoadTestdata(t, tc.MetaFile)
	metaContent = strings.ReplaceAll(metaContent, "__PROJECT_DIR__", projDir)
	metaPath := WriteMeta(t, projDir, metaContent)

	if tc.GlobalFile != "" {
		globalContent := LoadTestdata(t, tc.GlobalFile)
		WriteGlobal(t, td, globalContent)
	}

	if err := config.RegAdd(metaPath); err != nil {
		t.Fatalf("RegAdd: %v", err)
	}

	script := GenerateScript(t, tc.Command, tc.Dialect)

	AssertContains(t, script, "docker compose -f docker-compose.yml up -d")
	// не должно быть конкретных сервисов
	if strings.Contains(script, "api") || strings.Contains(script, "db") {
		t.Errorf("unexpected services in output: %s", script)
	}
}

func TestE2E_PowerShellRenderer(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "powershell",
		MetaFile:     "powershell.meta.yml",
		ExpectedFile: "powershell.expected",
		Command:      "pwsh :run",
		Dialect:      "pwsh",
	})
}

func TestE2E_UnknownCommand(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "unknown_command",
		MetaFile:     "unknown_command.meta.yml",
		ExpectedFile: "unknown_command.expected",
		Command:      "unknown :nonexistent",
		Dialect:      "bash",
	})
}

func TestE2E_GlobalVarsAndFuncs(t *testing.T) {
	RunTestCase(t, TestCase{
		Name:         "global_vars_funcs",
		MetaFile:     "global_vars_funcs.meta.yml",
		GlobalFile:   "global_vars_funcs.global.yml",
		ExpectedFile: "global_vars_funcs.expected",
		Command:      "global :welcome",
		Dialect:      "bash",
	})
}

func TestE2E_FuncWithoutRequiredParam(t *testing.T) {
	tc := TestCase{
		Name:         "func_required_param",
		MetaFile:     "func_required_param.meta.yml",
		ExpectedFile: "func_required_param.expected",
		Command:      "reqparam :bad",
		Dialect:      "bash",
	}

	td := t.TempDir()
	t.Setenv("PM_CONFIGS", td)

	for k, v := range tc.EnvVars {
		t.Setenv(k, v)
	}

	projDir := MustMkdir(t, filepath.Join(td, tc.Name))
	metaContent := LoadTestdata(t, tc.MetaFile)
	metaContent = strings.ReplaceAll(metaContent, "__PROJECT_DIR__", projDir)
	metaPath := WriteMeta(t, projDir, metaContent)

	if tc.GlobalFile != "" {
		globalContent := LoadTestdata(t, tc.GlobalFile)
		WriteGlobal(t, td, globalContent)
	}

	if err := config.RegAdd(metaPath); err != nil {
		t.Fatalf("RegAdd: %v", err)
	}

	script := GenerateScript(t, tc.Command, tc.Dialect)

	// Специальная проверка для этого теста
	if strings.Contains(script, "echo Deploying to") {
		t.Errorf("function should not execute without required param")
	}
}
