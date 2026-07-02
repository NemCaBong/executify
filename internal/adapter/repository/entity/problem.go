package entity

import (
	"time"

	"gorm.io/datatypes"

	"github.com/NemCaBong/executify/internal/domain"
)

type Problem struct {
	ID                       int                         `gorm:"column:id;primaryKey"`
	Name                     string                      `gorm:"column:name"`
	Slug                     *string                     `gorm:"column:slug"`
	Difficulty               *string                     `gorm:"column:difficulty"`
	IsPublic                 bool                        `gorm:"column:is_public;default:true"`
	AcceptedCount            int                         `gorm:"column:accepted_count;default:0"`
	SubmissionCount          int                         `gorm:"column:submission_count;default:0"`
	Description              string                      `gorm:"column:description"`
	OutputFormat             string                      `gorm:"column:output_format"`
	SampleInput              string                      `gorm:"column:sample_input"`
	SampleOutput             string                      `gorm:"column:sample_output"`
	TimeLimit                int                         `gorm:"column:time_limit"`
	MemoryLimit              int                         `gorm:"column:memory_limit"`
	CreatedAt                time.Time                   `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt                time.Time                   `gorm:"column:updated_at;autoUpdateTime"`
	InputDir                 string                      `gorm:"column:input_dir"`
	CPUTimeLimit             *float64                    `gorm:"column:cpu_time_limit"`
	CPUExtraTime             *float64                    `gorm:"column:cpu_extra_time"`
	WallTimeLimit            *float64                    `gorm:"column:wall_time_limit"`
	StackLimit               *int                        `gorm:"column:stack_limit"`
	MaxProcessesAndOrThreads *int                        `gorm:"column:max_processes_and_or_threads"`
	FloatPrecision           *int                        `gorm:"column:float_precision"`
	ExpectedOutputDir        string                      `gorm:"column:expected_output_dir"`
	Hints                    datatypes.JSONSlice[string] `gorm:"column:hints;type:json"`
	Tags                     []Tag                       `gorm:"many2many:problem_tags;joinForeignKey:problem_id;joinReferences:tag_id"`
	// has-many on problem_id. No DB foreign key (matches existing convention);
	// GORM associations work on column name regardless. Managed manually on
	// write (see problemRepository.Upsert) and Preloaded on read.
	IOSchema []ProblemIOSchema `gorm:"foreignKey:ProblemID;references:ID"`
}

func (Problem) TableName() string {
	return "problems"
}

func ProblemFromDomain(p *domain.Problem) *Problem {
	tags := make([]Tag, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = Tag{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}
	var difficulty *string
	if p.Difficulty != nil {
		s := string(*p.Difficulty)
		difficulty = &s
	}
	ioSchema := make([]ProblemIOSchema, len(p.IOSchema))
	for i, f := range p.IOSchema {
		ioSchema[i] = ProblemIOSchemaFromDomain(p.ID, f)
	}
	return &Problem{
		ID:                       p.ID,
		Name:                     p.Name,
		Slug:                     p.Slug,
		Difficulty:               difficulty,
		IsPublic:                 p.IsPublic,
		AcceptedCount:            p.AcceptedCount,
		SubmissionCount:          p.SubmissionCount,
		Description:              p.Description,
		OutputFormat:             p.OutputFormat,
		SampleInput:              p.SampleInput,
		SampleOutput:             p.SampleOutput,
		TimeLimit:                p.TimeLimit,
		MemoryLimit:              p.MemoryLimit,
		InputDir:                 p.InputDir,
		CPUTimeLimit:             p.CPUTimeLimit,
		CPUExtraTime:             p.CPUExtraTime,
		WallTimeLimit:            p.WallTimeLimit,
		StackLimit:               p.StackLimit,
		MaxProcessesAndOrThreads: p.MaxProcessesAndOrThreads,
		FloatPrecision:           p.FloatPrecision,
		ExpectedOutputDir:        p.ExpectedOutputDir,
		Hints:                    p.Hints,
		Tags:                     tags,
		IOSchema:                 ioSchema,
	}
}

func (p *Problem) ToDomain() *domain.Problem {
	if p == nil {
		return nil
	}
	tags := make([]domain.Tag, len(p.Tags))
	for i, t := range p.Tags {
		tags[i] = t.ToDomain()
	}
	var difficulty *domain.Difficulty
	if p.Difficulty != nil {
		d := domain.Difficulty(*p.Difficulty)
		difficulty = &d
	}
	ioSchema := make([]domain.ProblemIOField, len(p.IOSchema))
	for i := range p.IOSchema {
		ioSchema[i] = p.IOSchema[i].ToDomain()
	}
	return &domain.Problem{
		ID:                       p.ID,
		Name:                     p.Name,
		Slug:                     p.Slug,
		Difficulty:               difficulty,
		IsPublic:                 p.IsPublic,
		AcceptedCount:            p.AcceptedCount,
		SubmissionCount:          p.SubmissionCount,
		Description:              p.Description,
		OutputFormat:             p.OutputFormat,
		SampleInput:              p.SampleInput,
		SampleOutput:             p.SampleOutput,
		TimeLimit:                p.TimeLimit,
		MemoryLimit:              p.MemoryLimit,
		CreatedAt:                p.CreatedAt,
		UpdatedAt:                p.UpdatedAt,
		InputDir:                 p.InputDir,
		CPUTimeLimit:             p.CPUTimeLimit,
		CPUExtraTime:             p.CPUExtraTime,
		WallTimeLimit:            p.WallTimeLimit,
		StackLimit:               p.StackLimit,
		MaxProcessesAndOrThreads: p.MaxProcessesAndOrThreads,
		FloatPrecision:           p.FloatPrecision,
		ExpectedOutputDir:        p.ExpectedOutputDir,
		Hints:                    p.Hints,
		Tags:                     tags,
		IOSchema:                 ioSchema,
	}
}
