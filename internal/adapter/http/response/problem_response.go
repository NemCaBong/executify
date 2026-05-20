package response

import (
	"time"

	"github.com/NemCaBong/executify/internal/domain"
)

type ProblemTagResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type ProblemLanguageRefResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProblemSnippetResponse struct {
	Language     ProblemLanguageRefResponse `json:"language"`
	TemplateCode string                     `json:"template_code"`
	WrapperCode  string                     `json:"wrapper_code"`
}

type ProblemDetailsResponse struct {
	ID                 int                          `json:"id"`
	Name               string                       `json:"name"`
	Slug               *string                      `json:"slug"`
	Difficulty         *string                      `json:"difficulty"`
	IsPublic           bool                         `json:"is_public"`
	AcceptedCount      int                          `json:"accepted_count"`
	SubmissionCount    int                          `json:"submission_count"`
	Description        string                       `json:"description"`
	OutputFormat       string                       `json:"output_format"`
	SampleInput        string                       `json:"sample_input"`
	SampleOutput       string                       `json:"sample_output"`
	Hints              []string                     `json:"hints"`
	Tags               []ProblemTagResponse         `json:"tags"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
	Snippet            ProblemSnippetResponse       `json:"snippet"`
	SupportedLanguages []ProblemLanguageRefResponse `json:"supported_languages"`
}

func NewProblemDetailsResponse(d *domain.ProblemDetails) *ProblemDetailsResponse {
	p := d.Problem

	var difficulty *string
	if p.Difficulty != nil {
		s := string(*p.Difficulty)
		difficulty = &s
	}

	tags := make([]ProblemTagResponse, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = ProblemTagResponse{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}

	hints := p.Hints
	if hints == nil {
		hints = []string{}
	}

	langs := make([]ProblemLanguageRefResponse, len(d.SupportedLanguages))
	for i, l := range d.SupportedLanguages {
		langs[i] = ProblemLanguageRefResponse{ID: l.ID, Name: l.Name}
	}

	return &ProblemDetailsResponse{
		ID:              p.ID,
		Name:            p.Name,
		Slug:            p.Slug,
		Difficulty:      difficulty,
		IsPublic:        p.IsPublic,
		AcceptedCount:   p.AcceptedCount,
		SubmissionCount: p.SubmissionCount,
		Description:     p.Description,
		OutputFormat:    p.OutputFormat,
		SampleInput:     p.SampleInput,
		SampleOutput:    p.SampleOutput,
		Hints:           hints,
		Tags:            tags,
		CreatedAt:       p.CreatedAt,
		UpdatedAt:       p.UpdatedAt,
		Snippet: ProblemSnippetResponse{
			Language: ProblemLanguageRefResponse{
				ID:   d.Snippet.Language.ID,
				Name: d.Snippet.Language.Name,
			},
			TemplateCode: d.Snippet.TemplateCode,
			WrapperCode:  d.Snippet.WrapperCode,
		},
		SupportedLanguages: langs,
	}
}
