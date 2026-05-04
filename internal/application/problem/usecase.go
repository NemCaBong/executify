package problem

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Upsert(ctx context.Context, problem *domain.Problem) (*domain.Problem, error) {
	return u.repo.Upsert(ctx, problem)
}
