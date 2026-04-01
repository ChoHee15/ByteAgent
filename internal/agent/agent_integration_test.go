//go:build integration

package agent

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/adk"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"code_agent/internal/config"
	openaimodel "code_agent/internal/model"
	localtool "code_agent/internal/tool"
)

func TestIntegrationCodeAgentBashChain(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY is not set")
	}

	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	workspace := t.TempDir()
	targetFile := "integration_target.txt"
	targetContent := "INTEGRATION_SENTINEL_CODE_AGENT"
	if err := os.WriteFile(filepath.Join(workspace, targetFile), []byte(targetContent+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	cfg := &config.Config{
		APIKey:            apiKey,
		BaseURL:           os.Getenv("OPENAI_BASE_URL"),
		Model:             modelName,
		WorkspaceDir:      workspace,
		MaxHistoryTurns:   4,
		MaxCommandBytes:   8 * 1024,
		CommandTimeoutSec: 20,
	}

	chatModel, err := openaimodel.NewOpenAI(ctx, cfg)
	if err != nil {
		t.Fatalf("NewOpenAI() error = %v", err)
	}

	bashTool, err := localtool.NewBashTool(
		cfg.WorkspaceDir,
		time.Duration(cfg.CommandTimeoutSec)*time.Second,
		cfg.MaxCommandBytes,
	)
	if err != nil {
		t.Fatalf("NewBashTool() error = %v", err)
	}

	codeAgent, err := New(ctx, chatModel, []einotool.BaseTool{bashTool})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: codeAgent})
	prompt := "Use the bash tool to read the file " + targetFile + " in the current workspace. " +
		"Then answer in one sentence that includes the exact filename and the exact file content. " +
		"Do not claim you inspected the file unless you actually used the tool."

	iter := runner.Query(ctx, prompt)

	var (
		lastAssistant string
		sawBashTool   bool
		toolPayloads  []string
	)

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			t.Fatalf("runner event error = %v", event.Err)
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}

		msg, _, err := adk.GetMessage(event)
		if err != nil {
			t.Fatalf("GetMessage() error = %v", err)
		}
		if msg == nil {
			continue
		}

		if msg.Role == schema.Tool && event.Output.MessageOutput.ToolName == "bash" {
			sawBashTool = true
			toolPayloads = append(toolPayloads, msg.Content)
		}
		if msg.Role == schema.Assistant && strings.TrimSpace(msg.Content) != "" {
			lastAssistant = msg.Content
		}
	}

	if !sawBashTool {
		t.Fatalf("expected bash tool to be invoked, tool payloads: %v", toolPayloads)
	}
	if strings.TrimSpace(lastAssistant) == "" {
		t.Fatal("expected non-empty assistant response")
	}
	if !strings.Contains(lastAssistant, targetFile) {
		t.Fatalf("assistant response %q does not contain filename %q", lastAssistant, targetFile)
	}
	if !strings.Contains(lastAssistant, targetContent) {
		t.Fatalf("assistant response %q does not contain file content %q", lastAssistant, targetContent)
	}
}
