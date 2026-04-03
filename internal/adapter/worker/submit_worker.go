package worker

import (
	"context"

	"github.com/NemCaBong/executify/internal/application/job"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/redis/go-redis/v9"
)

type submitWorker struct {
	cfg          *config.Config
	cache        *redis.Client
	submissionUC *submission.Usecase
}

func NewSubmitWorker(cfg *config.Config, cache *redis.Client, submissionUC *submission.Usecase) job.Executor {
	return &submitWorker{
		cfg:          cfg,
		cache:        cache,
		submissionUC: submissionUC,
	}
}

func (w *submitWorker) Execute(ctx context.Context) error {
	return nil
}
