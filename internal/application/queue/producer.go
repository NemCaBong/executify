package queue

import "context"

type SubmissionEnqueuer interface {
	EnqueueSubmit(ctx context.Context, submissionID int) error
	EnqueueRun(ctx context.Context, submissionID int) error
}
