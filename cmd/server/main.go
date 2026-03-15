package main

import (
	"fmt"
	"os"

	"github.com/NemCaBong/executify/internal/adapter/api/http"
	"github.com/NemCaBong/executify/internal/adapter/storage/postgres"
	"github.com/NemCaBong/executify/internal/adapter/worker/executor"
	"github.com/NemCaBong/executify/internal/core/service"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "executify",
		Short: "Executify is a powerful backend app for online judge system",
		Long:  `A robust Go service to serve API requests for a scalable online judge system.`,
		Run: func(cmd *cobra.Command, args []string) {
			// 1. Initialize Adapters (Driven)
			repo := postgres.NewSubmissionRepository()
			exec := executor.NewJobExecutor()

			// 2. Initialize Core Service (Use Cases)
			submissionSvc := service.NewSubmissionService(repo, exec)

			// 3. Initialize Adapters (Driving)
			app := http.NewApp(submissionSvc)

			// 4. Setup Gin Router
			r := setupRouter(app)

			fmt.Println("Server starting on :8080")
			if err := r.Run(":8080"); err != nil {
				fmt.Printf("Failed to run server: %v\n", err)
			}
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupRouter(app *http.App) *gin.Engine {
	r := gin.Default()
	{
		v1 := r.Group("/api/v1")
		{
			v1.POST("/submissions", app.Submit)
			v1.GET("/submissions/:id", app.GetStatus)
		}
	}

	return r
}
