package domain

import (
	"context"
	"fmt"
	"math"
	"os"
	"strings"

	"github.com/NemCaBong/go-isolate"

	"github.com/NemCaBong/executify/internal/config"
)

type CodeRunner struct {
	submission           *SubmissionWithDetails
	userInput            *string // non-nil = run mode (user-provided input lines); nil = submit mode (uses problem dir)
	userExpectedOutput   *string
	exec                 *isolate.Executor
	codeRunnerCfg        config.CodeRunnerConfig
	stdoutFileName       string
	stderrFileName       string
	actualOutputFileName string
	metaFileName         string
	notifyStatus         StatusNotifier
}

func NewCodeRunner(
	submission *SubmissionWithDetails,
	userInput *string,
	userExpectedOutput *string,
	codeRunnerCfg config.CodeRunnerConfig,
) *CodeRunner {
	hash := hashFileName(submission.ID, codeRunnerCfg.FileSecret)
	return &CodeRunner{
		submission:           submission,
		userInput:            userInput,
		userExpectedOutput:   userExpectedOutput,
		codeRunnerCfg:        codeRunnerCfg,
		stdoutFileName:       hash + StdoutFileName,
		stderrFileName:       hash + StderrFileName,
		actualOutputFileName: hash + ActualOutputFileName,
		metaFileName:         hash + MetaFileName,
	}
}

// WithStatusNotifier wires a callback that is invoked at each lifecycle
// transition (COMPILING, RUNNING). Optional — when unset, only the terminal
// verdict is reflected on the submission.
func (r *CodeRunner) WithStatusNotifier(fn StatusNotifier) *CodeRunner {
	r.notifyStatus = fn
	return r
}

// emitStatus updates the in-memory submission status and fires the notifier
// (if configured). Notifier errors are intentionally swallowed: persistence
// failure on an intermediate state must not abort the run.
func (r *CodeRunner) emitStatus(ctx context.Context, s SubmissionStatus) {
	r.submission.Submission.Status = s
	if r.notifyStatus != nil {
		_ = r.notifyStatus(ctx, s)
	}
}

func (r *CodeRunner) init(ctx context.Context) error {
	boxId := r.submission.ID % r.codeRunnerCfg.BoxModulus
	r.exec = isolate.New().
		BoxID(boxId).
		FullEnv().
		DirSimple("/usr").
		DirSimple("/etc").
		DirSimple("/lib").
		DirSimple("/var").
		DirSimple("/tmp").
		Meta(r.metaFileName).
		Stdout(r.stdoutFileName).
		Stderr(r.stderrFileName).
		InheritFDs(). // to route user res to another fd
		Exec()
	if _, err := r.exec.Init(ctx); err != nil {
		r.exec.Cleanup(ctx)
		return err
	}

	// create sandbox files
	for _, name := range []string{
		r.stdoutFileName,
		r.stderrFileName,
		r.actualOutputFileName,
		r.metaFileName,
	} {
		err := r.exec.WriteToSandbox(name, []byte(""), 0644)
		if err != nil {
			r.exec.Cleanup(ctx)
			return err
		}
	}

	if err := r.exec.WriteToSandbox(
		r.submission.Language.SourceFile,
		[]byte(r.submission.SourceCode),
		0644,
	); err != nil {
		r.exec.Cleanup(ctx)
		return fmt.Errorf("failed to write source code: %w", err)
	}

	return nil
}

// compile runs the compiler inside the sandbox. Returns (compiled, err):
//   - compiled=true, err=nil  → source compiled successfully (or no compile step)
//   - compiled=false, err=nil → compiler ran but rejected the source (CE verdict);
//     the user's stderr has been recorded on the submission
//   - err != nil              → the compile invocation itself blew up (internal error)
func (r *CodeRunner) compile(ctx context.Context) (bool, error) {
	if r.submission.Language.CompileCmd == nil {
		return true, nil
	}

	r.exec.ApplyOptions(
		isolate.WithMemoryLimit(512*1024),
		isolate.WithTimeLimit(10.0),
		isolate.WithWallTimeLimit(20.0),
		isolate.WithProcesses(10),
	)

	result, err := r.exec.Run(ctx, "/bin/bash", "-c", *r.submission.Language.CompileCmd)
	if err != nil {
		return false, fmt.Errorf("compile command failed: %w", err)
	}
	if result.ExitCode != 0 {
		r.submission.Submission.Stderr = result.Stderr
		return false, nil
	}
	return true, nil
}

// applyExecOptions configures resource limits and I/O redirection for a single execution.
func (r *CodeRunner) applyExecOptions() {
	prob := &r.submission.Problem
	memLimitKB := prob.MemoryLimit * 1024
	cpuTime := float64(prob.TimeLimit)
	wallTime := cpuTime * 2.0
	processes := 1

	if prob.CPUTimeLimit != nil {
		cpuTime = *prob.CPUTimeLimit
	}
	if prob.WallTimeLimit != nil {
		wallTime = *prob.WallTimeLimit
	}
	if prob.MaxProcessesAndOrThreads != nil {
		processes = *prob.MaxProcessesAndOrThreads
	}

	opts := []isolate.ExecuteOption{
		isolate.WithMemoryLimit(memLimitKB),
		isolate.WithTimeLimit(cpuTime),
		isolate.WithWallTimeLimit(wallTime),
		isolate.WithProcesses(processes),
	}
	if prob.CPUExtraTime != nil {
		opts = append(opts, isolate.WithExtraTime(*prob.CPUExtraTime))
	}
	if prob.StackLimit != nil {
		opts = append(opts, isolate.WithStackLimit(*prob.StackLimit))
	}
	r.exec.ApplyOptions(opts...)
}

// runOnce feeds stdinContent to the program and captures fd3 as the actual output.
func (r *CodeRunner) runOnce(ctx context.Context, stdinContent string) (*isolate.Result, error) {
	r.applyExecOptions()
	cmd := fmt.Sprintf("%s <<< %q 3>%q", r.submission.Language.RunCmd, stdinContent, r.actualOutputFileName)
	result, err := r.exec.Run(ctx, "/bin/bash", "-c", cmd)
	if err != nil {
		return nil, fmt.Errorf("run command failed: %w", err)
	}
	return result, nil
}

// applyMeta updates resource-usage metrics (time, memory, exit code) from the result.
// Called after each runOnce; tracks the maximum time and memory seen across all runs.
func (r *CodeRunner) applyMeta(result *isolate.Result) {
	sub := &r.submission.Submission
	if result.Meta != nil {
		exitCode := result.Meta.ExitCode
		sub.ExitCode = &exitCode

		if result.Meta.ExitSignal != 0 {
			sig := result.Meta.ExitSignal
			sub.ExitSignal = &sig
		}

		timeMs := int(math.Round(result.Meta.Time * 1000))
		if sub.TimeUsed == nil || timeMs > *sub.TimeUsed {
			sub.TimeUsed = &timeMs
		}

		memKB := result.Meta.MaxRSS
		if sub.MemoryUsed == nil || memKB > *sub.MemoryUsed {
			sub.MemoryUsed = &memKB
		}
	} else {
		exitCode := result.ExitCode
		sub.ExitCode = &exitCode
	}
}

func resultSucceeded(result *isolate.Result) bool {
	if result.Meta != nil {
		return result.Meta.IsSuccess()
	}
	return result.ExitCode == 0 && result.Stderr == ""
}

// executeRun runs every test case line from the user-provided input without early exit.
// Each line of input is treated as one test case; outputs are stored newline-separated
// in ActualOutput so the caller can see all results at once.
func (r *CodeRunner) executeRun(ctx context.Context) error {
	inputLines := splitLines(*r.userInput)
	expectedLines := splitLines(*r.userExpectedOutput)
	if len(inputLines) != len(expectedLines) {
		return fmt.Errorf("expected %d lines of input, got %d", len(expectedLines), len(inputLines))
	}

	sub := &r.submission.Submission
	actualOutputs := make([]string, 0, len(inputLines))
	userOutputs := make([]string, 0, len(inputLines))
	verdict := StatusAccepted

	for i, line := range inputLines {
		result, err := r.runOnce(ctx, line)
		if err != nil {
			return err
		}
		r.applyMeta(result)

		userOutput := strings.TrimSpace(result.Stdout)
		userOutputs = append(userOutputs, userOutput)
		actual, err := r.getActualOutput()
		if err != nil {
			return err
		}
		actualOutputs = append(actualOutputs, actual)

		if verdict != StatusAccepted {
			continue
		}
		if !resultSucceeded(result) {
			verdict = ClassifyFromMeta(result.Meta)
		} else if i < len(expectedLines) && actual != strings.TrimSpace(expectedLines[i]) {
			verdict = StatusWrongAnswer
		}
	}

	sub.UserOutput = strings.Join(userOutputs, "\n")
	sub.ActualOutput = strings.Join(actualOutputs, "\n")
	sub.Status = verdict
	return nil
}

// executeSubmit reads the problem's single input file and single expected-output file,
// where each line is one test case. Compiles once, then runs each line sequentially.
// Stops at the first failure and stores only that test case's input/actual/expected.
func (r *CodeRunner) executeSubmit(ctx context.Context) error {
	prob := &r.submission.Problem

	inputData, err := os.ReadFile(prob.InputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file %q: %w", prob.InputFile, err)
	}
	expectedData, err := os.ReadFile(prob.ExpectedOutputFile)
	if err != nil {
		return fmt.Errorf("failed to read expected output file %q: %w", prob.ExpectedOutputFile, err)
	}

	inputLines := splitLines(string(inputData))
	expectedLines := splitLines(string(expectedData))

	if len(inputLines) != len(expectedLines) {
		return fmt.Errorf("line count mismatch: %d input lines vs %d expected output lines", len(inputLines), len(expectedLines))
	}

	sub := &r.submission.Submission
	for i, inputLine := range inputLines {
		result, runErr := r.runOnce(ctx, inputLine)
		if runErr != nil {
			return runErr
		}
		r.applyMeta(result)

		actual, err := r.getActualOutput()
		if err != nil {
			return err
		}
		actual = strings.TrimSpace(actual)
		expected := strings.TrimSpace(expectedLines[i])
		userOutput := strings.TrimSpace(result.Stdout)

		if !resultSucceeded(result) {
			sub.Status = ClassifyFromMeta(result.Meta)
			sub.ErrTestCaseInput = inputLine
			sub.ErrTestCaseOutput = expected
			sub.ActualOutput = actual
			sub.UserOutput = userOutput
			sub.Stderr = fmt.Sprintf("test case %d: %s", i+1, result.Stderr)
			return nil
		}
		if actual != expected {
			sub.Status = StatusWrongAnswer
			sub.ErrTestCaseInput = inputLine
			sub.ErrTestCaseOutput = expected
			sub.ActualOutput = actual
			sub.UserOutput = userOutput
			sub.Stderr = fmt.Sprintf("test case %d: wrong answer", i+1)
			return nil
		}
	}

	sub.Status = StatusAccepted
	return nil
}

// Execute runs the full lifecycle: init → compile (once) → execute all test cases → cleanup.
// Status transitions persisted via the notifier:
//
//	(caller: PROCESSING) -> COMPILING (only if language compiles) -> RUNNING -> terminal.
//
// A compile-step rejection is a terminal verdict (COMPILATION_ERROR), not an
// error return — only infrastructure failures bubble up as errors.
func (r *CodeRunner) Execute(ctx context.Context) error {
	if err := r.init(ctx); err != nil {
		return err
	}
	defer r.exec.Cleanup(ctx)

	if r.submission.Language.CompileCmd != nil {
		r.emitStatus(ctx, StatusCompiling)
		ok, err := r.compile(ctx)
		if err != nil {
			return err
		}
		if !ok {
			r.submission.Submission.Status = StatusCompilationError
			return nil
		}
	}

	r.emitStatus(ctx, StatusRunning)

	if r.userInput != nil {
		return r.executeRun(ctx)
	}
	return r.executeSubmit(ctx)
}

// splitLines splits s on newlines and drops trailing empty lines.
func splitLines(s string) []string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func (r *CodeRunner) getActualOutput() (string, error) {
	output, err := os.ReadFile(r.exec.GetSandboxDir() + "/" + r.actualOutputFileName)
	if err != nil {
		return "", err
	}
	return string(output), nil
}
