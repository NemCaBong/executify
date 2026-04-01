package repository

import (
	"context"

	"github.com/NemCaBong/executify/internal/adapter/repository/model"
	"github.com/NemCaBong/executify/internal/core/domain"
	"github.com/NemCaBong/executify/internal/core/port"
	"gorm.io/gorm"
)

type submissionRepository struct {
	db *gorm.DB
}

// NewSubmissionRepository creates a new instance of SubmissionRepository
func NewSubmissionRepository(db *gorm.DB) port.SubmissionRepository {
	return &submissionRepository{
		db: db,
	}
}

func (r *submissionRepository) Save(ctx context.Context, submission *domain.Submission) error {
	// Lấy Domain Model nhét vào khuôn DB Model
	dbModel := model.FromDomain(submission)
	
	// GORM's Create method inserts the record into the database
	if err := r.db.WithContext(ctx).Create(dbModel).Error; err != nil {
		return err
	}
	
	// Khi tạo mới db tự sinh ID và CreatedAt, ta phải map ngược gán lại cho Domain Model
	submission.ID = dbModel.ID
	submission.CreatedAt = dbModel.CreatedAt
	return nil
}

func (r *submissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	var dbModel model.Submission
	// GORM's First method retrieves the first record
	err := r.db.WithContext(ctx).First(&dbModel, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	
	// Map ngược từ dữ liệu DB thành Domain gửi ra ngoài
	return dbModel.ToDomain(), nil
}

func (r *submissionRepository) Update(ctx context.Context, submission *domain.Submission) error {
	dbModel := model.FromDomain(submission)
	// GORM's Save method updates all fields of the record
	return r.db.WithContext(ctx).Save(dbModel).Error
}

