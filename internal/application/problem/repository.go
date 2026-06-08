package problem

import (
	"context"

	"github.com/NemCaBong/executify/internal/domain"
)

type Repository interface {
	GetBySlug(ctx context.Context, slug string) (*domain.Problem, error)
	Upsert(ctx context.Context, problem *domain.Problem) (*domain.Problem, error)
	// GetWrapperCode returns the wrapper template registered for a
	// (problem, language) pair from problem_languages.
	GetWrapperCode(ctx context.Context, problemID, languageID int) (string, error)

	// GetProblemLanguageSnippet returns the template + wrapper code along
	// with language metadata for a (problem, language) pair.
	GetProblemLanguageSnippet(ctx context.Context, problemID, languageID int) (*domain.ProblemLanguageSnippet, error)

	// ListProblemLanguages returns every language registered for the problem,
	// ordered by language ID.
	ListProblemLanguages(ctx context.Context, problemID int) ([]domain.Language, error)

	// FindLanguageByName does a case-insensitive substring match against the
	// language name and returns the first hit (by ID).
	FindLanguageByName(ctx context.Context, query string) (*domain.Language, error)

	// FindLanguageBySlug returns the language with an exact slug match.
	FindLanguageBySlug(ctx context.Context, slug string) (*domain.Language, error)
}
