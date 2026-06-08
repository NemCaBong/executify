package queue

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"

	appqueue "github.com/NemCaBong/executify/internal/application/queue"
)

const (
	enqueueTimeout  = 5 * time.Second
	taskTimeout     = 15 * time.Second
	taskDeadline    = 2 * time.Minute
	defaultMaxRetry = 0
)

type asynqProducer struct {
	client *asynq.Client
}

func NewAsynqProducer(client *asynq.Client) appqueue.SubmissionEnqueuer {
	return &asynqProducer{client: client}
}

func (p *asynqProducer) EnqueueSubmit(ctx context.Context, submissionID int, enableCommandLog bool) error {
	return p.enqueue(ctx, TypeSubmissionSubmit, QueueSubmit, submissionID, enableCommandLog)
}

func (p *asynqProducer) EnqueueRun(ctx context.Context, submissionID int, enableCommandLog bool) error {
	return p.enqueue(ctx, TypeSubmissionRun, QueueRun, submissionID, enableCommandLog)
}

func (p *asynqProducer) enqueue(ctx context.Context, taskType, queueName string, submissionID int, enableCommandLog bool) error {
	payload, err := SubmissionPayload{SubmissionID: submissionID, EnableCommandLog: enableCommandLog}.Marshal()
	if err != nil {
		return fmt.Errorf("marshal submission payload: %w", err)
	}

	task := asynq.NewTask(taskType, payload,
		asynq.Queue(queueName),
		asynq.MaxRetry(defaultMaxRetry),
		asynq.Timeout(taskTimeout),
		asynq.Deadline(time.Now().Add(taskDeadline)),
	)

	enqueueCtx, cancel := context.WithTimeout(ctx, enqueueTimeout)
	defer cancel()

	if _, err := p.client.EnqueueContext(enqueueCtx, task); err != nil {
		return fmt.Errorf("enqueue %s task: %w", taskType, err)
	}
	return nil
}
