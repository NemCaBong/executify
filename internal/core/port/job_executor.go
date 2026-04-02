package port

// JobExecutor is a driven port for executing submissions
type JobExecutor interface {
	Execute() error
}
