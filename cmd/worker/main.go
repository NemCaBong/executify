package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NemCaBong/executify/internal/adapter/worker/executor"
	"github.com/NemCaBong/executify/internal/core/domain"
)

func main() {
	fmt.Println("Starting Executify Worker...")

	exec := executor.NewJobExecutor()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down worker...")
		cancel()
	}()

	// Simulating a job consumer loop
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	fmt.Println("Worker is ready to process jobs.")

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Worker stopped.")
			return
		case <-ticker.C:
			// simulate fetching a job
			job := &domain.Submission{
				ID:         1,
				SourceCode: "print('hello')",
				LanguageID: 1,
				Status:     domain.StatusPending,
				CreatedAt:  time.Now(),
			}

			fmt.Printf("Fetched job: %d\n", job.ID)
			err := exec.Execute(ctx, job)
			if err != nil {
				fmt.Printf("Error executing job %d: %v\n", job.ID, err)
			}
		}
	}
}
