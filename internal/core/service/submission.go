package service

import (
	"context"

	"github.com/NemCaBong/executify/internal/core/domain"
	"github.com/NemCaBong/executify/internal/core/port"
)

type submissionService struct {
	repo     port.SubmissionRepository
	executor port.JobExecutor
}

func NewSubmissionService(repo port.SubmissionRepository, executor port.JobExecutor) port.SubmissionService {
	return &submissionService{
		repo:     repo,
		executor: executor,
	}
}

func (s *submissionService) Submit(ctx context.Context, submission *domain.Submission) error {
	submission.Status = domain.StatusPending
	if err := s.repo.Save(ctx, submission); err != nil {
		return err
	}

	// In a real scenario, this might be sent to a queue
	// For now, we call the executor (which could be the worker)
	return nil
}

func (s *submissionService) GetStatus(ctx context.Context, id string) (*domain.Submission, error) {
	return s.repo.GetByID(ctx, id)
}
