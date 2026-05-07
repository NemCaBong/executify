package request

type SubmitRequest struct {
	LanguageID int    `json:"language_id"`
	ProblemID  int    `json:"problem_id"`
	SourceCode string `json:"source_code"`
}

type RunRequest struct {
	LanguageID     int    `json:"language_id" binding:"required"`
	ProblemID      int    `json:"problem_id" binding:"required"`
	SourceCode     string `json:"source_code" binding:"required"`
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}
