package domain

type CodeRunner struct {
	submission *SubmissionWithDetails
	stdin      *string
}

func NewCodeRunner(submission *SubmissionWithDetails, stdin *string) *CodeRunner {
	return &CodeRunner{
		submission: submission,
		stdin:      stdin,
	}
}

func (r *CodeRunner) Compile() error {
	if r.submission.Language.CompileCmd == nil {
		return nil
	}

	return nil
}

func (r *CodeRunner) Run() error {
	return nil
}

func (r *CodeRunner) Submit() error {
	return nil
}

func (r *CodeRunner) Execute() error {
	if r.stdin == nil {
		return r.Run()
	}
	return r.Submit()
}
