package repository

import (
	"context"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/domain"
	"gorm.io/gorm"
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
