package worker

import (
	"context"

	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/core/port"
	"github.com/redis/go-redis/v9"
)

type submitWorker struct {
	cfg           *config.Config
	cache         *redis.Client
	submissionSvc port.SubmissionService
}

func NewSubmitWorker(cfg *config.Config, cache *redis.Client, submissionSvc port.SubmissionService) port.JobExecutor {
	return &submitWorker{
		cfg:           cfg,
		cache:         cache,
		submissionSvc: submissionSvc,
	}
}

func (w *submitWorker) Execute(ctx context.Context) error {
	return nil
}
