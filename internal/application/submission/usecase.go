package submission

import (
	"context"
	"fmt"
	"strings"

	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/domain"
)

type Usecase struct {
	repo        Repository
	problemRepo problem.Repository
}

func NewUsecase(repo Repository, problemRepo problem.Repository) *Usecase {
	return &Usecase{
		repo:        repo,
		problemRepo: problemRepo,
	}
}

func (u *Usecase) buildFullSourceCode(ctx context.Context, problemID, languageID int, userCode string) (string, error) {
	wrapper, err := u.problemRepo.GetWrapperCode(ctx, problemID, languageID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch wrapper for problem %d / language %d: %w", problemID, languageID, err)
	}
	return strings.Replace(wrapper, "{{.}}", userCode, 1), nil
}

func (u *Usecase) Submit(ctx context.Context, input *CreateSubmissionInput) (int, error) {
	fullCode, err := u.buildFullSourceCode(ctx, input.ProblemID, input.LanguageID, input.SourceCode)
	if err != nil {
		return 0, err
	}
	input.SourceCode = fullCode

	sub := input.ToDomain()
	if err := u.repo.Save(ctx, sub); err != nil {
		return 0, err
	}

	return sub.ID, nil
}

func (u *Usecase) Run(ctx context.Context, input *CreateRunInput) (int, error) {
	fullCode, err := u.buildFullSourceCode(ctx, input.ProblemID, input.LanguageID, input.SourceCode)
	if err != nil {
		return 0, err
	}
	input.SourceCode = fullCode

	sub := input.ToDomain()
	if err := u.repo.Save(ctx, sub); err != nil {
		return 0, err
	}

	return sub.ID, nil
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
