package app

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"code_agent/internal/config"
)

func TestResolvePrompt(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		args    []string
		stdin   string
		want    string
		wantErr string
	}{
		{
			name:  "uses args before stdin",
			args:  []string{"scan", "workspace"},
			stdin: "ignored stdin",
			want:  "scan workspace",
		},
		{
			name:  "reads stdin when args empty",
			stdin: "  prompt from stdin  \n",
			want:  "prompt from stdin",
		},
		{
			name: "empty stdin returns empty prompt",
			want: "",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			input := newInputFile(t, tc.stdin)
			application := &App{stdin: input, stdout: &bytes.Buffer{}}

			got, err := application.resolvePrompt(tc.args)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("resolvePrompt() error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolvePrompt() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("resolvePrompt() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestAsk(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		runner      queryRunner
		want        string
		wantErr     string
		wantQueries []string
	}{
		{
			name: "returns assistant content",
			runner: &fakeRunner{
				events: [][]*adk.AgentEvent{
					{
						adk.EventFromMessage(schema.UserMessage("ignored"), nil, schema.User, ""),
						adk.EventFromMessage(schema.AssistantMessage("answer", nil), nil, schema.Assistant, ""),
					},
				},
			},
			want:        "answer",
			wantQueries: []string{"inspect"},
		},
		{
			name: "returns event error",
			runner: &fakeRunner{
				events: [][]*adk.AgentEvent{
					{
						{Err: errors.New("runner failed")},
					},
				},
			},
			wantErr:     "runner failed",
			wantQueries: []string{"inspect"},
		},
		{
			name: "errors on empty assistant message",
			runner: &fakeRunner{
				events: [][]*adk.AgentEvent{
					{
						adk.EventFromMessage(schema.AssistantMessage("", nil), nil, schema.Assistant, ""),
					},
				},
			},
			wantErr:     "agent returned no assistant message",
			wantQueries: []string{"inspect"},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			application := &App{runner: tc.runner, stdout: &bytes.Buffer{}}

			got, err := application.ask(context.Background(), "inspect")
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("ask() error = %v, want substring %q", err, tc.wantErr)
				}
			} else {
				if err != nil {
					t.Fatalf("ask() error = %v", err)
				}
				if got != tc.want {
					t.Fatalf("ask() = %q, want %q", got, tc.want)
				}
			}

			if runner, ok := tc.runner.(*fakeRunner); ok {
				if len(runner.queries) != len(tc.wantQueries) {
					t.Fatalf("runner queries = %v, want %v", runner.queries, tc.wantQueries)
				}
				for i := range runner.queries {
					if runner.queries[i] != tc.wantQueries[i] {
						t.Fatalf("runner query[%d] = %q, want %q", i, runner.queries[i], tc.wantQueries[i])
					}
				}
			}
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	t.Run("prints help without initialization", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		application := &App{
			stdin:  newInputFile(t, ""),
			stdout: &output,
		}

		if err := application.Run(context.Background(), []string{"-h"}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if !strings.Contains(output.String(), "Usage:") {
			t.Fatalf("Run() output = %q, want usage text", output.String())
		}
	})

	t.Run("uses bootstrap and writes assistant reply", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		runner := &fakeRunner{
			events: [][]*adk.AgentEvent{
				{
					adk.EventFromMessage(schema.AssistantMessage("done", nil), nil, schema.Assistant, ""),
				},
			},
		}

		application := &App{
			stdin:  newInputFile(t, ""),
			stdout: &output,
			bootstrap: func(context.Context) (*config.Config, queryRunner, error) {
				return &config.Config{
					WorkspaceDir:    "/tmp/workspace",
					MaxHistoryTurns: 4,
				}, runner, nil
			},
		}

		if err := application.Run(context.Background(), []string{"analyze", "repo"}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		if got := strings.TrimSpace(output.String()); got != "done" {
			t.Fatalf("Run() output = %q, want %q", got, "done")
		}
		if len(runner.queries) != 1 || runner.queries[0] != "analyze repo" {
			t.Fatalf("runner queries = %v, want single query %q", runner.queries, "analyze repo")
		}
	})

	t.Run("interactive mode exits cleanly", func(t *testing.T) {
		t.Parallel()

		var output bytes.Buffer
		application := &App{
			cfg: &config.Config{
				WorkspaceDir:    "/tmp/workspace",
				MaxHistoryTurns: 4,
			},
			runner: &fakeRunner{},
			stdin:  newInputFile(t, "exit\n"),
			stdout: &output,
		}

		if err := application.Run(context.Background(), []string{"-i"}); err != nil {
			t.Fatalf("Run() error = %v", err)
		}

		rendered := output.String()
		if !strings.Contains(rendered, "interactive mode") {
			t.Fatalf("Run() output = %q, want interactive banner", rendered)
		}
		if !strings.Contains(rendered, "> ") {
			t.Fatalf("Run() output = %q, want prompt marker", rendered)
		}
	})
}

func TestWithHistory(t *testing.T) {
	t.Parallel()

	history := []turn{
		{User: "first", Assistant: "reply one"},
		{User: "second", Assistant: "reply two"},
	}

	got := withHistory(history, "current")
	wantParts := []string{
		"Conversation context:",
		"User: first",
		"Assistant: reply one",
		"User: second",
		"Assistant: reply two",
		"Current user request:",
		"current",
	}

	for _, part := range wantParts {
		if !strings.Contains(got, part) {
			t.Fatalf("withHistory() = %q, missing %q", got, part)
		}
	}
}

type fakeRunner struct {
	queries []string
	events  [][]*adk.AgentEvent
}

func (f *fakeRunner) Query(_ context.Context, query string, _ ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent] {
	f.queries = append(f.queries, query)

	iterator, generator := adk.NewAsyncIteratorPair[*adk.AgentEvent]()
	var events []*adk.AgentEvent
	if len(f.events) > 0 {
		events = f.events[0]
		f.events = f.events[1:]
	}
	for _, event := range events {
		generator.Send(event)
	}
	generator.Close()

	return iterator
}

func newInputFile(t *testing.T, content string) *os.File {
	t.Helper()

	path := filepath.Join(t.TempDir(), "stdin.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}

	t.Cleanup(func() {
		_ = file.Close()
	})

	return file
}
