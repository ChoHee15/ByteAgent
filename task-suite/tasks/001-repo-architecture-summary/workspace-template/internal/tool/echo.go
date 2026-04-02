package tool

type EchoTool struct{}

func NewEchoTool() EchoTool {
	return EchoTool{}
}

func (EchoTool) Run(input string) string {
	return "echo:" + input
}
