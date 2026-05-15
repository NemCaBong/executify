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
		Status:     domain.StatusQueued,
	}
}

type CreateRunInput struct {
	LanguageID     int
	ProblemID      int
	SourceCode     string
	Input          string
	ExpectedOutput string
}

func (i CreateRunInput) ToDomain() *domain.Submission {
	return &domain.Submission{
		LanguageID:     i.LanguageID,
		ProblemID:      i.ProblemID,
		SourceCode:     i.SourceCode,
		Input:          i.Input,
		ExpectedOutput: i.ExpectedOutput,
		Status:         domain.StatusQueued,
	}
}
