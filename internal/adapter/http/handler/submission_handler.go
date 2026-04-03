package handler

import (
	"net/http"

	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/domain"
	"github.com/gin-gonic/gin"
)

type SubmissionHandler struct {
	submissionUC *submission.Usecase
}

func NewSubmissionHandler(submissionUC *submission.Usecase) *SubmissionHandler {
	return &SubmissionHandler{
		submissionUC: submissionUC,
	}
}

func (h *SubmissionHandler) Submit(c *gin.Context) {
	var submission domain.Submission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.submissionUC.Submit(c.Request.Context(), &submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, submission)
}

func (h *SubmissionHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	submission, err := h.submissionUC.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, submission)
}
