package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/adapter/http/response"
	"github.com/NemCaBong/executify/internal/adapter/queue"
	appqueue "github.com/NemCaBong/executify/internal/application/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/NemCaBong/executify/internal/logger"
	"github.com/NemCaBong/executify/pkg/httperr"
)

type SubmissionHandler struct {
	cfg          *config.Config
	submissionUC *submission.Usecase
	enqueuer     appqueue.SubmissionEnqueuer
}

func NewSubmissionHandler(cfg *config.Config, submissionUC *submission.Usecase, enqueuer appqueue.SubmissionEnqueuer) *SubmissionHandler {
	return &SubmissionHandler{
		cfg:          cfg,
		submissionUC: submissionUC,
		enqueuer:     enqueuer,
	}
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	l := logger.FromContext(c.Request.Context())

	var req request.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Warn("invalid submit request", zap.Error(err))
		httperr.BadRequest(c, err.Error())
		return
	}

	input := submission.CreateSubmissionInput{
		LanguageID: req.LanguageID,
		SourceCode: req.SourceCode,
		ProblemID:  req.ProblemID,
	}

	id, err := h.submissionUC.Submit(c.Request.Context(), &input)
	if err != nil {
		l.Error("failed to create submission", zap.Error(err))
		httperr.Internal(c)
		return
	}

	l = l.With(zap.Int("submission_id", id))

	if err = h.enqueuer.EnqueueSubmit(c.Request.Context(), id); err != nil {
		l.Error("failed to enqueue submission", zap.Error(err))
		httperr.Internal(c)
		return
	}

	l.Info("submission enqueued", zap.String("queue", queue.QueueSubmit))
	c.JSON(http.StatusAccepted, response.NewSubmitResponse(id))
}

func (h *SubmissionHandler) Run(c *gin.Context) {
	l := logger.FromContext(c.Request.Context())

	var req request.RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		l.Warn("invalid run request", zap.Error(err))
		httperr.BadRequest(c, err.Error())
		return
	}

	input := submission.CreateRunInput{
		LanguageID:     req.LanguageID,
		ProblemID:      req.ProblemID,
		SourceCode:     req.SourceCode,
		Input:          req.Input,
		ExpectedOutput: req.ExpectedOutput,
	}

	id, err := h.submissionUC.Run(c.Request.Context(), &input)
	if err != nil {
		l.Error("failed to create run submission", zap.Error(err))
		httperr.Internal(c)
		return
	}

	l = l.With(zap.Int("submission_id", id))

	if err = h.enqueuer.EnqueueRun(c.Request.Context(), id); err != nil {
		l.Error("failed to enqueue run submission", zap.Error(err))
		httperr.Internal(c)
		return
	}

	l.Info("run submission enqueued", zap.String("queue", queue.QueueRun))
	c.JSON(http.StatusAccepted, response.NewSubmitResponse(id))
}

func (h *SubmissionHandler) GetStatus(c *gin.Context) {
	l := logger.FromContext(c.Request.Context())

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		l.Warn("invalid submission id param", zap.String("id_param", idStr))
		httperr.BadRequest(c, "submission id must be an integer")
		return
	}

	l = l.With(zap.Int("submission_id", id))

	sub, err := h.submissionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		l.Warn("submission not found", zap.Error(err))
		httperr.NotFound(c, "submission not found")
		return
	}

	l.Info("submission status retrieved", zap.String("verdict", string(sub.Status)))
	c.JSON(http.StatusOK, sub)
}
