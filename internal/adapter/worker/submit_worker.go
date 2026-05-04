package worker

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/NemCaBong/executify/internal/adapter/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/application/worker"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/domain"
	"github.com/redis/go-redis/v9"
)

type submitWorker struct {
	cfg          *config.Config
	cache        *redis.Client
	submissionUC *submission.Usecase
}

func NewSubmitWorker(cfg *config.Config, cache *redis.Client, submissionUC *submission.Usecase) worker.Executor {
	return &submitWorker{
		cfg:          cfg,
		cache:        cache,
		submissionUC: submissionUC,
	}
}

func (w *submitWorker) Execute(ctx context.Context) error {
	var wg sync.WaitGroup
	queueKey := w.cfg.RedisConfig.SubmitQueueName
	log.Printf("Starting %d submit worker goroutines, listening on queue: %s", w.cfg.SubmitWorkerCount, queueKey)

	for i := 1; i <= w.cfg.SubmitWorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Printf("Submit worker %d: Shutting down", workerID)
					return
				default:
				}

				result, err := w.cache.BLPop(ctx, 0, queueKey).Result()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					log.Printf("Submit worker %d: Error popping from redis submit queue: %v", workerID, err)
					time.Sleep(1 * time.Second)
					continue
				}

				if len(result) == 2 {
					jobData := result[1]
					log.Printf("Submit worker %d: Picked up job from submit queue: %s", workerID, jobData)
					w.HandleSubmitSubmission(context.Background(), workerID, jobData)
				}
			}
		}(i)
	}

	wg.Wait()
	return nil
}

func (w *submitWorker) HandleSubmitSubmission(ctx context.Context, workerID int, jobData string) {
	var msg queue.SubmissionMessage
	if err := msg.Unmarshal(jobData); err != nil {
		log.Printf("Submit worker %d: Error unmarshaling job data: %v", workerID, err)
		return
	}

	submissionDetail, err := w.submissionUC.GetWithDetailsByID(ctx, msg.SubmissionID)
	if err != nil {
		log.Printf("Submit worker %d: Error getting submission %d: %v", workerID, msg.SubmissionID, err)
		return
	}

	// submit mode: nil stdin causes CodeRunner to use the problem's InputFile
	runner := domain.NewCodeRunner(submissionDetail, nil)
	if err := runner.Execute(ctx); err != nil {
		log.Printf("Submit worker %d: Execution error for submission %d: %v", workerID, msg.SubmissionID, err)
		submissionDetail.Status = domain.StatusFailed
		submissionDetail.Stderr = err.Error()
	}

	now := time.Now()
	submissionDetail.FinishedAt = &now

	if err := w.submissionUC.Update(ctx, &submissionDetail.Submission); err != nil {
		log.Printf("Submit worker %d: Failed to update submission %d: %v", workerID, msg.SubmissionID, err)
	}
}
