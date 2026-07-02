package domain

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/NemCaBong/go-isolate"

	"github.com/NemCaBong/executify/internal/config"
)

// CommandLogger receives the fully-rendered isolate command line at each
// lifecycle stage. Wired via WithCommandLogger;
// left nil to disable command logging entirely.
type CommandLogger func(stage, command string)

type CodeRunner struct {
	submission           *SubmissionWithDetails
	userInput            *string // non-nil = run mode (user-provided input lines); nil = submit mode (uses problem dir)
	userExpectedOutput   *string
	exec                 *isolate.Executor
	builder              *isolate.Builder
	codeRunnerCfg        config.CodeRunnerConfig
	stdoutFileName       string
	stderrFileName       string
	actualOutputFileName string
	stdinFileName        string
	metaFileName         string
	notifyStatus         StatusNotifier
	logCmd               CommandLogger
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
		stdinFileName:        hash + StdinFileName,
		// meta file need to have path
		metaFileName: filepath.Join(os.TempDir(), hash+MetaFileName),
	}
}

// WithStatusNotifier wires a callback that is invoked at each lifecycle
// transition (COMPILING, RUNNING). Optional — when unset, only the terminal
// verdict is reflected on the submission.
func (r *CodeRunner) WithStatusNotifier(fn StatusNotifier) *CodeRunner {
	r.notifyStatus = fn
	return r
}

// WithCommandLogger wires a callback that receives the rendered isolate command
// at each lifecycle stage (init → compile → run → cleanup). Optional — when
// unset, no command logging happens. Enabled per-request via the
// X-Enable-Log-Command header (see the submission handler / worker).
func (r *CodeRunner) WithCommandLogger(fn CommandLogger) *CodeRunner {
	r.logCmd = fn
	return r
}

// logCommand renders nothing and fires the command logger only when one is
// configured, so the BuildXxx() string is not even built in the common case.
func (r *CodeRunner) logCommand(stage string, cmd *isolate.Command) {
	if r.logCmd != nil && cmd != nil {
		r.logCmd(stage, cmd.String())
	}
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
	// Keep the builder: Exec() shares this exact pointer with the executor and
	// ApplyOptions mutates it, so builder.BuildInit/BuildRun/BuildCleanup render
	// the same command line the executor actually runs.
	r.builder = isolate.New().
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
		InheritFDs() // to route user res to another fd
	r.exec = r.builder.Exec()

	r.logCommand("init", r.builder.BuildInit())
	if _, err := r.exec.Init(ctx); err != nil {
		r.exec.Cleanup(ctx)
		return err
	}

	// create sandbox files
	for _, name := range []string{
		r.stdoutFileName,
		r.stderrFileName,
		r.actualOutputFileName,
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

	r.logCommand("compile", r.builder.BuildRun("/bin/bash", "-c", *r.submission.Language.CompileCmd))
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
	processes := 10

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

// runOnce feeds the sandbox-relative input file to the program and captures fd3
// as the actual output. sandboxInputPath must already be written to the sandbox
// before calling this function.
func (r *CodeRunner) runOnce(ctx context.Context, sandboxInputPath string) (*isolate.Result, error) {
	r.applyExecOptions()
	cmd := fmt.Sprintf("%s < %q 3>%q", r.submission.Language.RunCmd, sandboxInputPath, r.actualOutputFileName)
	r.logCommand("run", r.builder.BuildRun("/bin/bash", "-c", cmd))
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

// executeRun runs every test case from the user-provided input without early exit.
// Each test case is written to a dedicated sandbox file before execution.
func (r *CodeRunner) executeRun(ctx context.Context) error {
	prob := &r.submission.Problem
	inParamLen, outParamLen := ioLineCounts(prob)

	inputCases, err := groupTestCases(splitLines(*r.userInput), inParamLen)
	if err != nil {
		return fmt.Errorf("user input: %w", err)
	}
	expectedCases, err := groupTestCases(splitLines(*r.userExpectedOutput), outParamLen)
	if err != nil {
		return fmt.Errorf("user expected output: %w", err)
	}
	if len(inputCases) != len(expectedCases) {
		return fmt.Errorf("expected %d test cases, got %d", len(expectedCases), len(inputCases))
	}

	outputFields := outputIOFields(prob)

	sub := &r.submission.Submission
	actualOutputs := make([]string, 0, len(inputCases))
	userOutputs := make([]string, 0, len(inputCases))
	verdict := StatusAccepted

	for i, stdinContent := range inputCases {
		if err := r.exec.WriteToSandbox(r.stdinFileName, []byte(stdinContent), 0644); err != nil {
			return fmt.Errorf("write stdin to sandbox: %w", err)
		}
		result, err := r.runOnce(ctx, r.stdinFileName)
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
		}
		if !CompareOutput(outputFields, expectedCases[i], actual, prob.FloatPrecision) {
			verdict = StatusWrongAnswer
		}
	}

	sub.UserOutput = strings.Join(userOutputs, "\n")
	sub.ActualOutput = strings.Join(actualOutputs, "\n")
	sub.Status = verdict
	return nil
}

// executeSubmit reads individual input and expected-output files from the
// problem's directories and runs every test case.
func (r *CodeRunner) executeSubmit(ctx context.Context) error {
	prob := &r.submission.Problem

	inputFiles, err := readSortedTestCaseFiles(prob.InputDir)
	if err != nil {
		return fmt.Errorf("read input dir %q: %w", prob.InputDir, err)
	}
	expectedFiles, err := readSortedTestCaseFiles(prob.ExpectedOutputDir)
	if err != nil {
		return fmt.Errorf("read expected output dir %q: %w", prob.ExpectedOutputDir, err)
	}
	if len(inputFiles) != len(expectedFiles) {
		return fmt.Errorf("test case count mismatch: %d input vs %d expected", len(inputFiles), len(expectedFiles))
	}

	outputFields := outputIOFields(prob)
	sub := &r.submission.Submission

	for i, inputEntry := range inputFiles {
		inputData, err := os.ReadFile(filepath.Join(prob.InputDir, inputEntry.Name()))
		if err != nil {
			return fmt.Errorf("read input file %q: %w", inputEntry.Name(), err)
		}
		if err := r.exec.WriteToSandbox(r.stdinFileName, inputData, 0644); err != nil {
			return fmt.Errorf("write stdin to sandbox: %w", err)
		}

		result, runErr := r.runOnce(ctx, r.stdinFileName)
		if runErr != nil {
			return runErr
		}
		r.applyMeta(result)

		expectedData, err := os.ReadFile(filepath.Join(prob.ExpectedOutputDir, expectedFiles[i].Name()))
		if err != nil {
			return fmt.Errorf("read expected output file %q: %w", expectedFiles[i].Name(), err)
		}
		expected := strings.TrimSpace(string(expectedData))

		actual, err := r.getActualOutput()
		if err != nil {
			return err
		}
		userOutput := strings.TrimSpace(result.Stdout)

		if !resultSucceeded(result) {
			sub.Status = ClassifyFromMeta(result.Meta)
			sub.ErrTestCaseInput = strings.TrimSpace(string(inputData))
			sub.ErrTestCaseOutput = expected
			sub.ActualOutput = actual
			sub.UserOutput = userOutput
			sub.Stderr = fmt.Sprintf("test case %d: %s", i+1, result.Stderr)
			return nil
		}
		if !CompareOutput(outputFields, expected, actual, prob.FloatPrecision) {
			sub.Status = StatusWrongAnswer
			sub.ErrTestCaseInput = strings.TrimSpace(string(inputData))
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

// outputIOFields returns the problem's output IO fields.
func outputIOFields(prob *Problem) []ProblemIOField {
	var fields []ProblemIOField
	for _, f := range prob.IOSchema {
		if f.Kind == IOKindOutput {
			fields = append(fields, f)
		}
	}
	return fields
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
	defer func() {
		r.logCommand("cleanup", r.builder.BuildCleanup())
		r.exec.Cleanup(ctx)
	}()

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

func splitLines(s string) []string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	// remove empty spaces after splitting. E.g: ["h", "e", "", ""] -> ["h", "e"]
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

// ioLineCounts reports how many lines make up one test case's input and one
// test case's expected output, derived from the problem's IO schema (each field
// occupies one line). Falls back to 1/1 when no schema is declared, preserving
// the legacy one-line-per-test-case behavior.
func ioLineCounts(prob *Problem) (in, out int) {
	for _, f := range prob.IOSchema {
		if f.Kind == IOKindOutput {
			out++
		} else {
			in++
		}
	}
	if in == 0 {
		in = 1
	}
	if out == 0 {
		out = 1
	}
	return in, out
}

// groupTestCases joins every `size` consecutive lines into one test case blob
// (the multi-line stdin/expected text for a single case). Returns an error if
// the line total is not an exact multiple of size, which signals a malformed
// fixture rather than a user error.
func groupTestCases(lines []string, size int) ([]string, error) {
	if size <= 1 {
		return lines, nil
	}
	if len(lines)%size != 0 {
		return nil, fmt.Errorf("expected line count to be a multiple of %d, got %d", size, len(lines))
	}
	cases := make([]string, 0, len(lines)/size)
	for i := 0; i < len(lines); i += size {
		cases = append(cases, strings.Join(lines[i:i+size], "\n"))
	}
	return cases, nil
}

func (r *CodeRunner) getActualOutput() (string, error) {
	output, err := os.ReadFile(r.exec.GetSandboxDir() + "/" + r.actualOutputFileName)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// readSortedTestCaseFiles reads all .txt files from dir and returns them sorted
// by numeric base name (1.txt before 2.txt before 10.txt).
func readSortedTestCaseFiles(dir string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var files []os.DirEntry
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			files = append(files, e)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		ni, _ := strconv.Atoi(strings.TrimSuffix(files[i].Name(), ".txt"))
		nj, _ := strconv.Atoi(strings.TrimSuffix(files[j].Name(), ".txt"))
		return ni < nj
	})
	return files, nil
}
