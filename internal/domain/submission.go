package domain

import "time"

type SubmissionStatus string

// status flow: SUBMITTED -> PROCESSING (in queue) -> COMPLETED | FAILED
const (
	StatusCompleted  SubmissionStatus = "COMPLETED"
	StatusFailed     SubmissionStatus = "FAILED"
	StatusProcessing SubmissionStatus = "PROCESSING"
	StatusSubmitted  SubmissionStatus = "SUBMITTED"
)

type Submission struct {
	ID                int              `json:"id"`
	LanguageID        int              `json:"language_id"`
	ProblemID         int              `json:"problem_id"`
	SourceCode        string           `json:"source_code"`
	Input             string           `json:"input"`
	ExpectedOutput    string           `json:"expected_output"`
	ActualOutput      string           `json:"actual_output"`
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
