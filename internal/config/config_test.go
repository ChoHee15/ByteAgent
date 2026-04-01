package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	testCases := []struct {
		name           string
		env            map[string]string
		unset          []string
		wantErr        string
		wantModel      string
		wantBaseURL    string
		wantHistory    int
		wantIterations int
		wantCommandMax int
		wantTimeoutSec int
	}{
		{
			name:    "requires api key",
			unset:   []string{"OPENAI_API_KEY"},
			wantErr: "OPENAI_API_KEY is required",
		},
		{
			name: "uses defaults",
			env: map[string]string{
				"OPENAI_API_KEY": "test-key",
			},
			wantModel:      defaultModel,
			wantHistory:    defaultMaxTurns,
			wantIterations: defaultMaxIterations,
			wantCommandMax: defaultCommandLimit,
			wantTimeoutSec: defaultTimeoutSec,
		},
		{
			name: "uses explicit values and falls back on invalid ints",
			env: map[string]string{
				"OPENAI_API_KEY":                      "test-key",
				"OPENAI_MODEL":                        "deepseek-chat",
				"OPENAI_BASE_URL":                     "https://api.example.com",
				"CODE_AGENT_MAX_HISTORY_TURNS":        "bad",
				"CODE_AGENT_MAX_ITERATIONS":           "0",
				"CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES": "0",
				"CODE_AGENT_COMMAND_TIMEOUT_SEC":      "-1",
			},
			wantModel:      "deepseek-chat",
			wantBaseURL:    "https://api.example.com",
			wantHistory:    defaultMaxTurns,
			wantIterations: defaultMaxIterations,
			wantCommandMax: defaultCommandLimit,
			wantTimeoutSec: defaultTimeoutSec,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, key := range []string{
				"OPENAI_API_KEY",
				"OPENAI_MODEL",
				"OPENAI_BASE_URL",
				"CODE_AGENT_MAX_HISTORY_TURNS",
				"CODE_AGENT_MAX_ITERATIONS",
				"CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES",
				"CODE_AGENT_COMMAND_TIMEOUT_SEC",
			} {
				t.Setenv(key, "")
			}
			for _, key := range tc.unset {
				t.Setenv(key, "")
			}
			for key, value := range tc.env {
				t.Setenv(key, value)
			}

			cfg, err := Load()
			if tc.wantErr != "" {
				if err == nil || err.Error() != tc.wantErr {
					t.Fatalf("Load() error = %v, want %q", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}

			cwd, err := os.Getwd()
			if err != nil {
				t.Fatalf("Getwd() error = %v", err)
			}
			wantWorkspace, err := filepath.Abs(cwd)
			if err != nil {
				t.Fatalf("Abs() error = %v", err)
			}

			if cfg.WorkspaceDir != wantWorkspace {
				t.Fatalf("WorkspaceDir = %q, want %q", cfg.WorkspaceDir, wantWorkspace)
			}
			if cfg.Model != tc.wantModel {
				t.Fatalf("Model = %q, want %q", cfg.Model, tc.wantModel)
			}
			if cfg.BaseURL != tc.wantBaseURL {
				t.Fatalf("BaseURL = %q, want %q", cfg.BaseURL, tc.wantBaseURL)
			}
			if cfg.MaxHistoryTurns != tc.wantHistory {
				t.Fatalf("MaxHistoryTurns = %d, want %d", cfg.MaxHistoryTurns, tc.wantHistory)
			}
			if cfg.MaxIterations != tc.wantIterations {
				t.Fatalf("MaxIterations = %d, want %d", cfg.MaxIterations, tc.wantIterations)
			}
			if cfg.MaxCommandBytes != tc.wantCommandMax {
				t.Fatalf("MaxCommandBytes = %d, want %d", cfg.MaxCommandBytes, tc.wantCommandMax)
			}
			if cfg.CommandTimeoutSec != tc.wantTimeoutSec {
				t.Fatalf("CommandTimeoutSec = %d, want %d", cfg.CommandTimeoutSec, tc.wantTimeoutSec)
			}
		})
	}
}
