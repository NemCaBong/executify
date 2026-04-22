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
