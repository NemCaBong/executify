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
	ID        string           `json:"id"`
	UserID    string           `json:"user_id"`
	Code      string           `json:"code"`
	Language  string           `json:"language"`
	Status    SubmissionStatus `json:"status"`
	Result    string           `json:"result"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}
