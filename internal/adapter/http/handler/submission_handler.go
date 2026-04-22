package handler

import (
	"net/http"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/adapter/http/response"
	"github.com/NemCaBong/executify/internal/application/submission"
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

	c.JSON(http.StatusAccepted, response.NewSubmitResponse(id))
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
