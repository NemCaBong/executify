package problem

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/NemCaBong/executify/internal/domain"
)

var (
	ErrProblemNotFound      = errors.New("problem not found")
	ErrLanguageNotFound     = errors.New("language not found")
	ErrLanguageNotSupported = errors.New("language not supported for this problem")
)

const defaultDetailsLanguage = "python"

type Usecase struct {
	repo Repository
}

func NewUsecase(repo Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (u *Usecase) Upsert(ctx context.Context, problem *domain.Problem) (*domain.Problem, error) {
	return u.repo.Upsert(ctx, problem)
}

// GetDetails loads the problem along with a language-specific code snippet
// (template + wrapper) and the list of all languages the problem supports.
//
// languageQuery is matched case-insensitively against language names. When
// empty, Python is used as the default.
func (u *Usecase) GetDetails(ctx context.Context, slug string, languageQuery string) (*domain.ProblemDetails, error) {
	prob, err := u.repo.GetBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	q := strings.TrimSpace(languageQuery)
	if q == "" {
		q = defaultDetailsLanguage
	}
	lang, err := u.repo.FindLanguageByName(ctx, strings.ToLower(q))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLanguageNotFound
		}
		return nil, err
	}

	snippet, err := u.repo.GetProblemLanguageSnippet(ctx, prob.ID, lang.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLanguageNotSupported
		}
		return nil, err
	}

	langs, err := u.repo.ListProblemLanguages(ctx, prob.ID)
	if err != nil {
		return nil, err
	}

	return &domain.ProblemDetails{
		Problem:            prob,
		Snippet:            snippet,
		SupportedLanguages: langs,
	}, nil
}

// GetSnippet returns the template code and language metadata for a single
// (problem, language) pair, looked up by the problem slug and language slug.
// This is the lightweight endpoint used when the user switches language on a
// problem they already have loaded.
func (u *Usecase) GetSnippet(ctx context.Context, problemSlug, languageSlug string) (*domain.ProblemLanguageSnippet, error) {
	prob, err := u.repo.GetBySlug(ctx, problemSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	lang, err := u.repo.FindLanguageBySlug(ctx, strings.ToLower(strings.TrimSpace(languageSlug)))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLanguageNotFound
		}
		return nil, err
	}

	snippet, err := u.repo.GetProblemLanguageSnippet(ctx, prob.ID, lang.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLanguageNotSupported
		}
		return nil, err
	}

	return snippet, nil
}
