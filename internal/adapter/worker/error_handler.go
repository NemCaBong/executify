package worker

import (
	"context"
	"errors"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/adapter/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/domain"
)

type ErrorHandler struct {
	submissionUC *submission.Usecase
}

func NewErrorHandler(submissionUC *submission.Usecase) *ErrorHandler {
	return &ErrorHandler{submissionUC: submissionUC}
}

func (h *ErrorHandler) HandleError(ctx context.Context, task *asynq.Task, err error) {
	payload, perr := queue.UnmarshalSubmissionPayload(task.Payload())
	if perr != nil {
		zap.L().Error("error handler: unmarshal payload failed",
			zap.String("task_type", task.Type()),
			zap.Error(perr),
		)
		return
	}

	l := zap.L().With(
		zap.Int("submission_id", payload.SubmissionID),
		zap.String("task_type", task.Type()),
	)

	stderr := err.Error()
	if isQueueLimitError(ctx, err) {
		l.Warn("task aborted by queue timeout/deadline", zap.Error(err))
		stderr = "queue timeout/deadline exceeded: " + err.Error()
	} else {
		l.Error("task failed", zap.Error(err))
	}

	persistCtx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()

	sub, ferr := h.submissionUC.GetByID(persistCtx, payload.SubmissionID)
	if ferr != nil {
		l.Error("error handler: fetch submission failed", zap.Error(ferr))
		return
	}

	now := time.Now()
	sub.Status = domain.StatusInternalError
	sub.Stderr = stderr
	sub.FinishedAt = &now

	if uerr := h.submissionUC.Update(persistCtx, sub); uerr != nil {
		l.Error("error handler: persist failure status failed", zap.Error(uerr))
	}
}

func isQueueLimitError(ctx context.Context, err error) bool {
	if cerr := ctx.Err(); errors.Is(cerr, context.DeadlineExceeded) || errors.Is(cerr, context.Canceled) {
		return true
	}
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}
