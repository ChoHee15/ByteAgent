package tool

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestResolveWorkdir(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	nested := filepath.Join(workspace, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	testCases := []struct {
		name      string
		requested string
		want      string
		wantErr   string
	}{
		{
			name: "defaults to workspace",
			want: workspace,
		},
		{
			name:      "resolves relative path",
			requested: "nested",
			want:      nested,
		},
		{
			name:      "accepts absolute path in workspace",
			requested: nested,
			want:      nested,
		},
		{
			name:      "rejects escaping workspace",
			requested: "../outside",
			wantErr:   "outside workspace",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := resolveWorkdir(workspace, tc.requested)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("resolveWorkdir() error = %v, want substring %q", err, tc.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("resolveWorkdir() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("resolveWorkdir() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestNewBashTool(t *testing.T) {
	t.Parallel()

	workspace := t.TempDir()
	nested := filepath.Join(workspace, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	testCases := []struct {
		name            string
		timeout         time.Duration
		maxOutputBytes  int
		input           BashInput
		wantExitCode    int
		wantStdout      string
		wantStderr      string
		wantTimedOut    bool
		wantWorkdir     string
		wantInvokeError string
	}{
		{
			name:        "executes command successfully",
			timeout:     time.Second,
			input:       BashInput{Command: "printf hello"},
			wantStdout:  "hello",
			wantWorkdir: workspace,
		},
		{
			name:         "returns non zero exit code",
			timeout:      time.Second,
			input:        BashInput{Command: "printf boom >&2; exit 7"},
			wantExitCode: 7,
			wantStderr:   "boom",
			wantWorkdir:  workspace,
		},
		{
			name:         "times out command",
			timeout:      50 * time.Millisecond,
			input:        BashInput{Command: "sleep 1"},
			wantExitCode: -1,
			wantTimedOut: true,
			wantWorkdir:  workspace,
		},
		{
			name:            "rejects empty command",
			timeout:         time.Second,
			input:           BashInput{},
			wantInvokeError: "command is required",
		},
		{
			name:            "rejects workdir outside workspace",
			timeout:         time.Second,
			input:           BashInput{Command: "pwd", WorkingDirectory: "../outside"},
			wantInvokeError: "outside workspace",
		},
		{
			name:           "truncates stdout",
			timeout:        time.Second,
			maxOutputBytes: 4,
			input:          BashInput{Command: "printf 123456"},
			wantStdout:     "1234",
			wantWorkdir:    workspace,
		},
		{
			name:        "uses nested working directory",
			timeout:     time.Second,
			input:       BashInput{Command: "pwd", WorkingDirectory: "nested"},
			wantStdout:  nested,
			wantWorkdir: nested,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			maxOutputBytes := tc.maxOutputBytes
			if maxOutputBytes == 0 {
				maxOutputBytes = 32 * 1024
			}

			bashTool, err := NewBashTool(workspace, tc.timeout, maxOutputBytes)
			if err != nil {
				t.Fatalf("NewBashTool() error = %v", err)
			}

			arguments, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("Marshal() error = %v", err)
			}

			result, err := bashTool.InvokableRun(context.Background(), string(arguments))
			if tc.wantInvokeError != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantInvokeError) {
					t.Fatalf("InvokableRun() error = %v, want substring %q", err, tc.wantInvokeError)
				}
				return
			}

			if err != nil {
				t.Fatalf("InvokableRun() error = %v", err)
			}

			var output BashOutput
			if err := json.Unmarshal([]byte(result), &output); err != nil {
				t.Fatalf("Unmarshal() error = %v, raw = %s", err, result)
			}

			if output.ExitCode != tc.wantExitCode {
				t.Fatalf("ExitCode = %d, want %d", output.ExitCode, tc.wantExitCode)
			}
			if strings.TrimSpace(output.Stdout) != tc.wantStdout {
				t.Fatalf("Stdout = %q, want %q", output.Stdout, tc.wantStdout)
			}
			if strings.TrimSpace(output.Stderr) != tc.wantStderr {
				t.Fatalf("Stderr = %q, want %q", output.Stderr, tc.wantStderr)
			}
			if output.TimedOut != tc.wantTimedOut {
				t.Fatalf("TimedOut = %v, want %v", output.TimedOut, tc.wantTimedOut)
			}
			if output.WorkingDirectory != tc.wantWorkdir {
				t.Fatalf("WorkingDirectory = %q, want %q", output.WorkingDirectory, tc.wantWorkdir)
			}
		})
	}
}
