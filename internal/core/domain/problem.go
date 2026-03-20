package domain

import "time"

type Problem struct {
	ID                       int       `json:"id" db:"id"`
	Name                     string    `json:"name" db:"name"`
	Description              string    `json:"description" db:"description"`
	OutputFormat             string    `json:"output_format" db:"output_format"`
	SampleInput              string    `json:"sample_input" db:"sample_input"`
	SampleOutput             string    `json:"sample_output" db:"sample_output"`
	TimeLimit                int       `json:"time_limit" db:"time_limit"`
	MemoryLimit              int       `json:"memory_limit" db:"memory_limit"`
	CreatedAt                time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time `json:"updated_at" db:"updated_at"`
	InputFile                string    `json:"input_file" db:"input_file"`
	CPUTimeLimit             *float64  `json:"cpu_time_limit" db:"cpu_time_limit"`
	CPUExtraTime             *float64  `json:"cpu_extra_time" db:"cpu_extra_time"`
	WallTimeLimit            *float64  `json:"wall_time_limit" db:"wall_time_limit"`
	StackLimit               *int      `json:"stack_limit" db:"stack_limit"`
	MaxProcessesAndOrThreads *int      `json:"max_processes_and_or_threads" db:"max_processes_and_or_threads"`
	ExpectedOutputFile       string    `json:"expected_output_file" db:"expected_output_file"`
}
