package problem

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id int) (*domain.Problem, error)
}
