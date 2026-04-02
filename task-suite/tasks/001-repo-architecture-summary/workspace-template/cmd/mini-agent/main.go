package main

import (
	"fmt"

	"task/archsummary/internal/config"
	"task/archsummary/internal/runner"
	"task/archsummary/internal/tool"
)

func main() {
	cfg := config.Load()
	engine := runner.New(cfg, tool.NewEchoTool())
	fmt.Println(engine.Handle("demo request"))
}
