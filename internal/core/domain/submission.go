package domain

import "time"

type SubmissionStatus string

const (
	StatusPending   SubmissionStatus = "PENDING"
	StatusRunning   SubmissionStatus = "RUNNING"
	StatusCompleted SubmissionStatus = "COMPLETED"
	StatusFailed    SubmissionStatus = "FAILED"
)

type Submission struct {
	ID         int              `json:"id" db:"id"`
	LanguageID int              `json:"language_id" db:"language_id"`
	ProblemID  int              `json:"problem_id" db:"problem_id"`
	SourceCode string           `json:"source_code" db:"source_code"`
	Stdin      string           `json:"stdin" db:"stdin"`
	Stdout     string           `json:"stdout" db:"stdout"`
	Stderr     string           `json:"stderr" db:"stderr"`
	Status     SubmissionStatus `json:"status" db:"status"`
	CreatedAt  time.Time        `json:"created_at" db:"created_at"`
	ExitCode   *int             `json:"exit_code" db:"exit_code"`
	ExitSignal *int             `json:"exit_signal" db:"exit_signal"`
	FinishedAt *time.Time       `json:"finished_at" db:"finished_at"`
	TimeUsed   *int             `json:"time_used" db:"time_used"`
	MemoryUsed *int             `json:"memory_used" db:"memory_used"`
}
