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
	Slug string `json:"slug"`
}

type ProblemSnippetResponse struct {
	Language     ProblemLanguageRefResponse `json:"language"`
	TemplateCode string                     `json:"template_code"`
}

func NewProblemSnippetResponse(s *domain.ProblemLanguageSnippet) ProblemSnippetResponse {
	return ProblemSnippetResponse{
		Language: ProblemLanguageRefResponse{
			ID:   s.Language.ID,
			Name: s.Language.Name,
			Slug: s.Language.Slug,
		},
		TemplateCode: s.TemplateCode,
	}
}

// ProblemIOFieldResponse is one named value the FE renders as a key/value
// input.
type ProblemIOFieldResponse struct {
	Name        string      `json:"name"`
	DataType    string      `json:"data_type"`
	LineIndex   int         `json:"line_index"`
	SampleValue interface{} `json:"sample_value"`
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
	Hints              []string                     `json:"hints"`
	Tags               []ProblemTagResponse         `json:"tags"`
	CreatedAt          time.Time                    `json:"created_at"`
	UpdatedAt          time.Time                    `json:"updated_at"`
	Snippet            ProblemSnippetResponse       `json:"snippet"`
	SupportedLanguages []ProblemLanguageRefResponse `json:"supported_languages"`
	Inputs             []ProblemIOFieldResponse     `json:"inputs"`
	Outputs            []ProblemIOFieldResponse     `json:"outputs"`
}

// buildIOFields assembles the response objects for one kind, attaching each
// field's parsed sample value (looked up by line_index).
func buildIOFields(fields []domain.ProblemIOField, values map[int]interface{}) []ProblemIOFieldResponse {
	out := make([]ProblemIOFieldResponse, len(fields))
	for i, f := range fields {
		out[i] = ProblemIOFieldResponse{
			Name:        f.KeyName,
			DataType:    f.DataType,
			LineIndex:   f.LineIndex,
			SampleValue: values[f.LineIndex],
		}
	}
	return out
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
		langs[i] = ProblemLanguageRefResponse{ID: l.ID, Name: l.Name, Slug: l.Slug}
	}

	var inputFields, outputFields []domain.ProblemIOField
	for _, f := range p.IOSchema {
		if f.Kind == domain.IOKindOutput {
			outputFields = append(outputFields, f)
		} else {
			inputFields = append(inputFields, f)
		}
	}

	inputValues := domain.ParseSampleValues(inputFields, p.SampleInput)
	outputValues := domain.ParseSampleValues(outputFields, p.SampleOutput)

	return &ProblemDetailsResponse{
		ID:                 p.ID,
		Name:               p.Name,
		Slug:               p.Slug,
		Difficulty:         difficulty,
		IsPublic:           p.IsPublic,
		AcceptedCount:      p.AcceptedCount,
		SubmissionCount:    p.SubmissionCount,
		Description:        p.Description,
		OutputFormat:       p.OutputFormat,
		Hints:              hints,
		Tags:               tags,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
		Snippet:            NewProblemSnippetResponse(d.Snippet),
		SupportedLanguages: langs,
		Inputs:             buildIOFields(inputFields, inputValues),
		Outputs:            buildIOFields(outputFields, outputValues),
	}
}
