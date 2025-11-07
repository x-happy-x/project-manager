package render

import (
	"fmt"
	"strings"

	"pm/internal/plan"
)

type bashRenderer struct{}

func (b bashRenderer) Name() string { return "bash" }

func (b bashRenderer) Begin(root string) []string {
	return []string{
		fmt.Sprintf("# pm begin"),
		fmt.Sprintf("pushd %s >/dev/null", sh(root)),
	}
}
func (b bashRenderer) End() []string {
	return []string{
		"popd >/dev/null",
		"# pm end",
	}
}
func (b bashRenderer) RenderOp(op plan.Op) []string {
	switch v := op.(type) {
	case plan.OpPushd:
		return []string{fmt.Sprintf("pushd %s >/dev/null", sh(v.Dir))}
	case plan.OpPopd:
		return []string{"popd >/dev/null"}
	case plan.OpEcho:
		return []string{fmt.Sprintf("echo %s", sh(v.Line))}
	case plan.OpRun:
		return []string{v.Line}
	default:
		return nil
	}
}

func sh(s string) string {
	if strings.ContainsAny(s, " \t\"'") {
		return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'"
	}
	return s
}
