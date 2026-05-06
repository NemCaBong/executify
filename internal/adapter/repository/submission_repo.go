package repository

import (
	"context"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/domain"
	"gorm.io/gorm"
)

type submissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) submission.Repository {
	return &submissionRepository{
		db: db,
	}
}

func (r *submissionRepository) Save(ctx context.Context, submission *domain.Submission) error {
	dbEntity := entity.FromDomain(submission)

	if err := r.db.WithContext(ctx).Create(dbEntity).Error; err != nil {
		return err
	}

	// Khi tạo mới db tự sinh ID và CreatedAt, ta phải map ngược gán lại cho Domain entity
	submission.ID = dbEntity.ID
	submission.CreatedAt = dbEntity.CreatedAt
	return nil
}

func (r *submissionRepository) GetByID(ctx context.Context, id int) (*domain.Submission, error) {
	var dbEntity entity.Submission
	err := r.db.WithContext(ctx).First(&dbEntity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	return dbEntity.ToDomain(), nil
}

func (r *submissionRepository) Update(ctx context.Context, submission *domain.Submission) error {
	return r.db.WithContext(ctx).
		Model(&entity.Submission{}).
		Where("id = ?", submission.ID).
		Updates(submission).Error
}

func (r *submissionRepository) GetWithDetailsByID(ctx context.Context, id int) (*domain.SubmissionWithDetails, error) {
	var dbEntity entity.Submission
	err := r.db.WithContext(ctx).Joins("Language").Joins("Problem").Where("submissions.id = ?", id).First(&dbEntity).Error
	if err != nil {
		return nil, err
	}

	return dbEntity.ToDomainWithDetails(), nil
}
