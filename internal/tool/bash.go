package tool

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// ApprovalRequest describes a command that requires explicit user approval.
type ApprovalRequest struct {
	Command          string
	WorkingDirectory string
}

// ApprovalFunc asks whether the command should be allowed to proceed.
type ApprovalFunc func(context.Context, ApprovalRequest) (bool, error)

type bashToolOptions struct {
	approvalFunc ApprovalFunc
}

// BashToolOption customizes bash tool behavior.
type BashToolOption func(*bashToolOptions)

// ErrWriteNotApproved indicates the user denied a mutating command.
var ErrWriteNotApproved = errors.New("mutating bash command was not approved")

// BashInput is the shell tool input.
type BashInput struct {
	Command          string `json:"command" jsonschema:"description=The bash command to execute,required"`
	WorkingDirectory string `json:"working_directory,omitempty" jsonschema:"description=Optional working directory under the workspace"`
	TimeoutSeconds   int    `json:"timeout_seconds,omitempty" jsonschema:"description=Optional timeout in seconds for this command"`
}

// BashOutput is the shell tool output.
type BashOutput struct {
	Command          string `json:"command"`
	WorkingDirectory string `json:"working_directory"`
	ExitCode         int    `json:"exit_code"`
	Stdout           string `json:"stdout,omitempty"`
	Stderr           string `json:"stderr,omitempty"`
	TimedOut         bool   `json:"timed_out"`
}

// NewBashTool builds a bash execution tool scoped to the workspace.
func NewBashTool(workspaceDir string, defaultTimeout time.Duration, maxOutputBytes int, opts ...BashToolOption) (einotool.InvokableTool, error) {
	options := bashToolOptions{}
	for _, opt := range opts {
		if opt != nil {
			opt(&options)
		}
	}

	return utils.InferTool(
		"bash",
		"Execute a bash command inside the current workspace and return stdout, stderr, and exit code",
		func(ctx context.Context, input *BashInput) (*BashOutput, error) {
			if strings.TrimSpace(input.Command) == "" {
				return nil, fmt.Errorf("command is required")
			}

			workdir, err := resolveWorkdir(workspaceDir, input.WorkingDirectory)
			if err != nil {
				return nil, err
			}

			if commandNeedsApproval(input.Command) {
				if options.approvalFunc == nil {
					return nil, fmt.Errorf("mutating bash command requires interactive approval: %s", input.Command)
				}

				approved, err := options.approvalFunc(ctx, ApprovalRequest{
					Command:          input.Command,
					WorkingDirectory: workdir,
				})
				if err != nil {
					return nil, fmt.Errorf("request bash approval: %w", err)
				}
				if !approved {
					return nil, ErrWriteNotApproved
				}
			}

			timeout := defaultTimeout
			if input.TimeoutSeconds > 0 {
				timeout = time.Duration(input.TimeoutSeconds) * time.Second
			}
			if timeout <= 0 {
				timeout = 30 * time.Second
			}

			cmdCtx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			cmd := exec.CommandContext(cmdCtx, "bash", "-lc", input.Command)
			cmd.Dir = workdir

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &limitedBuffer{buf: &stdout, limit: maxOutputBytes}
			cmd.Stderr = &limitedBuffer{buf: &stderr, limit: maxOutputBytes}

			err = cmd.Run()

			result := &BashOutput{
				Command:          input.Command,
				WorkingDirectory: workdir,
				Stdout:           stdout.String(),
				Stderr:           stderr.String(),
				TimedOut:         cmdCtx.Err() == context.DeadlineExceeded,
			}

			if err == nil {
				return result, nil
			}

			var exitErr *exec.ExitError
			if result.TimedOut {
				result.ExitCode = -1
				return result, nil
			}
			if errors.As(err, &exitErr) {
				result.ExitCode = exitErr.ExitCode()
				return result, nil
			}

			return nil, fmt.Errorf("run bash command: %w", err)
		},
	)
}

// WithWriteApproval requests confirmation before running mutating commands.
func WithWriteApproval(fn ApprovalFunc) BashToolOption {
	return func(options *bashToolOptions) {
		options.approvalFunc = fn
	}
}

func resolveWorkdir(workspaceDir, requested string) (string, error) {
	if strings.TrimSpace(requested) == "" {
		return workspaceDir, nil
	}

	candidate := requested
	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(workspaceDir, requested)
	}

	candidate, err := filepath.Abs(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve working directory: %w", err)
	}

	rel, err := filepath.Rel(workspaceDir, candidate)
	if err != nil {
		return "", fmt.Errorf("check working directory: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("working directory %q is outside workspace %q", candidate, workspaceDir)
	}

	return candidate, nil
}

func commandNeedsApproval(command string) bool {
	normalized := strings.ToLower(strings.TrimSpace(command))
	if normalized == "" {
		return false
	}

	markers := []string{
		" > ", " >\t", " >> ", " >>\t", ">> ", ">>\t", "tee ", " tee",
		"rm ", " rm", "mv ", " mv", "cp ", " cp",
		"touch ", " touch", "mkdir ", " mkdir", "rmdir ", " rmdir",
		"chmod ", " chmod", "chown ", " chown", "ln ", " ln",
		"sed -i", "perl -pi", "truncate ", " truncate", "dd ",
		"git apply", "git commit", "git checkout", "git switch", "git restore",
		"git clean", "git reset", "patch ", " patch",
	}

	for _, marker := range markers {
		if strings.Contains(normalized, marker) {
			return true
		}
	}

	return false
}

type limitedBuffer struct {
	buf   *bytes.Buffer
	limit int
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.limit <= 0 {
		return len(p), nil
	}

	remaining := b.limit - b.buf.Len()
	if remaining <= 0 {
		return len(p), nil
	}

	if len(p) > remaining {
		b.buf.Write(p[:remaining])
		return len(p), nil
	}

	return b.buf.Write(p)
}
