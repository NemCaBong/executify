package request

type TagInput struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"`
	Slug string `json:"slug" binding:"required"`
}

type IOSchemaInput struct {
	Kind      string `json:"kind" binding:"required,oneof=input output"`
	LineIndex int    `json:"line_index"`
	KeyName   string `json:"key_name" binding:"required"`
	DataType  string `json:"data_type" binding:"required"`
}

type UpsertProblemRequest struct {
	ID                       int             `json:"id"`
	Name                     string          `json:"name" binding:"required"`
	Slug                     *string         `json:"slug"`
	Difficulty               *string         `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	IsPublic                 *bool           `json:"is_public"`
	Description              string          `json:"description" binding:"required"`
	OutputFormat             string          `json:"output_format"`
	SampleInput              string          `json:"sample_input"`
	SampleOutput             string          `json:"sample_output"`
	TimeLimit                int             `json:"time_limit" binding:"required"`
	MemoryLimit              int             `json:"memory_limit" binding:"required"`
	InputDir                 string          `json:"input_dir"`
	ExpectedOutputDir        string          `json:"expected_output_dir"`
	CPUTimeLimit             *float64        `json:"cpu_time_limit"`
	CPUExtraTime             *float64        `json:"cpu_extra_time"`
	WallTimeLimit            *float64        `json:"wall_time_limit"`
	StackLimit               *int            `json:"stack_limit"`
	MaxProcessesAndOrThreads *int            `json:"max_processes_and_or_threads"`
	FloatPrecision           *int            `json:"float_precision" binding:"omitempty,min=0"`
	Hints                    []string        `json:"hints"`
	Tags                     []TagInput      `json:"tags"`
	IOSchema                 []IOSchemaInput `json:"io_schema"`
}
