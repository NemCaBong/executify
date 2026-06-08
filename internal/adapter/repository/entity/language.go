package entity

import "github.com/NemCaBong/executify/internal/domain"

type Language struct {
	ID         int     `gorm:"column:id;primaryKey"`
	Name       string  `gorm:"column:name"`
	Slug       string  `gorm:"column:slug"`
	CompileCmd *string `gorm:"column:compile_cmd"`
	RunCmd     string  `gorm:"column:run_cmd"`
	SourceFile string  `gorm:"column:source_file"`
}

func (Language) TableName() string {
	return "languages"
}

func (m *Language) ToDomain() *domain.Language {
	if m == nil {
		return nil
	}
	return &domain.Language{
		ID:         m.ID,
		Name:       m.Name,
		Slug:       m.Slug,
		CompileCmd: m.CompileCmd,
		RunCmd:     m.RunCmd,
		SourceFile: m.SourceFile,
	}
}
