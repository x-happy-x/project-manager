package dsl

type Chunk struct {
	Name string
	Args []string
}

// [:build, -D, foo, :run, -d] -> [("build", ["-D","foo"]), ("run", ["-d"])]
func SplitColonCommands(args []string) []Chunk {
	var out []Chunk
	i := 0
	for i < len(args) {
		tok := args[i]
		if len(tok) > 0 && tok[0] == ':' {
			name := tok[1:]
			j := i + 1
			var buf []string
			for j < len(args) && !(len(args[j]) > 0 && args[j][0] == ':') {
				buf = append(buf, args[j])
				j++
			}
			out = append(out, Chunk{Name: name, Args: buf})
			i = j
		} else {
			// RAW mode
			return []Chunk{{Name: "__RAW__", Args: args[i:]}}
		}
	}
	return out
}
