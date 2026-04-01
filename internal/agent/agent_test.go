package agent

import (
	"context"
	"io"
	"testing"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func TestNew(t *testing.T) {
	t.Parallel()

	agent, err := New(context.Background(), fakeToolCallingChatModel{}, []tool.BaseTool{fakeTool{}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if agent == nil {
		t.Fatal("New() returned nil agent")
	}
	if got := agent.Name(context.Background()); got != "code_agent" {
		t.Fatalf("agent.Name() = %q, want %q", got, "code_agent")
	}
}

type fakeToolCallingChatModel struct{}

func (fakeToolCallingChatModel) Generate(context.Context, []*schema.Message, ...einomodel.Option) (*schema.Message, error) {
	return schema.AssistantMessage("ok", nil), nil
}

func (fakeToolCallingChatModel) Stream(context.Context, []*schema.Message, ...einomodel.Option) (*schema.StreamReader[*schema.Message], error) {
	return schema.StreamReaderWithConvert(schema.StreamReaderFromArray([]string{"ok"}), func(chunk string) (*schema.Message, error) {
		if chunk == "" {
			return nil, io.EOF
		}
		return schema.AssistantMessage(chunk, nil), nil
	}), nil
}

func (fakeToolCallingChatModel) WithTools([]*schema.ToolInfo) (einomodel.ToolCallingChatModel, error) {
	return fakeToolCallingChatModel{}, nil
}

type fakeTool struct{}

func (fakeTool) Info(context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "bash",
		Desc: "fake bash tool",
	}, nil
}
