package submission

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

// SubmissionRepository is a driven port (output port)
type Repository interface {
	Save(ctx context.Context, submission *domain.Submission) error
	GetByID(ctx context.Context, id string) (*domain.Submission, error)
	Update(ctx context.Context, submission *domain.Submission) error
}
