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

func (u *Usecase) Run(ctx context.Context, input *CreateRunInput) (int, error) {
	submission := input.ToDomain()
	if err := u.repo.Save(ctx, submission); err != nil {
		return 0, err
	}

	return submission.ID, nil
}

func (u *Usecase) GetByID(ctx context.Context, id int) (*domain.Submission, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *Usecase) GetWithDetailsByID(ctx context.Context, id int) (*domain.SubmissionWithDetails, error) {
	return u.repo.GetWithDetailsByID(ctx, id)
}

func (u *Usecase) Update(ctx context.Context, submission *domain.Submission) error {
	return u.repo.Update(ctx, submission)
}
