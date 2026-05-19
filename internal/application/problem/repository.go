package problem

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

type Repository interface {
	GetByID(ctx context.Context, id int) (*domain.Problem, error)
	Upsert(ctx context.Context, problem *domain.Problem) (*domain.Problem, error)
	// GetWrapperCode returns the wrapper template registered for a
	// (problem, language) pair from problem_languages.
	GetWrapperCode(ctx context.Context, problemID, languageID int) (string, error)
}
