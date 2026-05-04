package submission

import "github.com/NemCaBong/executify/internal/domain"

type CreateSubmissionInput struct {
	LanguageID int
	ProblemID  int
	SourceCode string
}

func (i CreateSubmissionInput) ToDomain() *domain.Submission {
	return &domain.Submission{
		LanguageID: i.LanguageID,
		ProblemID:  i.ProblemID,
		SourceCode: i.SourceCode,
		Status:     domain.StatusSubmitted,
	}
}

type CreateRunInput struct {
	LanguageID int
	ProblemID  int
	SourceCode string
	Stdin      string
}

func (i CreateRunInput) ToDomain() *domain.Submission {
	return &domain.Submission{
		LanguageID: i.LanguageID,
		ProblemID:  i.ProblemID,
		SourceCode: i.SourceCode,
		Stdin:      i.Stdin,
		Status:     domain.StatusSubmitted,
	}
}
