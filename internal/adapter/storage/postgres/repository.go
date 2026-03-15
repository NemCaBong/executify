package postgres

import (
	"context"
	"errors"
	"sync"

	"github.com/NemCaBong/executify/internal/core/domain"
	"github.com/NemCaBong/executify/internal/core/port"
)

type submissionRepository struct {
	mu          sync.RWMutex
	submissions map[string]*domain.Submission
}

func NewSubmissionRepository() port.SubmissionRepository {
	return &submissionRepository{
		submissions: make(map[string]*domain.Submission),
	}
}

func (r *submissionRepository) Save(ctx context.Context, submission *domain.Submission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.submissions[submission.ID] = submission
	return nil
}

func (r *submissionRepository) GetByID(ctx context.Context, id string) (*domain.Submission, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	submission, ok := r.submissions[id]
	if !ok {
		return nil, errors.New("submission not found")
	}
	return submission, nil
}

func (r *submissionRepository) Update(ctx context.Context, submission *domain.Submission) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.submissions[submission.ID]; !ok {
		return errors.New("submission not found")
	}
	r.submissions[submission.ID] = submission
	return nil
}
