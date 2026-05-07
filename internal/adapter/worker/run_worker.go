package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/NemCaBong/executify/internal/adapter/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/application/worker"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/domain"
)

type runWorker struct {
	cfg          *config.Config
	cache        *redis.Client
	submissionUC *submission.Usecase
}

func NewRunWorker(cfg *config.Config, cache *redis.Client, submissionUC *submission.Usecase) worker.Executor {
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
					// Use a fresh context so a shutdown doesn't abort an in-flight execution.
					w.HandleRunSubmission(context.Background(), workerID, jobData)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

func (w *runWorker) HandleRunSubmission(ctx context.Context, workerID int, jobData string) {
	var msg queue.SubmissionMessage
	if err := msg.Unmarshal(jobData); err != nil {
		log.Printf("Worker %d: Error unmarshaling job data: %v", workerID, err)
		return
	}

	submissionDetail, err := w.submissionUC.GetWithDetailsByID(ctx, msg.SubmissionID)
	if err != nil {
		log.Printf("Worker %d: Error getting submission %d: %v", workerID, msg.SubmissionID, err)
		return
	}

	runner := domain.NewCodeRunner(submissionDetail, &submissionDetail.Input)
	if err := runner.Execute(ctx); err != nil {
		log.Printf("Worker %d: Execution error for submission %d: %v", workerID, msg.SubmissionID, err)
		submissionDetail.Status = domain.StatusFailed
		submissionDetail.Submission.Stderr = err.Error()
	}

	now := time.Now()
	submissionDetail.FinishedAt = &now

	if err := w.submissionUC.Update(ctx, &submissionDetail.Submission); err != nil {
		log.Printf("Worker %d: Failed to update submission %d: %v", workerID, msg.SubmissionID, err)
	}
}
