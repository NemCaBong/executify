package port

import "context"

type JobExecutor interface {
	Execute(ctx context.Context) error
}
