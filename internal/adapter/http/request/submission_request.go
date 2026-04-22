package request

type SubmitRequest struct {
	LanguageID int    `json:"language_id"`
	ProblemID  int    `json:"problem_id"`
	SourceCode string `json:"source_code"`
}
