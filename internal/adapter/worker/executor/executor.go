package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/NemCaBong/executify/internal/core/domain"
	"github.com/NemCaBong/executify/internal/core/port"
)

type jobExecutor struct{}

func NewJobExecutor() port.JobExecutor {
	return &jobExecutor{}
}

func (e *jobExecutor) Execute(ctx context.Context, submission *domain.Submission) error {
	fmt.Printf("Executing submission %s...\n", submission.ID)

	// Simulate execution
	submission.Status = domain.StatusRunning
	time.Sleep(2 * time.Second)

	submission.Status = domain.StatusCompleted
	submission.Stdout = "Success: output matches expected"

	fmt.Printf("Finished submission %s\n", submission.ID)
	return nil
}
