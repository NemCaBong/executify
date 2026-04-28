package domain

import (
	"context"
	"fmt"

	isolate "github.com/NemCaBong/go-isolate"
)

type CodeRunner struct {
	submission *SubmissionWithDetails
	stdin      *string
	boxId      int
	exec       *isolate.Executor // only reusing 1 executor for entire process
	workDir    *string
}

func NewCodeRunner(submission *SubmissionWithDetails, stdin *string) *CodeRunner {
	return &CodeRunner{
		submission: submission,
		stdin:      stdin,
	}
}

func (r *CodeRunner) init(ctx context.Context) error {
	builder := isolate.New().BoxID(r.boxId)
	r.exec = builder.Exec()
	_, err := r.exec.Init(ctx)
	if err != nil {
		r.exec.Cleanup(ctx)
		return err
	}

	return nil
}

func (r *CodeRunner) compile(ctx context.Context) error {
	if r.submission.Language.CompileCmd == nil {
		return nil
	}
	// fixed resources to compile program
	r.exec.ApplyOptions(
		isolate.WithMemoryLimit(512*1024),
		isolate.WithTimeLimit(10.0),
		isolate.WithWallTimeLimit(20.0),
		isolate.WithProcesses(10),
		isolate.WithFullEnv(),
		isolate.WithDirSimple("/usr"),
		isolate.WithDirSimple("/etc"),
		isolate.WithDirSimple("/lib"),
		isolate.WithDirSimple("/var"),
	)

	// Initialize the sandbox with these settings
	_, err := r.exec.Init(ctx)
	if err != nil {
		return err
	}

	// Write the source code into the sandbox
	err = r.exec.WriteToSandbox(
		r.submission.Language.SourceFile,
		[]byte(r.submission.SourceCode),
		0644,
	)
	if err != nil {
		return fmt.Errorf("failed to write source code: %w", err)
	}

	// Run the compilation command. We use /bin/sh to support commands with spaces.
	result, err := r.exec.Run(ctx, "/bin/sh", "-c", *r.submission.Language.CompileCmd)
	if err != nil {
		return fmt.Errorf("failed to execute compile command: %w", err)
	}

	if result.ExitCode != 0 {
		return fmt.Errorf("compilation failed: %s", result.Stdout)
	}

	return nil
}

func (r *CodeRunner) run(ctx context.Context) error {
	err := r.compile(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *CodeRunner) submit(ctx context.Context) error {
	err := r.compile(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *CodeRunner) Execute(ctx context.Context) error {
	if r.stdin == nil {
		return r.run(ctx)
	}
	return r.submit(ctx)
}
