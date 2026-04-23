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

// NewSubmissionRepository creates a new instance of SubmissionRepository
func NewSubmissionRepository(db *gorm.DB) submission.Repository {
	return &submissionRepository{
		db: db,
	}
}

func (r *submissionRepository) Save(ctx context.Context, submission *domain.Submission) error {
	// Lấy Domain entity nhét vào khuôn DB entity
	dbEntity := entity.FromDomain(submission)

	// GORM's Create method inserts the record into the database
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
	// GORM's First method retrieves the first record
	err := r.db.WithContext(ctx).First(&dbEntity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	// Map ngược từ dữ liệu DB thành Domain gửi ra ngoài
	return dbEntity.ToDomain(), nil
}

func (r *submissionRepository) Update(ctx context.Context, submission *domain.Submission) error {
	dbEntity := entity.FromDomain(submission)
	// GORM's Save method updates all fields of the record
	return r.db.WithContext(ctx).Save(dbEntity).Error
}

func (r *submissionRepository) GetWithDetailsByID(ctx context.Context, id int) (*domain.SubmissionWithDetails, error) {
	var dbEntity entity.Submission
	// GORM's First method retrieves the first record
	err := r.db.WithContext(ctx).Joins("Language").Joins("Problem").First(&dbEntity, "id = ?", id).Error
	if err != nil {
		return nil, err
	}

	// Map ngược từ dữ liệu DB thành Domain gửi ra ngoài
	return dbEntity.ToDomainWithDetails(), nil
}
