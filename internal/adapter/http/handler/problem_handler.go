package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/domain"
)

type ProblemHandler struct {
	problemUC *problem.Usecase
}

func NewProblemHandler(problemUC *problem.Usecase) *ProblemHandler {
	return &ProblemHandler{problemUC: problemUC}
}

func (h *ProblemHandler) Upsert(c *gin.Context) {
	var req request.UpsertProblemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tags := make([]domain.Tag, len(req.Tags))
	for i, t := range req.Tags {
		tags[i] = domain.Tag{ID: t.ID, Name: t.Name, Slug: t.Slug}
	}

	var difficulty *domain.Difficulty
	if req.Difficulty != nil {
		d := domain.Difficulty(*req.Difficulty)
		difficulty = &d
	}
	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	prob := &domain.Problem{
		ID:                       req.ID,
		Name:                     req.Name,
		Slug:                     req.Slug,
		Difficulty:               difficulty,
		IsPublic:                 isPublic,
		Description:              req.Description,
		OutputFormat:             req.OutputFormat,
		SampleInput:              req.SampleInput,
		SampleOutput:             req.SampleOutput,
		TimeLimit:                req.TimeLimit,
		MemoryLimit:              req.MemoryLimit,
		InputFile:                req.InputFile,
		ExpectedOutputFile:       req.ExpectedOutputFile,
		CPUTimeLimit:             req.CPUTimeLimit,
		CPUExtraTime:             req.CPUExtraTime,
		WallTimeLimit:            req.WallTimeLimit,
		StackLimit:               req.StackLimit,
		MaxProcessesAndOrThreads: req.MaxProcessesAndOrThreads,
		TemplateCode:             req.TemplateCode,
		WrapperCode:              req.WrapperCode,
		Hints:                    req.Hints,
		Tags:                     tags,
	}

	result, err := h.problemUC.Upsert(c.Request.Context(), prob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
