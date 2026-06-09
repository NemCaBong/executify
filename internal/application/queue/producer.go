package queue

import "context"

type SubmissionEnqueuer interface {
	EnqueueSubmit(ctx context.Context, submissionID int, enableCommandLog bool) error
	EnqueueRun(ctx context.Context, submissionID int, enableCommandLog bool) error
}
