package entity

import "time"

type ProblemLanguage struct {
	ProblemID    int       `gorm:"column:problem_id;primaryKey"`
	LanguageID   int       `gorm:"column:language_id;primaryKey"`
	TemplateCode string    `gorm:"column:template_code"`
	WrapperCode  string    `gorm:"column:wrapper_code"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt    time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ProblemLanguage) TableName() string {
	return "problem_languages"
}
