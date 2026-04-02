package worker

import (
	"github.com/NemCaBong/executify/internal/core/port"
	"github.com/redis/go-redis/v9"
)

type submitWorker struct {
	cache         *redis.Client
	submissionSvc port.SubmissionService
}

func NewSubmitWorker(cache *redis.Client, submissionSvc port.SubmissionService) port.JobExecutor {
	return &submitWorker{
		cache:         cache,
		submissionSvc: submissionSvc,
	}
}

func (w *submitWorker) Execute() error {
	return nil
}
