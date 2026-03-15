package main

import (
	"fmt"
	"os"

	"github.com/NemCaBong/executify/internal/adapter/api/http"
	"github.com/NemCaBong/executify/internal/adapter/repository"
	"github.com/NemCaBong/executify/internal/adapter/worker/executor"
	"github.com/NemCaBong/executify/internal/config"
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
			cfg := config.LoadConfig()
			db := config.NewMySQLConnection(cfg)

			repo := repository.NewSubmissionRepository(db)
			exec := executor.NewJobExecutor()

			submissionSvc := service.NewSubmissionService(repo, exec)

			app := http.NewApp(submissionSvc)

			r := setupRouter(app)

			fmt.Println("Server starting on :" + cfg.ServerPort)
			if err := r.Run(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil {
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
