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
	InputFile                string                      `gorm:"column:input_file"`
	CPUTimeLimit             *float64                    `gorm:"column:cpu_time_limit"`
	CPUExtraTime             *float64                    `gorm:"column:cpu_extra_time"`
	WallTimeLimit            *float64                    `gorm:"column:wall_time_limit"`
	StackLimit               *int                        `gorm:"column:stack_limit"`
	MaxProcessesAndOrThreads *int                        `gorm:"column:max_processes_and_or_threads"`
	ExpectedOutputFile       string                      `gorm:"column:expected_output_file"`
	TemplateCode             string                      `gorm:"column:template_code"`
	WrapperCode              string                      `gorm:"column:wrapper_code"`
	Hints                    datatypes.JSONSlice[string] `gorm:"column:hints;type:json"`
	Tags                     []Tag                       `gorm:"many2many:problem_tags;joinForeignKey:problem_id;joinReferences:tag_id"`
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
		InputFile:                p.InputFile,
		CPUTimeLimit:             p.CPUTimeLimit,
		CPUExtraTime:             p.CPUExtraTime,
		WallTimeLimit:            p.WallTimeLimit,
		StackLimit:               p.StackLimit,
		MaxProcessesAndOrThreads: p.MaxProcessesAndOrThreads,
		ExpectedOutputFile:       p.ExpectedOutputFile,
		TemplateCode:             p.TemplateCode,
		WrapperCode:              p.WrapperCode,
		Hints:                    p.Hints,
		Tags:                     tags,
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
		InputFile:                p.InputFile,
		CPUTimeLimit:             p.CPUTimeLimit,
		CPUExtraTime:             p.CPUExtraTime,
		WallTimeLimit:            p.WallTimeLimit,
		StackLimit:               p.StackLimit,
		MaxProcessesAndOrThreads: p.MaxProcessesAndOrThreads,
		ExpectedOutputFile:       p.ExpectedOutputFile,
		TemplateCode:             p.TemplateCode,
		WrapperCode:              p.WrapperCode,
		Hints:                    p.Hints,
		Tags:                     tags,
	}
}
