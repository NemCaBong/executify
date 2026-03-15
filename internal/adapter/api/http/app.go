package http

import (
	"net/http"

	"github.com/NemCaBong/executify/internal/core/domain"
	"github.com/NemCaBong/executify/internal/core/port"
	"github.com/gin-gonic/gin"
)

type App struct {
	submissionSvc port.SubmissionService
}

func NewApp(submissionSvc port.SubmissionService) *App {
	return &App{
		submissionSvc: submissionSvc,
	}
}

func (a *App) Submit(c *gin.Context) {
	var submission domain.Submission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := a.submissionSvc.Submit(c.Request.Context(), &submission); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, submission)
}

func (a *App) GetStatus(c *gin.Context) {
	id := c.Param("id")
	submission, err := a.submissionSvc.GetStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, submission)
}
