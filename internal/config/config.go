package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	defaultModel         = "gpt-4o-mini"
	defaultMaxTurns      = 8
	defaultMaxIterations = 26
	defaultCommandLimit  = 32 * 1024
	defaultTimeoutSec    = 120
)

// Config holds runtime configuration.
type Config struct {
	APIKey                string
	BaseURL               string
	Model                 string
	WorkspaceDir          string
	MaxHistoryTurns       int
	MaxIterations         int
	MaxCommandBytes       int
	CommandTimeoutSec     int
	UnsafeAutoApproveBash bool
}

// Load reads configuration from environment.
func Load() (*Config, error) {
	workspaceDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	workspaceDir, err = filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("resolve working directory: %w", err)
	}

	cfg := &Config{
		APIKey:                os.Getenv("OPENAI_API_KEY"),
		BaseURL:               os.Getenv("OPENAI_BASE_URL"),
		Model:                 envOrDefault("OPENAI_MODEL", defaultModel),
		WorkspaceDir:          workspaceDir,
		MaxHistoryTurns:       envIntOrDefault("CODE_AGENT_MAX_HISTORY_TURNS", defaultMaxTurns),
		MaxIterations:         envIntOrDefault("CODE_AGENT_MAX_ITERATIONS", defaultMaxIterations),
		MaxCommandBytes:       envIntOrDefault("CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES", defaultCommandLimit),
		CommandTimeoutSec:     envIntOrDefault("CODE_AGENT_COMMAND_TIMEOUT_SEC", defaultTimeoutSec),
		UnsafeAutoApproveBash: envBool("CODE_AGENT_UNSAFE_AUTO_APPROVE_BASH_WRITES"),
	}

	if cfg.APIKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY is required")
	}

	if cfg.MaxHistoryTurns < 1 {
		cfg.MaxHistoryTurns = defaultMaxTurns
	}
	if cfg.MaxIterations < 1 {
		cfg.MaxIterations = defaultMaxIterations
	}
	if cfg.MaxCommandBytes < 1024 {
		cfg.MaxCommandBytes = defaultCommandLimit
	}
	if cfg.CommandTimeoutSec < 1 {
		cfg.CommandTimeoutSec = defaultTimeoutSec
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envIntOrDefault(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return n
}

func envBool(key string) bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch value {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}
