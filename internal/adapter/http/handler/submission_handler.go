package handler

import (
	"net/http"
	"strconv"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/adapter/http/response"
	"github.com/NemCaBong/executify/internal/adapter/queue"
	appqueue "github.com/NemCaBong/executify/internal/application/queue"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/config"
	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	cfg           *config.Config
	submissionUC  *submission.Usecase
	queueProducer appqueue.Producer
}

func NewSubmissionHandler(cfg *config.Config, submissionUC *submission.Usecase, queueProducer appqueue.Producer) *SubmissionHandler {
	return &SubmissionHandler{
		cfg:           cfg,
		submissionUC:  submissionUC,
		queueProducer: queueProducer,
	}
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	var req request.SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := submission.CreateSubmissionInput{
		LanguageID: req.LanguageID,
		SourceCode: req.SourceCode,
		ProblemID:  req.ProblemID,
	}

	id, err := h.submissionUC.Submit(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	err = h.queueProducer.Enqueue(c.Request.Context(), h.cfg.SubmitQueueName, queue.SubmissionMessage{SubmissionID: id}.ToBytes())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, response.NewSubmitResponse(id))
}

func (h *SubmissionHandler) GetStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	submission, err := h.submissionUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, submission)
}
