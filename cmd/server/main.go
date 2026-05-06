package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	api_http "github.com/NemCaBong/executify/internal/adapter/http"
	http_handler "github.com/NemCaBong/executify/internal/adapter/http/handler"
	"github.com/NemCaBong/executify/internal/adapter/http/middleware"
	"github.com/NemCaBong/executify/internal/adapter/queue/redis"
	"github.com/NemCaBong/executify/internal/adapter/repository"
	"github.com/NemCaBong/executify/internal/application/problem"
	"github.com/NemCaBong/executify/internal/application/submission"
	"github.com/NemCaBong/executify/internal/application/user"
	"github.com/NemCaBong/executify/internal/config"
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

			cfg := config.Load()
			db := config.NewMySQLConnection(cfg)
			redisClient := config.NewRedisClient(cfg)

			submissionRepo := repository.NewSubmissionRepository(db)
			problemRepo := repository.NewProblemRepository(db)
			userRepo := repository.NewUserRepository(db)
			refreshTokenRepo := repository.NewRefreshTokenRepository(db)
			redisProducer := redis.NewRedisProducer(redisClient)

			submissionUC := submission.NewUsecase(submissionRepo, problemRepo)
			problemUC := problem.NewUsecase(problemRepo)
			userUC := user.NewUsecase(
				userRepo,
				refreshTokenRepo,
				[]byte(cfg.JWTSecret),
				cfg.AccessTokenTTL,
				cfg.RefreshTokenTTL,
			)
			submissionHandler := http_handler.NewSubmissionHandler(&cfg, submissionUC, redisProducer)
			problemHandler := http_handler.NewProblemHandler(problemUC)
			authHandler := http_handler.NewAuthHandler(userUC)
			app := api_http.NewApp(submissionHandler, problemHandler, authHandler)

			r := setupRouter(app, []byte(cfg.JWTSecret))

			srv := &http.Server{
				Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
				Handler: r,
			}

			// Initializing the server in a goroutine so that
			// it won't block the graceful shutdown handling below
			go func() {
				fmt.Printf("Server starting on %s\n", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					fmt.Printf("listen: %v\n", err)
				}
			}()

			// Listen for the interrupt signal.
			<-ctx.Done()

			// Restore default behavior on the interrupt signal and notify user of shutdown.
			stop()
			fmt.Println("shutting down gracefully, press Ctrl+C again to force")

			// The context is used to inform the server it has 30 seconds to finish
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

func setupRouter(app *api_http.App, jwtSecret []byte) *gin.Engine {
	r := gin.Default()
	{
		v1 := r.Group("/api/v1")
		{
			// Auth (public)
			auth := v1.Group("/auth")
			{
				auth.POST("/register", app.AuthHandler.Register)
				auth.POST("/login", app.AuthHandler.Login)
				auth.POST("/refresh", app.AuthHandler.Refresh)
				auth.POST("/logout", app.AuthHandler.Logout)
			}

			// Protected routes
			protected := v1.Group("", middleware.Auth(jwtSecret))
			{
				protected.POST("/submissions", app.SubmissionHandler.Submit)
				protected.POST("/submissions/run", app.SubmissionHandler.Run)
				protected.GET("/submissions/:id", app.SubmissionHandler.GetStatus)
				protected.PUT("/problems", app.ProblemHandler.Upsert)
			}
		}
	}

	return r
}
