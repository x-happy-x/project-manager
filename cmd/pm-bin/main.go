package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"pm/internal/config"
	"pm/internal/dsl"
	"pm/internal/plan"
	"pm/internal/render"
	"pm/internal/templ"
)

func main() {
	var (
		dialect  string
		plugins  string
		showHelp bool
	)
	flag.StringVar(&dialect, "dialect", "", "script dialect to render (bash|pwsh|<external-plugin>) [env PM_DIALECT]")
	flag.StringVar(&plugins, "plugins", "", "plugins dir (defaults to ~/.config/pm/plugins) [env PM_PLUGIN_DIR]")
	flag.BoolVar(&showHelp, "h", false, "show help")
	flag.Parse()

	args := flag.Args()
	if showHelp || len(args) == 0 {
		fmt.Print(`# pm: usage
#   pm-bin [--dialect bash|pwsh|<plugin>] [--plugins DIR] <add|rm|ls|PROJECT|META.yml> [args...]
# examples:
#   pm-bin add ~/repos/subzero/.pm.meta.yml
#   pm-bin ls
#   pm-bin --dialect bash subzero :build -DskipTests :up @base api
`)
		return
	}

	// top-level commands that should print info (no script!)
	switch args[0] {
	case "add":
		if len(args) < 2 {
			fail("pm add /path/to/.pm.meta.yml")
		}
		if err := config.RegAdd(args[1]); err != nil {
			fail(err.Error())
		}
		fmt.Printf("# pm: added project\n")
		return
	case "rm":
		if len(args) < 2 {
			fail("pm rm <name>")
		}
		if err := config.RegRm(args[1]); err != nil {
			fail(err.Error())
		}
		fmt.Printf("# pm: removed project\n")
		return
	case "ls":
		if err := config.RegLs(); err != nil {
			fail(err.Error())
		}
		return
	}

	// else: build plan for project or meta.yml path
	projectRef := args[0]
	tail := args[1:]

	meta, root, err := config.ResolveProject(projectRef)
	if err != nil {
		fail(err.Error())
	}

	globalCfg, _ := config.LoadGlobal()

	chunks := dsl.SplitColonCommands(tail)
	pl := plan.New()
	pl.Pushd(root)

	// raw mode
	if len(chunks) == 1 && chunks[0].Name == "__RAW__" {
		line := strings.Join(chunks[0].Args, " ")
		if strings.TrimSpace(line) != "" {
			pl.Run(line)
		}
		renderAndPrint(pl, dialect, plugins)
		return
	}

	// named :commands
	for _, ch := range chunks {
		if ch.Name == "help" {
			pl.Echo("# pm: project commands:")
			for k, v := range meta.Commands {
				desc := strings.TrimSpace(v.Description)
				if desc == "" {
					desc = "-"
				}
				pl.Echo(fmt.Sprintf("  :%s  - %s", k, desc))
			}
			pl.Echo("# pm: built-ins: :up (docker compose up -d ...)")
			continue
		}
		// built-in :up
		if ch.Name == "up" {
			cmds := plan.DockerUp(meta, ch.Args)
			for _, c := range cmds {
				pl.Run(c)
			}
			continue
		}
		// user-defined
		cmd, ok := meta.Commands[ch.Name]
		if !ok {
			pl.Echo(fmt.Sprintf("# pm: unknown command :%s", ch.Name))
			continue
		}
		params := map[string]string{
			"args": strings.Join(ch.Args, " "),
		}
		cmdLines := cmd.AsLines()
		for _, raw := range cmdLines {
			rendered, err := templ.RenderString(raw, params, meta, globalCfg, nil)
			if err != nil {
				pl.Echo("# pm: template error: " + err.Error())
				continue
			}
			if strings.TrimSpace(rendered) != "" {
				pl.Run(rendered)
			}
		}
	}

	renderAndPrint(pl, dialect, plugins)
}

func renderAndPrint(pl *plan.Plan, dialect, plugins string) {
	// select dialect
	if dialect == "" {
		if v := os.Getenv("PM_DIALECT"); v != "" {
			dialect = v
		} else {
			// default per OS
			if isWindows() {
				dialect = "pwsh"
			} else {
				dialect = "bash"
			}
		}
	}
	if plugins == "" {
		if v := os.Getenv("PM_PLUGIN_DIR"); v != "" {
			plugins = v
		} else {
			home, _ := os.UserHomeDir()
			plugins = filepath.Join(home, ".config", "pm", "plugins")
		}
	}

	s, err := render.Render(pl, dialect, plugins)
	if err != nil {
		fail(err.Error())
	}
	fmt.Print(s)
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func isWindows() bool {
	return strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") ||
		(filepath.Separator == '\\')
}
