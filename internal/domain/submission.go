package domain

import (
	"context"
	"time"

	"github.com/NemCaBong/go-isolate"
)

type SubmissionStatus string

// Lifecycle: QUEUED -> PROCESSING (picked up by worker) ->
// COMPILING (only if the language has a compile step) -> RUNNING ->
// one of the terminal verdicts below. Terminal verdicts are final.
const (
	StatusQueued     SubmissionStatus = "QUEUED"
	StatusProcessing SubmissionStatus = "PROCESSING"
	StatusCompiling  SubmissionStatus = "COMPILING"
	StatusRunning    SubmissionStatus = "RUNNING"

	StatusAccepted            SubmissionStatus = "ACCEPTED"
	StatusWrongAnswer         SubmissionStatus = "WRONG_ANSWER"
	StatusTimeLimitExceeded   SubmissionStatus = "TIME_LIMIT_EXCEEDED"
	StatusMemoryLimitExceeded SubmissionStatus = "MEMORY_LIMIT_EXCEEDED"
	StatusRuntimeError        SubmissionStatus = "RUNTIME_ERROR"
	StatusCompilationError    SubmissionStatus = "COMPILATION_ERROR"
	StatusInternalError       SubmissionStatus = "INTERNAL_ERROR"
)

// StatusNotifier is invoked by CodeRunner at phase transitions so the worker
// can persist intermediate states (COMPILING, RUNNING) without coupling the
// domain layer to a repository. Returning a non-nil error is logged but does
// not abort execution — the final terminal verdict is what actually matters.
type StatusNotifier func(ctx context.Context, status SubmissionStatus) error

// ClassifyFromMeta maps an isolate meta-file result to a sandbox-level verdict.
// It only reflects what the sandbox observed: WRONG_ANSWER must be decided by
// the caller after comparing program output against expected output.
//
// A nil meta means isolate produced no metadata (the run itself was broken),
// which is treated as INTERNAL_ERROR rather than a program-level failure.
//
// MLE detection: isolate reports OOM kills as Status="SG" with CGOOMKilled set,
// so the CGOOMKilled check must come before the generic signal branch.
func ClassifyFromMeta(meta *isolate.Meta) SubmissionStatus {
	if meta == nil {
		return StatusInternalError
	}
	if meta.CGOOMKilled {
		return StatusMemoryLimitExceeded
	}
	switch meta.Status {
	case isolate.StatusOK:
		if meta.ExitCode == 0 && !meta.Killed {
			return StatusAccepted
		}
		return StatusRuntimeError
	case isolate.StatusTimeout:
		return StatusTimeLimitExceeded
	case isolate.StatusSignal, isolate.StatusRuntimeError:
		return StatusRuntimeError
	case isolate.StatusInternalError:
		return StatusInternalError
	default:
		return StatusInternalError
	}
}

type Submission struct {
	ID                int              `json:"id"`
	LanguageID        int              `json:"language_id"`
	ProblemID         int              `json:"problem_id"`
	SourceCode        string           `json:"source_code"`
	Input             string           `json:"input"`
	ExpectedOutput    string           `json:"expected_output"`
	ActualOutput      string           `json:"actual_output"`
	UserOutput        string           `json:"user_output"`
	Stderr            string           `json:"stderr"`
	Status            SubmissionStatus `json:"status"`
	CreatedAt         time.Time        `json:"created_at"`
	ExitCode          *int             `json:"exit_code"`
	ExitSignal        *int             `json:"exit_signal"`
	FinishedAt        *time.Time       `json:"finished_at"`
	TimeUsed          *int             `json:"time_used"`
	MemoryUsed        *int             `json:"memory_used"`
	ErrTestCaseInput  string           `json:"err_test_case_input"`
	ErrTestCaseOutput string           `json:"err_test_case_output"`
}

type SubmissionWithDetails struct {
	Submission
	Language Language `json:"language"`
	Problem  Problem  `json:"problem"`
}
