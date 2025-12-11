package templ

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"pm/internal/config"
)

var (
	envRe   = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)
	paramRe = regexp.MustCompile(`@\{([A-Za-z0-9_.-]+)\}`)
	cfgRe   = regexp.MustCompile(`#\{([^}]+)\}`)
	funcRe  = regexp.MustCompile(`_\{([A-Za-z0-9_.-]+)\((.*?)\)\}`)
)

func RenderString(text string, params map[string]string, proj *config.ProjectMeta, global *config.GlobalConfig, ctx []map[string]any) (string, error) {
	// ${ENV}
	text = envRe.ReplaceAllStringFunc(text, func(s string) string {
		m := envRe.FindStringSubmatch(s)
		if len(m) == 2 {
			return getenv(m[1])
		}
		return ""
	})
	// @{param}
	text = paramRe.ReplaceAllStringFunc(text, func(s string) string {
		m := paramRe.FindStringSubmatch(s)
		if len(m) == 2 {
			if v, ok := params[m[1]]; ok {
				return v
			}
		}
		return ""
	})
	// #{cfg.path}
	text = cfgRe.ReplaceAllStringFunc(text, func(s string) string {
		m := cfgRe.FindStringSubmatch(s)
		if len(m) == 2 {
			val, err := cfgGetPath(proj, global, m[1], ctx)
			if err != nil {
				return ""
			}
			switch v := val.(type) {
			case string:
				return v
			default:
				if b, err := json.Marshal(v); err == nil {
					return string(b)
				}
				return ""
			}
		}
		return ""
	})
	// _{func(...)}
	out := funcRe.ReplaceAllStringFunc(text, func(s string) string {
		m := funcRe.FindStringSubmatch(s)
		if len(m) != 3 {
			return ""
		}
		full := m[1]
		argstr := m[2]
		srcGlobal := false
		if strings.HasPrefix(full, "global.") {
			full = strings.TrimPrefix(full, "global.")
			srcGlobal = true
		}
		var node *config.FuncDef
		if srcGlobal {
			if global == nil || global.Func == nil {
				return ""
			}
			if f, ok := global.Func[full]; ok {
				node = &f
			}
		} else {
			if proj.Func != nil {
				if f, ok := proj.Func[full]; ok {
					node = &f
				}
			}
		}
		if node == nil {
			return ""
		}
		// defaults (only for non-required params with defaults)
		nodeParams := map[string]string{}
		if node.Params != nil {
			for k, meta := range node.Params {
				if meta.Default != "" {
					nodeParams[k] = meta.Default
				}
			}
		}
		// overrides
		for k, v := range parseKwargs(argstr) {
			nodeParams[k] = v
		}
		for k, meta := range node.Params {
			if meta.Required {
				if _, ok := nodeParams[k]; !ok {
					return "" // or error?
				}
			}
		}
		lines := funcToLines(node.Script)
		var rendered []string
		for _, raw := range lines {
			r, err := RenderString(raw, nodeParams, proj, global, append(ctx, map[string]any{
				"func": node,
			}))
			if err != nil {
				continue
			}
			rendered = append(rendered, r)
		}
		return strings.Join(rendered, " && ")
	})
	return out, nil
}

func funcToLines(s any) []string {
	switch v := s.(type) {
	case string:
		return []string{v}
	case []any:
		out := make([]string, 0, len(v))
		for _, it := range v {
			if str, ok := it.(string); ok {
				out = append(out, str)
			}
		}
		return out
	case []string:
		return v
	default:
		return nil
	}
}

func parseKwargs(s string) map[string]string {
	out := map[string]string{}
	for _, tok := range splitByCommaRespectingQuotes(s) {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		kv := strings.SplitN(tok, "=", 2)
		if len(kv) == 1 {
			out[strings.TrimSpace(kv[0])] = "true"
			continue
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		if unq, ok := unquote(val); ok {
			val = unq
		}
		out[key] = val
	}
	return out
}

func splitByCommaRespectingQuotes(s string) []string {
	var res []string
	var b strings.Builder
	inSingle, inDouble := false, false
	escape := false

	for _, r := range s {
		switch {
		case escape:
			b.WriteRune(r)
			escape = false
		case r == '\\' && !inSingle:
			escape = true
		case r == '\'' && !inDouble:
			inSingle = !inSingle
			b.WriteRune(r) // оставляем кавычку; потом unquote снимет
		case r == '"' && !inSingle:
			inDouble = !inDouble
			b.WriteRune(r)
		case r == ',' && !inSingle && !inDouble:
			res = append(res, b.String())
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	res = append(res, b.String())
	return res
}

func unquote(s string) (string, bool) {
	if len(s) >= 2 && s[0] == '\'' && s[len(s)-1] == '\'' {
		// PowerShell-стиль '' внутри одинарных кавычек -> '
		return strings.ReplaceAll(s[1:len(s)-1], "''", "'"), true
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		u := s[1 : len(s)-1]
		u = strings.ReplaceAll(u, `\"`, `"`)
		u = strings.ReplaceAll(u, `\\`, `\`)
		return u, true
	}
	return s, false
}

func cfgGetPath(proj *config.ProjectMeta, global *config.GlobalConfig, path string, ctx []map[string]any) (any, error) {
	if strings.HasPrefix(path, "global.") && global != nil {
		return dig(global.Raw, strings.TrimPrefix(path, "global."), ctx)
	}
	return dig(structToMap(proj), path, ctx)
}

func dig(root any, path string, ctx []map[string]any) (any, error) {
	// relative . .. ... lookups
	if strings.HasPrefix(path, ".") {
		dots := 0
		for dots < len(path) && path[dots] == '.' {
			dots++
		}
		key := path[dots:]
		idx := len(ctx) - dots
		if idx < 0 {
			return nil, errors.New("relative path out of range")
		}
		if idx >= len(ctx) {
			return nil, errors.New("ctx index")
		}
		return dig(ctx[idx], key, ctx)
	}
	cur := root
	if path == "" {
		return cur, nil
	}
	parts := strings.Split(path, ".")
	for _, p := range parts {
		switch vv := cur.(type) {
		case map[string]any:
			if nx, ok := vv[p]; ok {
				cur = nx
			} else {
				return nil, fmt.Errorf("path not found: %s", path)
			}
		default:
			return nil, fmt.Errorf("path not found: %s", path)
		}
	}
	return cur, nil
}

func structToMap(p *config.ProjectMeta) map[string]any {
	// minimal projection used by #{info.*}, #{commands.*} etc.
	raw := map[string]any{
		"info": map[string]any{
			"name":        p.Info.Name,
			"description": p.Info.Description,
			"root":        p.Info.Root,
		},
	}
	// allow referencing whole YAML via re-marshal? out of scope for now.
	return raw
}

func getenv(k string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return ""
}
