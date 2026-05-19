package request

type TagInput struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type UpsertProblemRequest struct {
	ID                       int        `json:"id"`
	Name                     string     `json:"name" binding:"required"`
	Slug                     *string    `json:"slug"`
	Difficulty               *string    `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	IsPublic                 *bool      `json:"is_public"`
	Description              string     `json:"description" binding:"required"`
	OutputFormat             string     `json:"output_format"`
	SampleInput              string     `json:"sample_input"`
	SampleOutput             string     `json:"sample_output"`
	TimeLimit                int        `json:"time_limit" binding:"required"`
	MemoryLimit              int        `json:"memory_limit" binding:"required"`
	InputFile                string     `json:"input_file"`
	ExpectedOutputFile       string     `json:"expected_output_file"`
	CPUTimeLimit             *float64   `json:"cpu_time_limit"`
	CPUExtraTime             *float64   `json:"cpu_extra_time"`
	WallTimeLimit            *float64   `json:"wall_time_limit"`
	StackLimit               *int       `json:"stack_limit"`
	MaxProcessesAndOrThreads *int       `json:"max_processes_and_or_threads"`
	Hints                    []string   `json:"hints"`
	Tags                     []TagInput `json:"tags"`
}
