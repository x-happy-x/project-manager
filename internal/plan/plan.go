package plan

import (
	"fmt"
	"strings"

	"pm/internal/config"
)

type Op interface{ isOp() }
type OpPushd struct{ Dir string }
type OpPopd struct{}
type OpRun struct{ Line string }
type OpEcho struct{ Line string }

func (OpPushd) isOp() {}
func (OpPopd) isOp()  {}
func (OpRun) isOp()   {}
func (OpEcho) isOp()  {}

type Plan struct {
	Ops []Op
}

func New() *Plan { return &Plan{Ops: []Op{}} }

func (p *Plan) Pushd(dir string) { p.Ops = append(p.Ops, OpPushd{Dir: dir}) }
func (p *Plan) Popd()            { p.Ops = append(p.Ops, OpPopd{}) }
func (p *Plan) Run(line string)  { p.Ops = append(p.Ops, OpRun{Line: line}) }
func (p *Plan) Echo(line string) { p.Ops = append(p.Ops, OpEcho{Line: line}) }

// DockerUp returns shell lines for docker compose up -d with groups
func DockerUp(meta *config.ProjectMeta, args []string) []string {
	d := meta.Docker
	compose := d.ComposeFile
	if strings.TrimSpace(compose) == "" {
		compose = "docker-compose.yml"
	}
	var services []string
	for _, a := range args {
		if strings.HasPrefix(a, "@") {
			g := strings.TrimPrefix(a, "@")
			if grp, ok := d.Groups[g]; ok {
				services = append(services, grp...)
			} else {
				// just echo a warning
				services = append(services, a) // keep as literal, harmless
			}
		} else {
			services = append(services, a)
		}
	}
	base := fmt.Sprintf("docker compose -f %s up -d", shellQuote(compose))
	if len(services) > 0 {
		return []string{base + " " + strings.Join(shellQuoteAll(services), " ")}
	}
	return []string{base}
}

func shellQuote(s string) string {
	// minimal quoting
	if strings.ContainsAny(s, " \t\"'") {
		return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
	}
	return s
}
func shellQuoteAll(ss []string) []string {
	out := make([]string, len(ss))
	for i, s := range ss {
		out[i] = shellQuote(s)
	}
	return out
}
