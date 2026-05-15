package domain

import "time"

type Difficulty string

const (
	DifficultyEasy   Difficulty = "easy"
	DifficultyMedium Difficulty = "medium"
	DifficultyHard   Difficulty = "hard"
)

type Problem struct {
	ID                       int         `json:"id"`
	Name                     string      `json:"name"`
	Slug                     *string     `json:"slug"`
	Difficulty               *Difficulty `json:"difficulty"`
	IsPublic                 bool        `json:"is_public"`
	AcceptedCount            int         `json:"accepted_count"`
	SubmissionCount          int         `json:"submission_count"`
	Description              string      `json:"description"`
	OutputFormat             string      `json:"output_format"`
	SampleInput              string      `json:"sample_input"`
	SampleOutput             string      `json:"sample_output"`
	TimeLimit                int         `json:"time_limit"`
	MemoryLimit              int         `json:"memory_limit"`
	CreatedAt                time.Time   `json:"created_at"`
	UpdatedAt                time.Time   `json:"updated_at"`
	InputFile                string      `json:"input_file"`
	CPUTimeLimit             *float64    `json:"cpu_time_limit"`
	CPUExtraTime             *float64    `json:"cpu_extra_time"`
	WallTimeLimit            *float64    `json:"wall_time_limit"`
	StackLimit               *int        `json:"stack_limit"`
	MaxProcessesAndOrThreads *int        `json:"max_processes_and_or_threads"`
	ExpectedOutputFile       string      `json:"expected_output_file"`
	TemplateCode             string      `json:"template_code"`
	WrapperCode              string      `json:"wrapper_code"`
	Tags                     []Tag       `json:"tags"`
	Hints                    []string    `json:"hints"`
}
