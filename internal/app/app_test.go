package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"

	"code_agent/internal/config"
	localtool "code_agent/internal/tool"
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
			name: "translates max iteration error",
			runner: &fakeRunner{
				events: [][]*adk.AgentEvent{
					{
						{Err: errors.New("pre processor fail: exceeds max iterations")},
					},
				},
			},
			wantErr:     "task exceeded max iterations (21); try narrowing the request or increasing CODE_AGENT_MAX_ITERATIONS",
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

			application := &App{
				cfg: &config.Config{
					MaxIterations: 21,
				},
				runner: tc.runner,
				stdout: &bytes.Buffer{},
			}

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
		if strings.Contains(output.String(), "agent running") {
			t.Fatalf("Run() output = %q, did not expect spinner text for non-terminal output", output.String())
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

func TestAskShowsProgressForTerminalOutput(t *testing.T) {
	t.Parallel()

	progress := &fakeProgressIndicator{}
	application := &App{
		runner: &fakeRunner{
			events: [][]*adk.AgentEvent{
				{
					adk.EventFromMessage(schema.AssistantMessage("done", nil), nil, schema.Assistant, ""),
				},
			},
		},
		stdout: &bytes.Buffer{},
		progressFactory: func(io.Writer) progressIndicator {
			return progress
		},
		forceProgress: true,
	}

	answer, err := application.ask(context.Background(), "inspect")
	if err != nil {
		t.Fatalf("ask() error = %v", err)
	}
	if answer != "done" {
		t.Fatalf("ask() = %q, want %q", answer, "done")
	}
	if progress.startCount != 1 {
		t.Fatalf("progress start count = %d, want 1", progress.startCount)
	}
	if progress.stopCount != 1 {
		t.Fatalf("progress stop count = %d, want 1", progress.stopCount)
	}
}

func TestConfirmMutatingCommand(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		input       string
		forcePrompt bool
		want        bool
		wantErr     string
	}{
		{
			name:        "approves yes",
			input:       "yes\n",
			forcePrompt: true,
			want:        true,
		},
		{
			name:        "defaults to deny",
			input:       "n\n",
			forcePrompt: true,
			want:        false,
		},
		{
			name:    "requires interactive terminal",
			input:   "yes\n",
			wantErr: "interactive terminal",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var output bytes.Buffer
			application := &App{
				stdin:         newInputFile(t, tc.input),
				stdout:        &output,
				forceApproval: tc.forcePrompt,
			}

			approved, err := application.confirmMutatingCommand(context.Background(), localtool.ApprovalRequest{
				Command:          "touch demo.txt",
				WorkingDirectory: "/tmp/workspace",
			})
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("confirmMutatingCommand() error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("confirmMutatingCommand() error = %v", err)
			}
			if approved != tc.want {
				t.Fatalf("confirmMutatingCommand() = %v, want %v", approved, tc.want)
			}
			if !strings.Contains(output.String(), "Proceed? [y/N]: ") {
				t.Fatalf("confirmMutatingCommand() output = %q, want approval prompt", output.String())
			}
		})
	}
}

func TestHandleInteractiveError(t *testing.T) {
	t.Parallel()

	var output bytes.Buffer
	application := &App{stdout: &output}

	handled := application.handleInteractiveError(errors.Join(errors.New("tool failed"), localtool.ErrWriteNotApproved))
	if !handled {
		t.Fatal("handleInteractiveError() = false, want true")
	}
	if !strings.Contains(output.String(), "Write was not approved. Current task was canceled.") {
		t.Fatalf("handleInteractiveError() output = %q, want rejection notice", output.String())
	}

	output.Reset()
	handled = application.handleInteractiveError(errors.New("other error"))
	if handled {
		t.Fatal("handleInteractiveError() = true, want false")
	}
	if output.Len() != 0 {
		t.Fatalf("handleInteractiveError() wrote unexpected output: %q", output.String())
	}
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

type fakeProgressIndicator struct {
	mu         sync.Mutex
	startCount int
	stopCount  int
}

func (f *fakeProgressIndicator) Start() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.startCount++
}

func (f *fakeProgressIndicator) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stopCount++
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
