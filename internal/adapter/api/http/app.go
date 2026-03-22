package http

type App struct {
	SubmissionHandler *SubmissionHandler
}

func NewApp(submissionHandler *SubmissionHandler) *App {
	return &App{
		SubmissionHandler: submissionHandler,
	}
}
