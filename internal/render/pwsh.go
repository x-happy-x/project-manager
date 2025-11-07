package render

import (
	"fmt"
	"strings"

	"pm/internal/plan"
)

type pwshRenderer struct{}

func (p pwshRenderer) Name() string { return "pwsh" }

func (p pwshRenderer) Begin(root string) []string {
	return []string{
		"# pm begin",
		fmt.Sprintf("Push-Location %s", pwshQuote(root)),
	}
}
func (p pwshRenderer) End() []string {
	return []string{
		"Pop-Location",
		"# pm end",
	}
}
func (p pwshRenderer) RenderOp(op plan.Op) []string {
	switch v := op.(type) {
	case plan.OpPushd:
		return []string{fmt.Sprintf("Push-Location %s", pwshQuote(v.Dir))}
	case plan.OpPopd:
		return []string{"Pop-Location"}
	case plan.OpEcho:
		return []string{fmt.Sprintf("Write-Host %s", pwshQuote(v.Line))}
	case plan.OpRun:
		return []string{v.Line}
	default:
		return nil
	}
}

func pwshQuote(s string) string {
	// single-quote with escaping
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
