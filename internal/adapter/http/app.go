package http

import "github.com/NemCaBong/executify/internal/adapter/http/handler"

type App struct {
	SubmissionHandler *handler.SubmissionHandler
}

func NewApp(submissionHandler *handler.SubmissionHandler) *App {
	return &App{
		SubmissionHandler: submissionHandler,
	}
}
