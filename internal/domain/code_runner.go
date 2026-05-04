package domain

import (
	"context"
	"fmt"
	"math"
	"os"

	isolate "github.com/NemCaBong/go-isolate"
)

const (
	sandboxStdin  = "stdin.txt"
	sandboxStdout = "stdout.txt"
	sandboxStderr = "stderr.txt"

	// BoxModulus is the Mersenne prime 2^31−1 used to map submission IDs to box IDs.
	BoxModulus = 2147483647
)

type CodeRunner struct {
	submission *SubmissionWithDetails
	stdin      *string // non-nil = run mode (user-provided stdin); nil = submit mode (uses problem InputFile)
	exec       *isolate.Executor
}

func NewCodeRunner(submission *SubmissionWithDetails, stdin *string) *CodeRunner {
	return &CodeRunner{
		submission: submission,
		stdin:      stdin,
	}
}

// init creates the sandbox, mounts system dirs, and configures meta-file output.
func (r *CodeRunner) init(ctx context.Context) error {
	boxId := r.submission.ID % BoxModulus
	metaPath := fmt.Sprintf("/tmp/meta-%d.txt", boxId)
	builder := isolate.New().
		BoxID(boxId).
		Meta(metaPath).
		FullEnv().
		DirSimple("/usr").
		DirSimple("/etc").
		DirSimple("/lib").
		DirSimple("/var")
	r.exec = builder.Exec()
	if _, err := r.exec.Init(ctx); err != nil {
		r.exec.Cleanup(ctx)
		return err
	}
	return nil
}

// compile writes the source code and runs the compiler inside the sandbox.
// It is a no-op for interpreted languages.
func (r *CodeRunner) compile(ctx context.Context) error {
	if r.submission.Language.CompileCmd == nil {
		return nil
	}

	r.exec.ApplyOptions(
		isolate.WithMemoryLimit(512*1024),
		isolate.WithTimeLimit(10.0),
		isolate.WithWallTimeLimit(20.0),
		isolate.WithProcesses(10),
	)

	if err := r.exec.WriteToSandbox(
		r.submission.Language.SourceFile,
		[]byte(r.submission.SourceCode),
		0644,
	); err != nil {
		return fmt.Errorf("failed to write source code: %w", err)
	}

	result, err := r.exec.Run(ctx, "/bin/sh", "-c", *r.submission.Language.CompileCmd)
	if err != nil {
		return fmt.Errorf("compile command failed: %w", err)
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("compilation failed: %s", result.Stderr)
	}

	return nil
}

// executeProgram runs the program with stdin/stdout redirected and applies problem limits.
// Results are written back into the embedded Submission.
func (r *CodeRunner) executeProgram(ctx context.Context) error {
	prob := &r.submission.Problem

	var stdinContent []byte
	if r.stdin != nil {
		// run with stdin flow
		stdinContent = []byte(*r.stdin)
	} else {
		// TODO: Maybe store input file in object storage rather in own computer
		data, err := os.ReadFile(prob.InputFile)
		if err != nil {
			return fmt.Errorf("failed to read problem input file: %w", err)
		}
		stdinContent = data
	}

	if err := r.exec.WriteToSandbox(sandboxStdin, stdinContent, 0644); err != nil {
		return fmt.Errorf("failed to write stdin to sandbox: %w", err)
	}

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
		isolate.WithStdin(r.submission.Problem.InputFile),
		isolate.WithStdout(sandboxStdout),
		isolate.WithStderr(sandboxStderr),
	}
	if prob.CPUExtraTime != nil {
		opts = append(opts, isolate.WithExtraTime(*prob.CPUExtraTime))
	}
	if prob.StackLimit != nil {
		opts = append(opts, isolate.WithStackLimit(*prob.StackLimit))
	}
	r.exec.ApplyOptions(opts...)

	result, err := r.exec.Run(ctx, "/bin/sh", "-c", r.submission.Language.RunCmd)
	if err != nil {
		return fmt.Errorf("run command failed: %w", err)
	}

	r.applyResult(result)
	return nil
}

func (r *CodeRunner) applyResult(result *isolate.Result) {
	sub := &r.submission.Submission
	sub.Stdout = result.Stdout
	sub.Stderr = result.Stderr

	if result.Meta != nil {
		exitCode := result.Meta.ExitCode
		sub.ExitCode = &exitCode

		if result.Meta.ExitSignal != 0 {
			sig := result.Meta.ExitSignal
			sub.ExitSignal = &sig
		}

		timeMs := int(math.Round(result.Meta.Time * 1000))
		sub.TimeUsed = &timeMs

		memKB := result.Meta.MaxRSS
		sub.MemoryUsed = &memKB

		if result.Meta.IsSuccess() {
			sub.Status = StatusCompleted
		} else {
			sub.Status = StatusFailed
		}
	} else {
		exitCode := result.ExitCode
		sub.ExitCode = &exitCode
		if result.ExitCode == 0 && result.Stderr == "" {
			sub.Status = StatusCompleted
		} else {
			sub.Status = StatusFailed
		}
	}
}

// Execute runs the full lifecycle: init → compile (if needed) → execute → cleanup.
// Results are written into the embedded Submission fields.
func (r *CodeRunner) Execute(ctx context.Context) error {
	if err := r.init(ctx); err != nil {
		return err
	}
	defer r.exec.Cleanup(ctx)

	if err := r.compile(ctx); err != nil {
		return err
	}
	return r.executeProgram(ctx)
}
