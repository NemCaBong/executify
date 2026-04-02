package worker

import (
	"github.com/NemCaBong/executify/internal/core/port"
	"github.com/redis/go-redis/v9"
)

type runWorker struct {
	cache         *redis.Client
	submissionSvc port.SubmissionService
}

func NewRunWorker(cache *redis.Client, submissionSvc port.SubmissionService) port.JobExecutor {
	return &runWorker{
		cache:         cache,
		submissionSvc: submissionSvc,
	}
}

func (w *runWorker) Execute() error {
	return nil
}
