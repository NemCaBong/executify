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
	if err := r.db.WithContext(ctx).Session(&gorm.Session{FullSaveAssociations: true}).Save(dbEntity).Error; err != nil {
		return nil, err
	}
	return dbEntity.ToDomain(), nil
}
