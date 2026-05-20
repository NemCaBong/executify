package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/adapter/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/domain"
)

const persistTimeout = 5 * time.Second

type SubmitHandler struct {
	cfg          *config.Config
	submissionUC *submission.Usecase
}

func NewSubmitHandler(cfg *config.Config, submissionUC *submission.Usecase) *SubmitHandler {
	return &SubmitHandler{
		cfg:          cfg,
		submissionUC: submissionUC,
	}
}

func (h *SubmitHandler) Handle(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.UnmarshalSubmissionPayload(task.Payload())
	if err != nil {
		return fmt.Errorf("unmarshal submit payload: %w", err)
	}

	boxID := payload.SubmissionID % h.cfg.CodeRunnerConfig.BoxModulus
	l := zap.L().With(
		zap.Int("submission_id", payload.SubmissionID),
		zap.Int("box_id", boxID),
	)

	submissionDetail, err := h.submissionUC.GetWithDetailsByID(ctx, payload.SubmissionID)
	if err != nil {
		return fmt.Errorf("fetch submission: %w", err)
	}

	submissionDetail.Submission.Status = domain.StatusProcessing
	if err = h.submissionUC.Update(ctx, &submissionDetail.Submission); err != nil {
		return fmt.Errorf("mark submission as processing: %w", err)
	}

	l.Info("executing submission")

	runner := domain.NewCodeRunner(submissionDetail, nil, nil, h.cfg.CodeRunnerConfig).
		WithStatusNotifier(func(notifyCtx context.Context, status domain.SubmissionStatus) error {
			submissionDetail.Submission.Status = status
			return h.submissionUC.Update(notifyCtx, &submissionDetail.Submission)
		})
	if err = runner.Execute(ctx); err != nil {
		return fmt.Errorf("execute submission: %w", err)
	}

	now := time.Now()
	submissionDetail.FinishedAt = &now
	l.Info("submission completed", zap.String("verdict", string(submissionDetail.Status)))

	persistCtx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()
	if err = h.submissionUC.Update(persistCtx, &submissionDetail.Submission); err != nil {
		l.Error("failed to persist submission result", zap.Error(err))
	}
	return nil
}
