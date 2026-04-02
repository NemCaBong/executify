package port

import (
	"context"

	"github.com/NemCaBong/executify/internal/core/domain"
)

// SubmissionRepository is a driven port (output port)
type SubmissionRepository interface {
	Save(ctx context.Context, submission *domain.Submission) error
	GetByID(ctx context.Context, id string) (*domain.Submission, error)
	Update(ctx context.Context, submission *domain.Submission) error
}

// SubmissionService is a driving port (input port)
type SubmissionService interface {
	Submit(ctx context.Context, submission *domain.Submission) error
	GetStatus(ctx context.Context, id string) (*domain.Submission, error)
}
