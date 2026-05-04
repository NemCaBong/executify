package problem

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id int) (*domain.Problem, error)
	Upsert(ctx context.Context, problem *domain.Problem) (*domain.Problem, error)
}
