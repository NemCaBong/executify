package domain

import "time"

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type ProblemLanguageSnippet struct {
	Language     Language
	TemplateCode string
	WrapperCode  string
}

type IOKind string

const (
	IOKindInput  IOKind = "input"
	IOKindOutput IOKind = "output"
)

// ProblemIOField describes a single named value on one line of a problem's
// input or output file. DataType (e.g. "int[]", "int", "string", "int[][]")
// tells callers how to parse that line.
type ProblemIOField struct {
	Kind      IOKind `json:"kind"`
	LineIndex int    `json:"line_index"`
	KeyName   string `json:"key_name"`
	DataType  string `json:"data_type"`
}

type ProblemDetails struct {
	Problem            *Problem
	Snippet            *ProblemLanguageSnippet
	SupportedLanguages []Language
}

type Problem struct {
	ID                       int              `json:"id"`
	Name                     string           `json:"name"`
	Slug                     *string          `json:"slug"`
	Difficulty               *Difficulty      `json:"difficulty"`
	IsPublic                 bool             `json:"is_public"`
	AcceptedCount            int              `json:"accepted_count"`
	SubmissionCount          int              `json:"submission_count"`
	Description              string           `json:"description"`
	OutputFormat             string           `json:"output_format"`
	SampleInput              string           `json:"sample_input"`
	SampleOutput             string           `json:"sample_output"`
	TimeLimit                int              `json:"time_limit"`
	MemoryLimit              int              `json:"memory_limit"`
	CreatedAt                time.Time        `json:"created_at"`
	UpdatedAt                time.Time        `json:"updated_at"`
	InputFile                string           `json:"input_file"`
	CPUTimeLimit             *float64         `json:"cpu_time_limit"`
	CPUExtraTime             *float64         `json:"cpu_extra_time"`
	WallTimeLimit            *float64         `json:"wall_time_limit"`
	StackLimit               *int             `json:"stack_limit"`
	MaxProcessesAndOrThreads *int             `json:"max_processes_and_or_threads"`
	FloatPrecision           *int             `json:"float_precision"`
	ExpectedOutputFile       string           `json:"expected_output_file"`
	Tags                     []Tag            `json:"tags"`
	Hints                    []string         `json:"hints"`
	IOSchema                 []ProblemIOField `json:"io_schema"`
}
