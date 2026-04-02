package runner

import (
	"task/archsummary/internal/config"
	"task/archsummary/internal/tool"
)

type Runner struct {
	cfg  config.Config
	tool tool.EchoTool
}

func New(cfg config.Config, echo tool.EchoTool) Runner {
	return Runner{cfg: cfg, tool: echo}
}

func (r Runner) Handle(request string) string {
	return r.tool.Run(r.cfg.Workspace + ":" + request)
}
