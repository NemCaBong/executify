package queue

import "context"

type SubmissionEnqueuer interface {
	EnqueueSubmit(ctx context.Context, submissionID int, logCommand bool) error
	EnqueueRun(ctx context.Context, submissionID int, logCommand bool) error
}
