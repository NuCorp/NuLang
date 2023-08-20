package config

type ToolMode string

const (
	ModeInterpreter = ToolMode("interpreter")
	ModeCompiler    = ToolMode("compiler")
	ModeInteractive = ToolMode("interactive")
)

type ToolInfo struct {
	kind               ToolMode
	interactiveElement chan any
}

func (t ToolInfo) Mode() ToolMode {
	return t.kind
}

func Interactive() ToolInfo {
	return ToolInfo{
		kind:               ModeInteractive,
		interactiveElement: make(chan any),
	}
}

const InteractiveFile = ".interactive.temp.nu"
