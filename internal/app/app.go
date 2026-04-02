package app

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/chzyer/readline"
	"github.com/cloudwego/eino/adk"
	einotool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"code_agent/internal/agent"
	"code_agent/internal/config"
	openaimodel "code_agent/internal/model"
	localtool "code_agent/internal/tool"
)

const defaultMaxIterations = 26

type queryRunner interface {
	Query(ctx context.Context, query string, opts ...adk.AgentRunOption) *adk.AsyncIterator[*adk.AgentEvent]
}

type bootstrapFunc func(context.Context) (*config.Config, queryRunner, error)

type spinnerFactory func(io.Writer) progressIndicator
type lineEditorFactory func() (lineEditor, error)

type progressIndicator interface {
	Start()
	Stop()
}

type lineEditor interface {
	Readline() (string, error)
	SetPrompt(string)
	Close() error
}

// App is the CLI application.
type App struct {
	cfg    *config.Config
	runner queryRunner

	stdin     *os.File
	stdout    io.Writer
	bootstrap bootstrapFunc

	progressFactory   spinnerFactory
	lineEditorFactory lineEditorFactory
	forceProgress     bool
	forceApproval     bool

	progressMu     sync.Mutex
	activeProgress progressIndicator
	activeEditor   lineEditor
}

// New creates the CLI application.
func New() (*App, error) {
	app := &App{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		progressFactory: func(w io.Writer) progressIndicator {
			return newSpinner(w, "agent running")
		},
	}
	app.bootstrap = app.defaultBootstrap

	return app, nil
}

// Run executes the CLI command.
func (a *App) Run(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("code-agent", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	interactive := fs.Bool("i", false, "start interactive mode")
	help := fs.Bool("h", false, "show help")

	if err := fs.Parse(args); err != nil {
		return a.usageError(err)
	}

	if *help {
		a.printUsage()
		return nil
	}

	if err := a.initialize(ctx); err != nil {
		return err
	}

	prompt, err := a.resolvePrompt(fs.Args())
	if err != nil {
		return err
	}

	if *interactive || prompt == "" {
		return a.runInteractive(ctx)
	}

	answer, err := a.ask(ctx, prompt)
	if err != nil {
		return err
	}

	fmt.Fprintln(a.outputWriter(), answer)
	return nil
}

func (a *App) resolvePrompt(args []string) (string, error) {
	if len(args) > 0 {
		return strings.TrimSpace(strings.Join(args, " ")), nil
	}

	info, err := a.inputFile().Stat()
	if err != nil {
		return "", fmt.Errorf("inspect stdin: %w", err)
	}

	if info.Mode()&os.ModeCharDevice != 0 {
		return "", nil
	}

	data, err := io.ReadAll(a.inputFile())
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

func (a *App) initialize(ctx context.Context) error {
	if a.cfg != nil && a.runner != nil {
		return nil
	}

	bootstrap := a.bootstrap
	if bootstrap == nil {
		bootstrap = a.defaultBootstrap
	}

	cfg, runner, err := bootstrap(ctx)
	if err != nil {
		return err
	}

	a.cfg = cfg
	a.runner = runner

	return nil
}

func (a *App) runInteractive(ctx context.Context) error {
	fmt.Fprintf(a.outputWriter(), "workspace: %s\n", a.cfg.WorkspaceDir)
	fmt.Fprintln(a.outputWriter(), "interactive mode. type 'exit' or 'quit' to leave.")

	editor, err := a.newLineEditor()
	if err != nil {
		return err
	}
	if editor != nil {
		defer func() { _ = editor.Close() }()
	}

	history := make([]turn, 0, a.cfg.MaxHistoryTurns)
	a.activeEditor = editor
	defer func() { a.activeEditor = nil }()

	for {
		input, err := a.readInteractiveInput("> ")
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Fprintln(a.outputWriter())
				return nil
			}
			return err
		}
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			return nil
		}

		prompt := withHistory(history, input)
		answer, err := a.ask(ctx, prompt)
		if err != nil {
			if a.handleInteractiveError(err) {
				continue
			}
			return err
		}

		fmt.Fprintln(a.outputWriter(), answer)
		history = append(history, turn{User: input, Assistant: answer})
		if len(history) > a.cfg.MaxHistoryTurns {
			history = history[len(history)-a.cfg.MaxHistoryTurns:]
		}
	}
}

func (a *App) ask(ctx context.Context, prompt string) (string, error) {
	a.startProgress()
	defer a.stopProgress()

	iter := a.runner.Query(ctx, prompt)

	var lastAssistant string

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", a.presentableError(event.Err)
		}
		msg, _, err := adk.GetMessage(event)
		if err != nil {
			return "", err
		}
		if msg == nil {
			continue
		}
		if msg.Role == schema.Assistant && strings.TrimSpace(msg.Content) != "" {
			lastAssistant = msg.Content
		}
	}

	if strings.TrimSpace(lastAssistant) == "" {
		return "", errors.New("agent returned no assistant message")
	}

	return lastAssistant, nil
}

func (a *App) usageError(err error) error {
	a.printUsage()
	return err
}

func (a *App) printUsage() {
	fmt.Fprintln(a.outputWriter(), `Usage:
  code-agent [options] <prompt>
  echo "your prompt" | code-agent
  code-agent -i

Options:
  -i    interactive mode
  -h    show help

Environment:
  OPENAI_API_KEY                     required
  OPENAI_MODEL                       optional, default gpt-4o-mini
  OPENAI_BASE_URL                    optional
  CODE_AGENT_MAX_HISTORY_TURNS       optional, default 8
  CODE_AGENT_MAX_ITERATIONS          optional, default 26
  CODE_AGENT_COMMAND_TIMEOUT_SEC     optional, default 120
  CODE_AGENT_MAX_COMMAND_OUTPUT_BYTES optional, default 32768
  CODE_AGENT_UNSAFE_AUTO_APPROVE_BASH_WRITES optional, default off`)
}

type turn struct {
	User      string
	Assistant string
}

func withHistory(history []turn, input string) string {
	if len(history) == 0 {
		return input
	}

	var b strings.Builder
	b.WriteString("Conversation context:\n")
	for _, item := range history {
		b.WriteString("User: ")
		b.WriteString(item.User)
		b.WriteString("\nAssistant: ")
		b.WriteString(item.Assistant)
		b.WriteString("\n")
	}
	b.WriteString("\nCurrent user request:\n")
	b.WriteString(input)
	return b.String()
}

func (a *App) defaultBootstrap(ctx context.Context) (*config.Config, queryRunner, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	model, err := openaimodel.NewOpenAI(ctx, cfg)
	if err != nil {
		return nil, nil, err
	}

	bashTool, err := localtool.NewBashTool(
		cfg.WorkspaceDir,
		time.Duration(cfg.CommandTimeoutSec)*time.Second,
		cfg.MaxCommandBytes,
		localtool.WithWriteApproval(a.confirmMutatingCommand),
	)
	if err != nil {
		return nil, nil, err
	}

	codeAgent, err := agent.New(ctx, model, []einotool.BaseTool{bashTool}, cfg.MaxIterations)
	if err != nil {
		return nil, nil, err
	}

	return cfg, adk.NewRunner(ctx, adk.RunnerConfig{
		Agent: codeAgent,
	}), nil
}

func (a *App) inputFile() *os.File {
	if a.stdin != nil {
		return a.stdin
	}

	return os.Stdin
}

func (a *App) newLineEditor() (lineEditor, error) {
	if a.lineEditorFactory != nil {
		editor, err := a.lineEditorFactory()
		if err != nil {
			return nil, fmt.Errorf("create line editor: %w", err)
		}

		return editor, nil
	}
	if !a.canUseLineEditor() {
		return nil, nil
	}

	editor, err := a.newDefaultLineEditor()
	if err != nil {
		return nil, fmt.Errorf("create line editor: %w", err)
	}

	return editor, nil
}

func (a *App) newDefaultLineEditor() (lineEditor, error) {
	outputFile, ok := a.outputWriter().(*os.File)
	if !ok {
		return nil, nil
	}

	instance, err := readline.NewEx(&readline.Config{
		Prompt:                 "> ",
		Stdin:                  a.inputFile(),
		Stdout:                 outputFile,
		Stderr:                 outputFile,
		HistoryLimit:           256,
		DisableAutoSaveHistory: true,
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
	})
	if err != nil {
		return nil, err
	}

	return &readlineEditor{instance: instance}, nil
}

func (a *App) outputWriter() io.Writer {
	if a.stdout != nil {
		return a.stdout
	}

	return os.Stdout
}

func (a *App) readInteractiveInput(prompt string) (string, error) {
	if a.activeEditor != nil {
		for {
			a.activeEditor.SetPrompt(prompt)
			line, err := a.activeEditor.Readline()
			if err == nil {
				return strings.TrimSpace(line), nil
			}
			if errors.Is(err, readline.ErrInterrupt) {
				fmt.Fprintln(a.outputWriter())
				continue
			}
			if errors.Is(err, io.EOF) {
				return "", io.EOF
			}

			return "", fmt.Errorf("read input: %w", err)
		}
	}

	fmt.Fprint(a.outputWriter(), prompt)
	reader := bufio.NewReader(a.inputFile())
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("read input: %w", err)
	}
	if errors.Is(err, io.EOF) && strings.TrimSpace(line) == "" {
		return "", io.EOF
	}

	return strings.TrimSpace(line), nil
}

func (a *App) newProgressIndicator() progressIndicator {
	if !a.shouldShowProgress() {
		return nil
	}

	factory := a.progressFactory
	if factory == nil {
		factory = func(w io.Writer) progressIndicator {
			return newSpinner(w, "agent running")
		}
	}

	return factory(a.outputWriter())
}

func (a *App) shouldShowProgress() bool {
	if a.forceProgress {
		return true
	}

	outputFile, ok := a.outputWriter().(*os.File)
	if !ok {
		return false
	}

	info, err := outputFile.Stat()
	if err != nil {
		return false
	}

	return info.Mode()&os.ModeCharDevice != 0
}

func (a *App) canPromptForApproval() bool {
	if a.forceApproval {
		return true
	}

	stdinInfo, err := a.inputFile().Stat()
	if err != nil || stdinInfo.Mode()&os.ModeCharDevice == 0 {
		return false
	}

	stdoutFile, ok := a.outputWriter().(*os.File)
	if !ok {
		return false
	}
	stdoutInfo, err := stdoutFile.Stat()
	if err != nil {
		return false
	}

	return stdoutInfo.Mode()&os.ModeCharDevice != 0
}

func (a *App) canUseLineEditor() bool {
	if a.forceApproval {
		return true
	}

	stdinInfo, err := a.inputFile().Stat()
	if err != nil || stdinInfo.Mode()&os.ModeCharDevice == 0 {
		return false
	}

	stdoutFile, ok := a.outputWriter().(*os.File)
	if !ok {
		return false
	}
	stdoutInfo, err := stdoutFile.Stat()
	if err != nil {
		return false
	}

	return stdoutInfo.Mode()&os.ModeCharDevice != 0
}

func (a *App) confirmMutatingCommand(ctx context.Context, request localtool.ApprovalRequest) (bool, error) {
	_ = ctx

	resume := a.pauseProgress()
	defer a.resumeProgress(resume)

	if a.cfg != nil && a.cfg.UnsafeAutoApproveBash {
		return true, nil
	}

	if !a.canPromptForApproval() {
		return false, fmt.Errorf("mutating bash commands require confirmation from an interactive terminal")
	}

	fmt.Fprintln(a.outputWriter())
	fmt.Fprintln(a.outputWriter(), "Command requires confirmation before modifying files:")
	fmt.Fprintf(a.outputWriter(), "  working directory: %s\n", request.WorkingDirectory)
	fmt.Fprintf(a.outputWriter(), "  command: %s\n", request.Command)
	answer, err := a.readInteractiveInput("Proceed? [y/N]: ")
	if err != nil {
		if errors.Is(err, io.EOF) {
			return false, nil
		}
		return false, fmt.Errorf("read approval input: %w", err)
	}

	switch strings.ToLower(answer) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func (a *App) handleInteractiveError(err error) bool {
	if errors.Is(err, localtool.ErrWriteNotApproved) {
		fmt.Fprintln(a.outputWriter(), "Write was not approved. Current task was canceled.")
		return true
	}

	return false
}

func (a *App) presentableError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, localtool.ErrWriteNotApproved) {
		return err
	}

	if strings.Contains(strings.ToLower(err.Error()), "exceeds max iterations") {
		return fmt.Errorf(
			"task exceeded max iterations (%d); try narrowing the request or increasing CODE_AGENT_MAX_ITERATIONS",
			a.maxIterationsLimit(),
		)
	}

	return err
}

func (a *App) maxIterationsLimit() int {
	if a.cfg != nil && a.cfg.MaxIterations > 0 {
		return a.cfg.MaxIterations
	}

	return defaultMaxIterations
}

func (a *App) startProgress() {
	progress := a.newProgressIndicator()
	if progress == nil {
		return
	}

	a.progressMu.Lock()
	a.activeProgress = progress
	a.progressMu.Unlock()

	progress.Start()
}

func (a *App) stopProgress() {
	a.progressMu.Lock()
	progress := a.activeProgress
	a.activeProgress = nil
	a.progressMu.Unlock()

	if progress != nil {
		progress.Stop()
	}
}

func (a *App) pauseProgress() bool {
	a.progressMu.Lock()
	progress := a.activeProgress
	a.activeProgress = nil
	a.progressMu.Unlock()

	if progress == nil {
		return false
	}

	progress.Stop()
	return true
}

func (a *App) resumeProgress(shouldResume bool) {
	if shouldResume {
		a.startProgress()
	}
}

type spinner struct {
	writer io.Writer
	label  string
	stopCh chan struct{}
	doneCh chan struct{}
	frames []string
	mu     sync.Once
	stopMu sync.Once
}

type readlineEditor struct {
	instance *readline.Instance
}

func (r *readlineEditor) Readline() (string, error) {
	return r.instance.Readline()
}

func (r *readlineEditor) SetPrompt(prompt string) {
	r.instance.SetPrompt(prompt)
}

func (r *readlineEditor) Close() error {
	return r.instance.Close()
}

func newSpinner(writer io.Writer, label string) progressIndicator {
	return &spinner{
		writer: writer,
		label:  label,
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
		frames: []string{"|", "/", "-", `\`},
	}
}

func (s *spinner) Start() {
	s.mu.Do(func() {
		go func() {
			ticker := time.NewTicker(120 * time.Millisecond)
			defer ticker.Stop()
			defer close(s.doneCh)

			frameIndex := 0
			for {
				fmt.Fprintf(s.writer, "\r%s %s", s.frames[frameIndex], s.label)
				frameIndex = (frameIndex + 1) % len(s.frames)

				select {
				case <-ticker.C:
				case <-s.stopCh:
					fmt.Fprintf(s.writer, "\r%s\r", strings.Repeat(" ", len(s.label)+2))
					return
				}
			}
		}()
	})
}

func (s *spinner) Stop() {
	s.stopMu.Do(func() {
		close(s.stopCh)
		<-s.doneCh
	})
}
