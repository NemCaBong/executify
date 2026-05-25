package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/NemCaBong/executify/internal/adapter/http/request"
	"github.com/NemCaBong/executify/internal/adapter/http/response"
	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/domain"
	"github.com/NemCaBong/executify/internal/logger"
	"github.com/NemCaBong/executify/pkg/httperr"
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
		httperr.BadRequest(c, err.Error())
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
		Hints:                    req.Hints,
		Tags:                     tags,
	}

	result, err := h.problemUC.Upsert(c.Request.Context(), prob)
	if err != nil {
		httperr.Internal(c)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *ProblemHandler) GetDetails(c *gin.Context) {
	l := logger.FromContext(c.Request.Context())

	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		httperr.BadRequest(c, "problem slug is required")
		return
	}

	languageQuery := c.Query("language")

	details, err := h.problemUC.GetDetails(c.Request.Context(), slug, languageQuery)
	if err != nil {
		switch {
		case errors.Is(err, problem.ErrProblemNotFound):
			httperr.NotFound(c, "problem not found")
		case errors.Is(err, problem.ErrLanguageNotFound):
			httperr.BadRequest(c, "language not found")
		case errors.Is(err, problem.ErrLanguageNotSupported):
			httperr.BadRequest(c, "language not supported for this problem")
		default:
			l.Error("failed to load problem details", zap.String("problem_slug", slug), zap.Error(err))
			httperr.Internal(c)
		}
		return
	}

	c.JSON(http.StatusOK, response.NewProblemDetailsResponse(details))
}
