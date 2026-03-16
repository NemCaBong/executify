package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	api_http "github.com/NemCaBong/executify/internal/adapter/api/http"
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
			// Create context that listens for the interrupt signal from the OS.
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			cfg := config.LoadConfig()
			db := config.NewMySQLConnection(cfg)

			repo := repository.NewSubmissionRepository(db)
			exec := executor.NewJobExecutor()

			submissionSvc := service.NewSubmissionService(repo, exec)

			app := api_http.NewApp(submissionSvc)

			r := setupRouter(app)

			srv := &http.Server{
				Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
				Handler: r,
			}

			// Initializing the server in a goroutine so that
			// it won't block the graceful shutdown handling below
			go func() {
				fmt.Printf("Server starting on %s\n", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					fmt.Printf("listen: %v\n", err)
				}
			}()

			// Listen for the interrupt signal.
			<-ctx.Done()

			// Restore default behavior on the interrupt signal and notify user of shutdown.
			stop()
			fmt.Println("shutting down gracefully, press Ctrl+C again to force")

			// The context is used to inform the server it has 5 seconds to finish
			// the request it is currently handling
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			if err := srv.Shutdown(ctx); err != nil {
				fmt.Printf("Server forced to shutdown: %v\n", err)
			}

			fmt.Println("Server exiting")
		},
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func setupRouter(app *api_http.App) *gin.Engine {
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
