package submission

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{
		repo: repo,
	}
}

func (u *Usecase) Submit(ctx context.Context, submission *domain.Submission) error {
	submission.Status = domain.StatusPending
	if err := u.repo.Save(ctx, submission); err != nil {
		return err
	}

	// In a real scenario, this might be sent to a queue
	// For now, we call the executor (which could be the worker)
	return nil
}

func (u *Usecase) GetStatus(ctx context.Context, id string) (*domain.Submission, error) {
	return u.repo.GetByID(ctx, id)
}
