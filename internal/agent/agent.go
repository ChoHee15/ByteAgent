package agent

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

const instructionTemplate = `You are a lightweight coding assistant running in a CLI.

Your job:
- help the user analyze code, propose changes, and use the bash tool when shell access is needed
- prefer concise, practical answers
- when using bash, keep commands focused on the current workspace

Safety rules:
- do not run destructive commands unless the user explicitly asks
- avoid commands that modify files outside the current workspace
- explain the result clearly after using tools

When bash output is relevant, summarize it instead of dumping unnecessary noise.`

// New creates the code agent.
func New(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (adk.Agent, error) {
	agent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "code_agent",
		Description: "A lightweight code agent that can answer questions and execute bash commands in the workspace",
		Instruction: instructionTemplate,
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		MaxIterations: 12,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model agent: %w", err)
	}

	return agent, nil
}
