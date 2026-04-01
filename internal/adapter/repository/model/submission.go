package model

import (
	"time"

	"github.com/NemCaBong/executify/internal/core/domain"
)

// Submission là struct đại diện cho Table "submissions" dưới Database. (Dành riêng cho tầng GORM)
type Submission struct {
	ID         int        `gorm:"primaryKey;column:id;autoIncrement"`
	LanguageID int        `gorm:"column:language_id"`
	ProblemID  int        `gorm:"column:problem_id"`
	SourceCode string     `gorm:"column:source_code"`
	Stdin      string     `gorm:"column:stdin"`
	Stdout     string     `gorm:"column:stdout"`
	Stderr     string     `gorm:"column:stderr"`
	Status     string     `gorm:"column:status"`
	CreatedAt  time.Time  `gorm:"column:created_at;autoCreateTime"`
	ExitCode   *int       `gorm:"column:exit_code"`
	ExitSignal *int       `gorm:"column:exit_signal"`
	FinishedAt *time.Time `gorm:"column:finished_at"`
	TimeUsed   *int       `gorm:"column:time_used"`
	MemoryUsed *int       `gorm:"column:memory_used"`
}

// TableName overrides the table name used by GORM
func (Submission) TableName() string {
	return "submissions"
}

// ToDomain ánh xạ từ Database Model lên Core Domain Model
func (m *Submission) ToDomain() *domain.Submission {
	if m == nil {
		return nil
	}
	return &domain.Submission{
		ID:         m.ID,
		LanguageID: m.LanguageID,
		ProblemID:  m.ProblemID,
		SourceCode: m.SourceCode,
		Stdin:      m.Stdin,
		Stdout:     m.Stdout,
		Stderr:     m.Stderr,
		Status:     domain.SubmissionStatus(m.Status),
		CreatedAt:  m.CreatedAt,
		ExitCode:   m.ExitCode,
		ExitSignal: m.ExitSignal,
		FinishedAt: m.FinishedAt,
		TimeUsed:   m.TimeUsed,
		MemoryUsed: m.MemoryUsed,
	}
}

// FromDomain ánh xạ từ Core Domain Model sang Database Model
func FromDomain(d *domain.Submission) *Submission {
	if d == nil {
		return nil
	}
	return &Submission{
		ID:         d.ID,
		LanguageID: d.LanguageID,
		ProblemID:  d.ProblemID,
		SourceCode: d.SourceCode,
		Stdin:      d.Stdin,
		Stdout:     d.Stdout,
		Stderr:     d.Stderr,
		Status:     string(d.Status),
		CreatedAt:  d.CreatedAt,
		ExitCode:   d.ExitCode,
		ExitSignal: d.ExitSignal,
		FinishedAt: d.FinishedAt,
		TimeUsed:   d.TimeUsed,
		MemoryUsed: d.MemoryUsed,
	}
}
