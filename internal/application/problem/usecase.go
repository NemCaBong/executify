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
func (u *Usecase) GetDetails(ctx context.Context, id int, languageQuery string) (*domain.ProblemDetails, error) {
	prob, err := u.repo.GetByID(ctx, id)
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
