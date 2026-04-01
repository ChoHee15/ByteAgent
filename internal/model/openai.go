package model

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"

	"code_agent/internal/config"
)

// NewOpenAI creates an Eino OpenAI chat model.
func NewOpenAI(ctx context.Context, cfg *config.Config) (einomodel.ToolCallingChatModel, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  cfg.APIKey,
		Model:   cfg.Model,
		BaseURL: cfg.BaseURL,
	})
	if err != nil {
		return nil, fmt.Errorf("create openai chat model: %w", err)
	}

	return chatModel, nil
}
