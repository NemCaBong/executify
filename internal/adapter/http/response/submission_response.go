package response

import (
	"time"

	"github.com/NemCaBong/executify/internal/domain"
)

type SubmitResponse struct {
	Message string `json:"message"`
	Data    struct {
		ID int `json:"id"`
	} `json:"data"`
}

func NewSubmitResponse(id int) *SubmitResponse {
	return &SubmitResponse{
		Message: "Submission created successfully",
		Data: struct {
			ID int `json:"id"`
		}{
			ID: id,
		},
	}
}

type SubmissionStatusResponse struct {
	ID         int    `json:"id"`
	Status     string `json:"status"`
	IsTerminal bool   `json:"is_terminal"`
	LanguageID int    `json:"language_id"`
	ProblemID  int    `json:"problem_id"`

	ActualOutput      string     `json:"actual_output"`
	UserOutput        string     `json:"user_output"`
	ExpectedOutput    string     `json:"expected_output"`
	Stderr            string     `json:"stderr"`
	ExitCode          *int       `json:"exit_code"`
	ExitSignal        *int       `json:"exit_signal"`
	TimeUsed          *int       `json:"time_used"`
	MemoryUsed        *int       `json:"memory_used"`
	ErrTestCaseInput  string     `json:"err_test_case_input"`
	ErrTestCaseOutput string     `json:"err_test_case_output"`
	CreatedAt         time.Time  `json:"created_at"`
	FinishedAt        *time.Time `json:"finished_at"`
}

func NewSubmissionStatusResponse(s *domain.Submission) *SubmissionStatusResponse {
	return &SubmissionStatusResponse{
		ID:                s.ID,
		Status:            string(s.Status),
		IsTerminal:        s.Status.IsTerminal(),
		LanguageID:        s.LanguageID,
		ProblemID:         s.ProblemID,
		ActualOutput:      s.ActualOutput,
		UserOutput:        s.UserOutput,
		ExpectedOutput:    s.ExpectedOutput,
		Stderr:            s.Stderr,
		ExitCode:          s.ExitCode,
		ExitSignal:        s.ExitSignal,
		TimeUsed:          s.TimeUsed,
		MemoryUsed:        s.MemoryUsed,
		ErrTestCaseInput:  s.ErrTestCaseInput,
		ErrTestCaseOutput: s.ErrTestCaseOutput,
		CreatedAt:         s.CreatedAt,
		FinishedAt:        s.FinishedAt,
	}
}
