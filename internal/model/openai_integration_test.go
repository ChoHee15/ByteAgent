//go:build integration

package model

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/eino/schema"

	"code_agent/internal/config"
)

func TestIntegrationOpenAISmoke(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY is not set")
	}

	modelName := os.Getenv("OPENAI_MODEL")
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	chatModel, err := NewOpenAI(ctx, &config.Config{
		APIKey:  apiKey,
		BaseURL: os.Getenv("OPENAI_BASE_URL"),
		Model:   modelName,
	})
	if err != nil {
		t.Fatalf("NewOpenAI() error = %v", err)
	}

	msg, err := chatModel.Generate(ctx, []*schema.Message{
		schema.UserMessage("Reply with a short acknowledgment."),
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if msg == nil {
		t.Fatal("Generate() returned nil message")
	}
	if strings.TrimSpace(msg.Content) == "" {
		t.Fatalf("Generate() content = %q, want non-empty response", msg.Content)
	}
}
