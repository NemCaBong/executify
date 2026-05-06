package entity

import "github.com/NemCaBong/executify/internal/domain"

type Tag struct {
	ID   int    `gorm:"column:id;primaryKey"`
	Name string `gorm:"column:name"`
	Slug string `gorm:"column:slug"`
}

func (Tag) TableName() string {
	return "tags"
}

func (t *Tag) ToDomain() domain.Tag {
	return domain.Tag{
		ID:   t.ID,
		Name: t.Name,
		Slug: t.Slug,
	}
}

type ProblemTag struct {
	ProblemID int `gorm:"column:problem_id;primaryKey"`
	TagID     int `gorm:"column:tag_id;primaryKey"`
}

func (ProblemTag) TableName() string {
	return "problem_tags"
}
