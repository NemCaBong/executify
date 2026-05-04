package http

import "github.com/NemCaBong/executify/internal/adapter/http/handler"

type App struct {
	SubmissionHandler *handler.SubmissionHandler
	ProblemHandler    *handler.ProblemHandler
}

func NewApp(submissionHandler *handler.SubmissionHandler, problemHandler *handler.ProblemHandler) *App {
	return &App{
		SubmissionHandler: submissionHandler,
		ProblemHandler:    problemHandler,
	}
}
