package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/NemCaBong/executify/internal/application/job"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/redis/go-redis/v9"
)

type runWorker struct {
	cfg          *config.Config
	cache        *redis.Client
	submissionUC *submission.Usecase
}

func NewRunWorker(cfg *config.Config, cache *redis.Client, submissionUC *submission.Usecase) job.Executor {
	return &runWorker{
		cfg:          cfg,
		cache:        cache,
		submissionUC: submissionUC,
	}
}

func (w *runWorker) Execute(ctx context.Context) error {
	var wg sync.WaitGroup
	queueKey := w.cfg.RedisConfig.RunQueueName
	log.Printf("Starting %d run worker goroutines, listening on queue: %s", w.cfg.RunWorkerCount, queueKey)

	for i := 1; i <= w.cfg.RunWorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Printf("Worker %d: Shutting down", workerID)
					return
				default:
				}
				// BLPop blocks waiting for an element. 0 means no timeout.
				result, err := w.cache.BLPop(ctx, 0, queueKey).Result()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Printf("Worker %d: Error popping from redis run queue: %v", workerID, err)
					time.Sleep(1 * time.Second)
					continue
				}

				if len(result) == 2 {
					jobData := result[1]
					log.Printf("Worker %d: Picked up job from run queue: %s", workerID, jobData)
					// TODO: Add actual job unmarshaling and processing logic here
					// actual job processing here need to use new context to not get cancel while running
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}
