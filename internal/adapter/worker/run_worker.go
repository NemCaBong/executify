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

type RunHandler struct {
	cfg          *config.Config
	submissionUC *submission.Usecase
}

func NewRunHandler(cfg *config.Config, submissionUC *submission.Usecase) *RunHandler {
	return &RunHandler{
		cfg:          cfg,
		submissionUC: submissionUC,
	}
}

func (h *RunHandler) Handle(ctx context.Context, task *asynq.Task) error {
	payload, err := queue.UnmarshalSubmissionPayload(task.Payload())
	if err != nil {
		return fmt.Errorf("unmarshal run payload: %w", err)
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

	l.Info("executing run submission")

	runner := domain.NewCodeRunner(submissionDetail, &submissionDetail.Input, &submissionDetail.ExpectedOutput, h.cfg.CodeRunnerConfig).
		WithStatusNotifier(func(notifyCtx context.Context, status domain.SubmissionStatus) error {
			submissionDetail.Submission.Status = status
			return h.submissionUC.Update(notifyCtx, &submissionDetail.Submission)
		})
	if payload.EnableCommandLog {
		runner.WithCommandLogger(func(stage, command string) {
			l.Info("isolate command", zap.String("stage", stage), zap.String("command", command))
		})
	}
	if err = runner.Execute(ctx); err != nil {
		return fmt.Errorf("execute submission: %w", err)
	}

	now := time.Now()
	submissionDetail.FinishedAt = &now
	l.Info("run submission completed", zap.String("verdict", string(submissionDetail.Status)))

	persistCtx, cancel := context.WithTimeout(context.Background(), persistTimeout)
	defer cancel()
	if err = h.submissionUC.Update(persistCtx, &submissionDetail.Submission); err != nil {
		l.Error("failed to persist run submission result", zap.Error(err))
	}
	return nil
}
