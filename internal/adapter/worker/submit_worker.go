package worker

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/adapter/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/application/worker"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/domain"
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
	var inFlight sync.WaitGroup

	// drainCtx is intentionally derived from Background, NOT from ctx.
	// ctx is the shutdown signal context; cancelling it should stop the BLPop
	// loop from accepting new jobs, but must not immediately abort executions
	// that are already running inside the sandbox. drainCtx stays alive until
	// either all jobs finish or the 30-second hard deadline fires.
	drainCtx, drainCancel := context.WithCancel(context.Background())
	defer drainCancel()

	queueKey := w.cfg.RedisConfig.SubmitQueueName
	zap.L().Info("starting submit workers",
		zap.Int("count", w.cfg.SubmitWorkerCount),
		zap.String("queue", queueKey),
	)

	for i := 1; i <= w.cfg.SubmitWorkerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					// Shutdown signal received: stop polling, let current in-flight jobs
					// drain via drainCtx before the process exits.
					zap.L().Info("submit worker shutting down", zap.Int("worker_id", workerID))
					return
				default:
				}

				// BLPop blocks until a job arrives or ctx is cancelled.
				// When ctx is cancelled it returns an error, which sends us back
				// to the select above and exits the loop on the next iteration.
				result, err := w.cache.BLPop(ctx, 0, queueKey).Result()
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					zap.L().Error("submit worker queue pop failed",
						zap.Int("worker_id", workerID),
						zap.Error(err),
					)
					time.Sleep(1 * time.Second)
					continue
				}

				if len(result) == 2 {
					zap.L().Debug("submit worker picked up job",
						zap.Int("worker_id", workerID),
						zap.String("job_data", result[1]),
					)
					// inFlight.Add must happen here, in the loop goroutine, before
					// wg.Done can fire. This guarantees wg.Wait() below won't return
					// until every dispatched job is already counted in inFlight.
					inFlight.Add(1)
					go func() {
						defer inFlight.Done()
						w.handleJob(drainCtx, workerID, result[1])
					}()
				}
			}
		}(i)
	}

	// All loop goroutines have exited: no new jobs will be dispatched.
	wg.Wait()
	zap.L().Info("submit workers stopped, draining in-flight jobs")

	// Give in-flight sandbox executions up to 30 seconds to finish cleanly.
	// If they exceed the deadline, drainCtx is cancelled, which unblocks
	// runner.Execute and causes handleJob to return with an error.
	//
	// KNOWN LIMITATION: CodeRunner.Execute defers Cleanup(drainCtx). When the
	// hard deadline fires and drainCtx is already cancelled, that Cleanup call
	// runs with a cancelled context and may silently fail, leaving the isolate
	// box unreleased until the next process start reclaims it.
	timer := time.AfterFunc(30*time.Second, drainCancel)
	defer timer.Stop() // cancel the timer early if all jobs finish before the deadline
	inFlight.Wait()
	return nil
}

func (w *submitWorker) handleJob(ctx context.Context, workerID int, jobData string) {
	var msg queue.SubmissionMessage
	if err := msg.Unmarshal(jobData); err != nil {
		zap.L().Error("submit worker failed to unmarshal job",
			zap.Int("worker_id", workerID),
			zap.Error(err),
		)
		return
	}

	boxID := msg.SubmissionID % w.cfg.CodeRunnerConfig.BoxModulus
	l := zap.L().With(
		zap.Int("worker_id", workerID),
		zap.Int("submission_id", msg.SubmissionID),
		zap.Int("box_id", boxID),
	)

	submissionDetail, err := w.submissionUC.GetWithDetailsByID(ctx, msg.SubmissionID)
	if err != nil {
		l.Error("failed to fetch submission", zap.Error(err))
		return
	}

	submissionDetail.Submission.Status = domain.StatusProcessing
	if err = w.submissionUC.Update(ctx, &submissionDetail.Submission); err != nil {
		l.Error("failed to update submission status to processing", zap.Error(err))
		return
	}

	l.Info("executing submission")

	runner := domain.NewCodeRunner(submissionDetail, nil, nil, w.cfg.CodeRunnerConfig)
	if err = runner.Execute(ctx); err != nil {
		l.Error("execution failed", zap.Error(err))
		submissionDetail.Status = domain.StatusFailed
		submissionDetail.Stderr = err.Error()
	}

	now := time.Now()
	submissionDetail.FinishedAt = &now
	l.Info("submission completed", zap.String("verdict", string(submissionDetail.Status)))

	if err = w.submissionUC.Update(ctx, &submissionDetail.Submission); err != nil {
		l.Error("failed to persist submission result", zap.Error(err))
	}
}
