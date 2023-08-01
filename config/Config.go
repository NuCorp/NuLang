package config

type Kind string

const (
	Interpreter = Kind("interpreter")
	Compiler    = Kind("compiler")
	Interactive = Kind("interactive")
)

type ToolInfo struct {
	kind Kind
}

func (t ToolInfo) WithKind(kind Kind) ToolInfo {
	t.kind = kind
	return t
}
func (t ToolInfo) Kind() Kind {
	return t.kind
}

const InteractiveFile = ".interactive.temp.nu"
