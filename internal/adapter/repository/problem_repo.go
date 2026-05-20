package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/domain"
)

type problemRepository struct {
	db *gorm.DB
}

func NewProblemRepository(db *gorm.DB) problem.Repository {
	return &problemRepository{db: db}
}

func (r *problemRepository) GetByID(ctx context.Context, id int) (*domain.Problem, error) {
	var dbEntity entity.Problem
	if err := r.db.WithContext(ctx).Preload("Tags").First(&dbEntity, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return dbEntity.ToDomain(), nil
}

func (r *problemRepository) Upsert(ctx context.Context, problem *domain.Problem) (*domain.Problem, error) {
	dbEntity := entity.ProblemFromDomain(problem)

	// Omit Tags from the main save so GORM doesn't try to upsert them as rows.
	if err := r.db.WithContext(ctx).Omit("Tags.*").Save(dbEntity).Error; err != nil {
		return nil, err
	}

	// Replace syncs the join-table rows; dbEntity.Tags is already correct in memory.
	if err := r.db.WithContext(ctx).Model(dbEntity).Association("Tags").Replace(dbEntity.Tags); err != nil {
		return nil, err
	}

	return dbEntity.ToDomain(), nil
}

func (r *problemRepository) GetWrapperCode(ctx context.Context, problemID, languageID int) (string, error) {
	var row entity.ProblemLanguage
	if err := r.db.WithContext(ctx).
		Select("wrapper_code").
		Where("problem_id = ? AND language_id = ?", problemID, languageID).
		First(&row).Error; err != nil {
		return "", err
	}
	return row.WrapperCode, nil
}

func (r *problemRepository) GetProblemLanguageSnippet(ctx context.Context, problemID, languageID int) (*domain.ProblemLanguageSnippet, error) {
	type joined struct {
		TemplateCode string  `gorm:"column:template_code"`
		WrapperCode  string  `gorm:"column:wrapper_code"`
		LanguageID   int     `gorm:"column:language_id"`
		Name         string  `gorm:"column:name"`
		CompileCmd   *string `gorm:"column:compile_cmd"`
		RunCmd       string  `gorm:"column:run_cmd"`
		SourceFile   string  `gorm:"column:source_file"`
	}
	var row joined
	err := r.db.WithContext(ctx).
		Table("problem_languages AS pl").
		Select("pl.template_code, pl.wrapper_code, l.id AS language_id, l.name, l.compile_cmd, l.run_cmd, l.source_file").
		Joins("INNER JOIN languages l ON l.id = pl.language_id").
		Where("pl.problem_id = ? AND pl.language_id = ?", problemID, languageID).
		Take(&row).Error
	if err != nil {
		return nil, err
	}
	return &domain.ProblemLanguageSnippet{
		TemplateCode: row.TemplateCode,
		WrapperCode:  row.WrapperCode,
		Language: domain.Language{
			ID:         row.LanguageID,
			Name:       row.Name,
			CompileCmd: row.CompileCmd,
			RunCmd:     row.RunCmd,
			SourceFile: row.SourceFile,
		},
	}, nil
}

func (r *problemRepository) ListProblemLanguages(ctx context.Context, problemID int) ([]domain.Language, error) {
	var rows []entity.Language
	err := r.db.WithContext(ctx).
		Table("languages AS l").
		Select("l.id, l.name, l.compile_cmd, l.run_cmd, l.source_file").
		Joins("INNER JOIN problem_languages pl ON pl.language_id = l.id").
		Where("pl.problem_id = ?", problemID).
		Order("l.id ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]domain.Language, len(rows))
	for i := range rows {
		out[i] = *rows[i].ToDomain()
	}
	return out, nil
}

func (r *problemRepository) FindLanguageByName(ctx context.Context, query string) (*domain.Language, error) {
	var lang entity.Language
	if err := r.db.WithContext(ctx).
		Where("LOWER(name) LIKE ?", "%"+query+"%").
		Order("id ASC").
		First(&lang).Error; err != nil {
		return nil, err
	}
	return lang.ToDomain(), nil
}
