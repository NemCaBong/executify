package http

import "github.com/NemCaBong/executify/internal/adapter/http/handler"

type App struct {
	SubmissionHandler *handler.SubmissionHandler
	ProblemHandler    *handler.ProblemHandler
	AuthHandler       *handler.AuthHandler
}

func NewApp(
	submissionHandler *handler.SubmissionHandler,
	problemHandler *handler.ProblemHandler,
	authHandler *handler.AuthHandler,
) *App {
	return &App{
		SubmissionHandler: submissionHandler,
		ProblemHandler:    problemHandler,
		AuthHandler:       authHandler,
	}
}
