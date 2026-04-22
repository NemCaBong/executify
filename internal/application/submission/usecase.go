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

func (u *Usecase) Submit(ctx context.Context, input *CreateSubmissionInput) (int, error) {
	submission := input.ToDomain()
	if err := u.repo.Save(ctx, submission); err != nil {
		return 0, err
	}

	return submission.ID, nil
}

func (u *Usecase) GetStatus(ctx context.Context, id string) (*domain.Submission, error) {
	return u.repo.GetByID(ctx, id)
}
