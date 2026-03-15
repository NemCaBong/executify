package repository

import (
	"context"

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
	// GORM's Create method inserts the record into the database
	return r.db.WithContext(ctx).Create(submission).Error
}

func (r *submissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	var submission domain.Submission
	// GORM's First method retrieves the first record that matches the given condition
	err := r.db.WithContext(ctx).First(&submission, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &submission, nil
}

func (r *submissionRepository) Update(ctx context.Context, submission *domain.Submission) error {
	// GORM's Save method updates all fields of the record
	return r.db.WithContext(ctx).Save(submission).Error
}
