package entity

import "github.com/NemCaBong/executify/internal/domain"

type ProblemIOSchema struct {
	ID        int    `gorm:"column:id;primaryKey;autoIncrement"`
	ProblemID int    `gorm:"column:problem_id"`
	Kind      string `gorm:"column:kind"`       // 'input' | 'output'
	LineIndex int    `gorm:"column:line_index"` // 0-based line number within the file
	KeyName   string `gorm:"column:key_name"`
	DataType  string `gorm:"column:data_type"`
}

func (ProblemIOSchema) TableName() string {
	return "problem_io_schema"
}

func (m *ProblemIOSchema) ToDomain() domain.ProblemIOField {
	return domain.ProblemIOField{
		Kind:      domain.IOKind(m.Kind),
		LineIndex: m.LineIndex,
		KeyName:   m.KeyName,
		DataType:  m.DataType,
	}
}

func ProblemIOSchemaFromDomain(problemID int, f domain.ProblemIOField) ProblemIOSchema {
	return ProblemIOSchema{
		ProblemID: problemID,
		Kind:      string(f.Kind),
		LineIndex: f.LineIndex,
		KeyName:   f.KeyName,
		DataType:  f.DataType,
	}
}
