package handler

import (
	"net/http"

	"github.com/NemCaBong/executify/internal/core/domain"
	"github.com/NemCaBong/executify/internal/core/port"
	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	submissionSvc port.SubmissionService
}

func NewSubmissionHandler(submissionSvc port.SubmissionService) *SubmissionHandler {
	return &SubmissionHandler{
		submissionSvc: submissionSvc,
	}
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	var submission domain.Submission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.submissionSvc.Submit(c.Request.Context(), &submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, submission)
}

func (h *SubmissionHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	submission, err := h.submissionSvc.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, submission)
}
